package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
)

// EnsureGCloudIDToken obtains an ID token for Cloud Run.
// Audience is taken from `url.propuestas_api`.
// Credentials file is expected at `<config_path>/orgmdev_google.json`.
func EnsureGCloudIDToken() (string, error) {
    // Try disk cache first
    if cachedTok, cachedExp, ok := loadCachedToken(); ok {
        if time.Unix(cachedExp, 0).After(time.Now().Add(2 * time.Minute)) {
            fmt.Println("💾 Token obtenido utilizando el caché")
            return cachedTok, nil
        }
    }

    audience := viper.GetString("url.propuestas_api")
    if audience == "" {
        return "", fmt.Errorf("url.propuestas_api no está configurado")
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
    fmt.Printf("🔑 Usando credenciales de Google: %s\n", filepath.Base(credFile))

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
    _ = saveCachedToken(tok.AccessToken, expiryUnix)

    return tok.AccessToken, nil
}

// GAuthCmd fetches and stores the Google ID token in the config directory cache file
var GAuthCmd = &cobra.Command{
    Use:   "gauth",
    Short: "Obtener token de autenticación de Google para Cloud Run",
    Run: func(cmd *cobra.Command, args []string) {
        token, err := EnsureGCloudIDToken()
        if err != nil {
            fmt.Println("❌ ", err.Error())
            return
        }
        // Do not print full token; just confirm
        fmt.Println("✅ Token obtenido y almacenado en archivo de caché")
        if _, exp, ok := loadCachedToken(); ok {
            fmt.Printf("🕒 Expira: %s\n", time.Unix(exp, 0).Format(time.RFC3339))
        }
        _ = token
    },
}

func init() {
    RootCmd.AddCommand(GAuthCmd)
}

// tokenCachePath returns the path where the token cache is stored
func tokenCachePath() string {
    configPath := viper.GetString("config_path")
    if configPath == "" {
        home, _ := os.UserHomeDir()
        configPath = filepath.Join(home, ".config", "orgm")
    }
    return filepath.Join(configPath, ".gauth_token.json")
}

func loadCachedToken() (string, int64, bool) {
    path := tokenCachePath()
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

func saveCachedToken(token string, expiryUnix int64) error {
    path := tokenCachePath()
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


