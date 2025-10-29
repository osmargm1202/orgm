package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
)

// EnsureGCloudIDTokenForAudience obtains an ID token for a specific audience URL.
// Credentials file is expected at `<config_path>/orgmdev_google.json`.
// quiet parameter suppresses output (useful for CLI mode)
func EnsureGCloudIDTokenForAudience(audience string, quiet bool) (string, error) {
    // Try disk cache first
    if cachedTok, cachedExp, ok := LoadCachedToken(); ok {
        if time.Unix(cachedExp, 0).After(time.Now().Add(2 * time.Minute)) {
            if !quiet {
                fmt.Println("💾 Token obtenido utilizando el caché")
            }
            return cachedTok, nil
        }
    }

    if audience == "" {
        return "", fmt.Errorf("audience URL no está configurado")
    }

    configPath := viper.GetString("config_path")
    if configPath == "" {
        // fallback to default path
        home, _ := os.UserHomeDir()
        configPath = filepath.Join(home, ".config", "orgm")
    }

    // Find any file ending with google.json
    var credFile string
    entries, err := os.ReadDir(configPath)
    if err != nil {
        return "", fmt.Errorf("error leyendo directorio de configuración: %v", err)
    }
    
    for _, entry := range entries {
        if !entry.IsDir() && strings.HasSuffix(entry.Name(), "google.json") {
            credFile = filepath.Join(configPath, entry.Name())
            break
        }
    }
    
    if credFile == "" {
        return "", fmt.Errorf("no se encontró ningún archivo que termine en 'google.json' en %s", configPath)
    }

    // Print which credentials file is being used
    if !quiet {
        fmt.Printf("🔑 Usando credenciales de Google: %s\n", filepath.Base(credFile))
    }

    ctx := context.Background()

    // Prefer idtoken for Cloud Run (OIDC ID token for the service URL)
    ts, err := idtoken.NewTokenSource(ctx, audience, option.WithCredentialsFile(credFile))
    if err != nil {
        // Fallback: try default credentials path if provided via env
        if _, derr := google.FindDefaultCredentials(ctx); derr != nil {
            return "", fmt.Errorf("no se pudo crear TokenSource: %v", err)
        }
        ts, err = idtoken.NewTokenSource(ctx, audience)
        if err != nil {
            return "", fmt.Errorf("no se pudo crear TokenSource (fallback): %v", err)
        }
    }

    tok, err := ts.Token()
    if err != nil {
        return "", fmt.Errorf("no se pudo obtener token: %v", err)
    }

    // Determine expiry
    var expiryUnix int64
    if !tok.Expiry.IsZero() {
        expiryUnix = tok.Expiry.Unix()
    } else {
        // If expiry is missing, set a short-lived default (55 minutes typical for ID tokens)
        expiryUnix = time.Now().Add(55 * time.Minute).Unix()
    }

    // Persist to disk for reuse across runs
    _ = SaveCachedToken(tok.AccessToken, expiryUnix)

    return tok.AccessToken, nil
}

// EnsureGCloudIDToken obtains an ID token for Cloud Run.
// Audience is taken from `url.propuestas_api`.
// Credentials file is expected at `<config_path>/orgmdev_google.json`.
func EnsureGCloudIDToken() (string, error) {
    audience := viper.GetString("url.propuestas_api")
    if audience == "" {
        return "", fmt.Errorf("url.propuestas_api no está configurado")
    }
    return EnsureGCloudIDTokenForAudience(audience, false)
}

// Cloud Run user management functions

const (
	PROJECT_ID = "orgm-797f1"
	DISPLAY_NAME = "Cuenta para Cloud Run"
)

// CreateServiceAccount creates a new Google Cloud service account
func CreateServiceAccount(accountName string) error {
	email := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", accountName, PROJECT_ID)
	
	// Create service account
	cmd := exec.Command("gcloud", "iam", "service-accounts", "create", accountName,
		"--display-name="+DISPLAY_NAME,
		"--project="+PROJECT_ID)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error creando cuenta de servicio: %v\n%s", err, string(output))
	}
	
	fmt.Printf("✅ Cuenta de servicio '%s' creada exitosamente\n", accountName)
	fmt.Printf("📧 Email: %s\n", email)
	
	return nil
}

