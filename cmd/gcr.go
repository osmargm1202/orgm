package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/osmargm1202/orgm/pkg/cliconfig"
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
// Audience is taken from API worker using key `propuestas_api`.
// Credentials file is expected at `<config_path>/orgmdev_google.json`.
// DEPRECATED: Use EnsureGCloudIDTokenForAPI instead for better control
func EnsureGCloudIDToken() (string, error) {
	return EnsureGCloudIDTokenForAPI("propuestas_api")
}

// EnsureGCloudIDTokenForAPI obtains an ID token for Cloud Run for a specific API.
// apiKey: the worker key to fetch the API URL (e.g., "propuestas_api", "api_calc_management", "api_admapp")
// Credentials file is expected at `<config_path>/orgmdev_google.json`.
func EnsureGCloudIDTokenForAPI(apiKey string) (string, error) {
	// timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	// log.Printf("[DEBUG %s] Obteniendo URL de API desde worker para llave: %s", timestamp, apiKey)

	// Test function to validate URL (simple HTTP GET request with timeout)
	testURL := func(url string) error {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		resp.Body.Close()
		// Accept any 2xx or 3xx status code as valid
		if resp.StatusCode >= 400 {
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		return nil
	}

	// Try cached config first
	audience, err := cliconfig.GetCachedConfig(apiKey, testURL)
	if err != nil {
		// log.Printf("[DEBUG %s] Error obteniendo URL desde caché/worker: %v, intentando con viper como fallback", timestamp, err)
		// Fallback to viper for backwards compatibility (only for propuestas_api)
		if apiKey == "propuestas_api" {
			audience = viper.GetString("url.propuestas_api")
		}
	}

	if audience == "" {
		return "", fmt.Errorf("url para %s no está configurado (ni en API worker ni en config)", apiKey)
	}

	// log.Printf("[DEBUG %s] URL obtenida: %s", timestamp, audience)
	return EnsureGCloudIDTokenForAudience(audience, true)
}

// Cloud Run user management functions

const (
	PROJECT_ID = "orgm-797f1"
	DISPLAY_NAME = "Cuenta para Cloud Run"
)

// debugLog prints a debug message with timestamp
func debugLog(format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)
	fmt.Printf("[DEBUG %s] %s\n", timestamp, message)
}

// debugLogCmd prints the command that will be executed
func debugLogCmd(cmd *exec.Cmd) {
	cmdStr := cmd.Path
	for _, arg := range cmd.Args[1:] {
		cmdStr += " " + arg
	}
	debugLog("Ejecutando comando: %s", cmdStr)
}

// CreateServiceAccount creates a new Google Cloud service account
func CreateServiceAccount(accountName string) error {
	startTime := time.Now()
	email := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", accountName, PROJECT_ID)
	
	debugLog("Iniciando creación de cuenta de servicio: %s", accountName)
	debugLog("Email de la cuenta: %s", email)
	
	// Create service account
	cmd := exec.Command("gcloud", "iam", "service-accounts", "create", accountName,
		"--display-name="+DISPLAY_NAME,
		"--project="+PROJECT_ID)
	
	debugLogCmd(cmd)
	debugLog("Inicio de ejecución del comando: %s", startTime.Format("2006-01-02 15:04:05.000"))
	
	output, err := cmd.CombinedOutput()
	
	elapsed := time.Since(startTime)
	debugLog("Comando finalizado. Tiempo transcurrido: %v", elapsed)
	
	if err != nil {
		debugLog("Error en la ejecución del comando: %v", err)
		debugLog("Output del comando: %s", string(output))
		return fmt.Errorf("error creando cuenta de servicio: %v\n%s", err, string(output))
	}
	
	debugLog("Comando ejecutado exitosamente")
	fmt.Printf("✅ Cuenta de servicio '%s' creada exitosamente\n", accountName)
	fmt.Printf("📧 Email: %s\n", email)
	debugLog("Tiempo total de creación: %v", elapsed)
	
	return nil
}