// CreateCompleteUser creates service account, adds permissions, and downloads credentials
func CreateCompleteUser(accountName string) error {
	email := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", accountName, PROJECT_ID)
	jsonFile := fmt.Sprintf("./%s.json", accountName)
	
	// 1. Create service account
	fmt.Printf("🔧 Creando cuenta de servicio '%s'...\n", accountName)
	err := CreateServiceAccount(accountName)
	if err != nil {
		return fmt.Errorf("error en paso 1 (crear cuenta): %v", err)
	}
	
	// 2. Add Cloud Run permissions
	fmt.Printf("🔐 Agregando permisos de Cloud Run...\n")
	err = AddUserToCloudRun(email)
	if err != nil {
		return fmt.Errorf("error en paso 2 (agregar permisos): %v", err)
	}
	
	// 3. Download JSON credentials
	fmt.Printf("📥 Descargando credenciales JSON...\n")
	err = DownloadServiceAccountJSON(accountName, jsonFile)
	if err != nil {
		return fmt.Errorf("error en paso 3 (descargar credenciales): %v", err)
	}
	
	fmt.Printf("🎉 Usuario '%s' configurado completamente!\n", accountName)
	fmt.Printf("📁 Archivo de credenciales: %s\n", jsonFile)
	
	return nil
}

// DeleteServiceAccount deletes a Google Cloud service account
func DeleteServiceAccount(accountName string) error {
	email := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", accountName, PROJECT_ID)
	
	// Delete service account
	cmd := exec.Command("gcloud", "iam", "service-accounts", "delete", email,
		"--project="+PROJECT_ID,
		"--quiet")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error eliminando cuenta de servicio: %v\n%s", err, string(output))
	}
	
	fmt.Printf("✅ Cuenta de servicio '%s' eliminada exitosamente\n", accountName)
	
	return nil
}

// AddUserToCloudRun adds a user to Cloud Run with appropriate permissions
func AddUserToCloudRun(userEmail string) error {
	// Add Cloud Run Developer role
	cmd := exec.Command("gcloud", "projects", "add-iam-policy-binding", PROJECT_ID,
		"--member=serviceAccount:"+userEmail,
		"--role=roles/run.developer")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error agregando usuario a Cloud Run: %v\n%s", err, string(output))
	}
	
	fmt.Printf("✅ Usuario '%s' agregado a Cloud Run exitosamente\n", userEmail)
	
	return nil
}

// RemoveUserFromCloudRun removes a user from Cloud Run permissions
func RemoveUserFromCloudRun(userEmail string) error {
	// Remove Cloud Run Developer role
	cmd := exec.Command("gcloud", "projects", "remove-iam-policy-binding", PROJECT_ID,
		"--member=user:"+userEmail,
		"--role=roles/run.developer")
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error removiendo usuario de Cloud Run: %v\n%s", err, string(output))
	}
	
	fmt.Printf("✅ Usuario '%s' removido de Cloud Run exitosamente\n", userEmail)
	
	return nil
}

// DownloadServiceAccountJSON downloads the JSON credentials for a service account
func DownloadServiceAccountJSON(accountName, outputPath string) error {
	email := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", accountName, PROJECT_ID)
	
	if outputPath == "" {
		outputPath = fmt.Sprintf("./%s.json", accountName)
	}
	
	// Create JSON key
	cmd := exec.Command("gcloud", "iam", "service-accounts", "keys", "create", outputPath,
		"--iam-account="+email,
		"--project="+PROJECT_ID)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error descargando credenciales JSON: %v\n%s", err, string(output))
	}
	
	fmt.Printf("✅ Credenciales JSON descargadas: %s\n", outputPath)
	
	return nil
}

// GAuthCmd is the main command for Google Cloud authentication (token)
var GAuthCmd = &cobra.Command{
    Use:   "gauth",
    Short: "Obtener token de autenticación de Google para Cloud Run",
    Long:  "Obtiene y almacena un token de autenticación de Google para Cloud Run",
    Run: func(cmd *cobra.Command, args []string) {
        printToken, _ := cmd.Flags().GetBool("print-token")
        
        // Get audience URL
        audience := viper.GetString("url.propuestas_api")
        if audience == "" {
            if printToken {
                fmt.Fprintf(os.Stderr, "Error: url.propuestas_api no está configurado\n")
                os.Exit(1)
            } else {
                fmt.Println("❌ url.propuestas_api no está configurado")
            }
            return
        }
        
        token, err := EnsureGCloudIDTokenForAudience(audience, printToken)
        if err != nil {
            if printToken {
                fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
                os.Exit(1)
            } else {
                fmt.Println("❌ ", err.Error())
            }
            return
        }
        
        // If print-token flag, just output the token
        if printToken {
            fmt.Print(token)
            return
        }
        
        // Do not print full token; just confirm
        fmt.Println("✅ Token obtenido y almacenado en archivo de caché")
        if _, exp, ok := LoadCachedToken(); ok {
            fmt.Printf("🕒 Expira: %s\n", time.Unix(exp, 0).Format(time.RFC3339))
        }
        _ = token
    },
}

// Subcommands for GAuthCmd

// CreateCmd creates a new service account, adds permissions, and downloads credentials
var CreateCmd = &cobra.Command{
    Use:   "create [usuario]",
    Short: "Crear usuario completo (cuenta + permisos + credenciales)",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        accountName := args[0]
        
        err := CreateCompleteUser(accountName)
        if err != nil {
            fmt.Println("❌ ", err.Error())
            return
        }
    },
}

// AddCmd adds permissions to an existing service account
var AddCmd = &cobra.Command{
    Use:   "add [usuario]",
    Short: "Agregar permisos a una cuenta de servicio existente",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        accountName := args[0]
        email := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", accountName, PROJECT_ID)
        
        err := AddUserToCloudRun(email)
        if err != nil {
            fmt.Println("❌ ", err.Error())
            return
        }
    },
}

// DownloadCmd downloads JSON credentials for a service account
var DownloadCmd = &cobra.Command{
    Use:   "download [usuario]",
    Short: "Descargar credenciales JSON de una cuenta de servicio",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        accountName := args[0]
        outputPath, _ := cmd.Flags().GetString("output")
        
        err := DownloadServiceAccountJSON(accountName, outputPath)
        if err != nil {
            fmt.Println("❌ ", err.Error())
            return
        }
    },
}

// DeleteCmd deletes a service account
var DeleteCmd = &cobra.Command{
    Use:   "delete [usuario]",
    Short: "Eliminar cuenta de servicio",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        accountName := args[0]
        
        err := DeleteServiceAccount(accountName)
        if err != nil {
            fmt.Println("❌ ", err.Error())
            return
        }
    },
}

func init() {
    RootCmd.AddCommand(GAuthCmd)
    
    // Add subcommands
    GAuthCmd.AddCommand(CreateCmd)
    GAuthCmd.AddCommand(AddCmd)
    GAuthCmd.AddCommand(DownloadCmd)
    GAuthCmd.AddCommand(DeleteCmd)
    
    // Add flags
    GAuthCmd.Flags().BoolP("print-token", "p", false, "Print only the token (for use in scripts)")
    DownloadCmd.Flags().StringP("output", "o", "", "Ruta de salida para el archivo JSON (por defecto: ./[usuario].json)")
}

// TokenCachePath returns the path where the token cache is stored
func TokenCachePath() string {
    configPath := viper.GetString("config_path")
    if configPath == "" {
        home, _ := os.UserHomeDir()
        configPath = filepath.Join(home, ".config", "orgm")
    }
    return filepath.Join(configPath, ".gauth_token.json")
}

// LoadCachedToken loads a cached token from disk
func LoadCachedToken() (string, int64, bool) {
    path := TokenCachePath()
    b, err := os.ReadFile(path)
    if err != nil {
        return "", 0, false
    }
    var data struct {
        Token       string `json:"token"`
        ExpiryUnix  int64  `json:"expiry_unix"`
    }
    if err := json.Unmarshal(b, &data); err != nil {
        return "", 0, false
    }
    if data.Token == "" || data.ExpiryUnix == 0 {
        return "", 0, false
    }
    return data.Token, data.ExpiryUnix, true
}

// SaveCachedToken saves a token to disk cache
func SaveCachedToken(token string, expiryUnix int64) error {
    path := TokenCachePath()
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return err
    }
    data := struct {
        Token      string `json:"token"`
        ExpiryUnix int64  `json:"expiry_unix"`
    }{Token: token, ExpiryUnix: expiryUnix}
    b, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(path, b, 0600)
}