// CreateCompleteUser creates service account, adds permissions, and downloads credentials
func CreateCompleteUser(accountName string) error {
	overallStartTime := time.Now()
	email := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", accountName, PROJECT_ID)
	jsonFile := fmt.Sprintf("./%s.json", accountName)
	
	debugLog("========================================")
	debugLog("Iniciando proceso completo de creación de usuario: %s", accountName)
	debugLog("Email: %s", email)
	debugLog("Archivo JSON de salida: %s", jsonFile)
	debugLog("Hora de inicio: %s", overallStartTime.Format("2006-01-02 15:04:05.000"))
	debugLog("========================================")
	
	// 1. Create service account
	fmt.Printf("🔧 Creando cuenta de servicio '%s'...\n", accountName)
	debugLog("--- PASO 1: Crear cuenta de servicio ---")
	step1Start := time.Now()
	err := CreateServiceAccount(accountName)
	step1Elapsed := time.Since(step1Start)
	if err != nil {
		debugLog("PASO 1 falló después de %v", step1Elapsed)
		return fmt.Errorf("error en paso 1 (crear cuenta): %v", err)
	}
	debugLog("PASO 1 completado en %v", step1Elapsed)
	debugLog("--- Fin PASO 1 ---")
	
	// 2. Add Cloud Run permissions
	fmt.Printf("🔐 Agregando permisos de Cloud Run...\n")
	debugLog("--- PASO 2: Agregar permisos de Cloud Run ---")
	step2Start := time.Now()
	err = AddUserToCloudRun(email)
	step2Elapsed := time.Since(step2Start)
	if err != nil {
		debugLog("PASO 2 falló después de %v", step2Elapsed)
		return fmt.Errorf("error en paso 2 (agregar permisos): %v", err)
	}
	debugLog("PASO 2 completado en %v", step2Elapsed)
	debugLog("--- Fin PASO 2 ---")
	
	// 3. Download JSON credentials
	fmt.Printf("📥 Descargando credenciales JSON...\n")
	debugLog("--- PASO 3: Descargar credenciales JSON ---")
	step3Start := time.Now()
	err = DownloadServiceAccountJSON(accountName, jsonFile)
	step3Elapsed := time.Since(step3Start)
	if err != nil {
		debugLog("PASO 3 falló después de %v", step3Elapsed)
		return fmt.Errorf("error en paso 3 (descargar credenciales): %v", err)
	}
	debugLog("PASO 3 completado en %v", step3Elapsed)
	debugLog("--- Fin PASO 3 ---")
	
	overallElapsed := time.Since(overallStartTime)
	debugLog("========================================")
	debugLog("Proceso completo finalizado exitosamente")
	debugLog("Tiempo total: %v", overallElapsed)
	debugLog("Resumen de tiempos:")
	debugLog("  - Paso 1 (Crear cuenta): %v", step1Elapsed)
	debugLog("  - Paso 2 (Agregar permisos): %v", step2Elapsed)
	debugLog("  - Paso 3 (Descargar JSON): %v", step3Elapsed)
	debugLog("========================================")
	
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
	startTime := time.Now()
	
	debugLog("Iniciando agregado de permisos de Cloud Run para: %s", userEmail)
	debugLog("Proyecto: %s", PROJECT_ID)
	debugLog("Rol: roles/run.developer")
	debugLog("NOTA: Esta operación puede tardar mucho porque modifica la política IAM del proyecto")
	
	// Add Cloud Run Developer role
	cmd := exec.Command("gcloud", "projects", "add-iam-policy-binding", PROJECT_ID,
		"--member=serviceAccount:"+userEmail,
		"--role=roles/run.developer")
	
	debugLogCmd(cmd)
	debugLog("Inicio de ejecución del comando: %s", startTime.Format("2006-01-02 15:04:05.000"))
	
	output, err := cmd.CombinedOutput()
	
	elapsed := time.Since(startTime)
	debugLog("Comando finalizado. Tiempo transcurrido: %v", elapsed)
	
	if err != nil {
		debugLog("Error en la ejecución del comando: %v", err)
		debugLog("Output del comando: %s", string(output))
		return fmt.Errorf("error agregando usuario a Cloud Run: %v\n%s", err, string(output))
	}
	
	debugLog("Comando ejecutado exitosamente")
	fmt.Printf("✅ Usuario '%s' agregado a Cloud Run exitosamente\n", userEmail)
	debugLog("Tiempo total de agregado de permisos: %v", elapsed)
	
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
	startTime := time.Now()
	email := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", accountName, PROJECT_ID)
	
	debugLog("Iniciando descarga de credenciales JSON para: %s", accountName)
	debugLog("Email de la cuenta: %s", email)
	
	if outputPath == "" {
		outputPath = fmt.Sprintf("./%s.json", accountName)
	}
	
	debugLog("Ruta de archivo de salida: %s", outputPath)
	
	// Create JSON key
	cmd := exec.Command("gcloud", "iam", "service-accounts", "keys", "create", outputPath,
		"--iam-account="+email,
		"--project="+PROJECT_ID)
	
	debugLogCmd(cmd)
	debugLog("Inicio de ejecución del comando: %s", startTime.Format("2006-01-02 15:04:05.000"))
	
	output, err := cmd.CombinedOutput()
	
	elapsed := time.Since(startTime)
	debugLog("Comando finalizado. Tiempo transcurrido: %v", elapsed)
	
	if err != nil {
		debugLog("Error en la ejecución del comando: %v", err)
		debugLog("Output del comando: %s", string(output))
		return fmt.Errorf("error descargando credenciales JSON: %v\n%s", err, string(output))
	}
	
	debugLog("Comando ejecutado exitosamente")
	fmt.Printf("✅ Credenciales JSON descargadas: %s\n", outputPath)
	debugLog("Tiempo total de descarga: %v", elapsed)
	
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


