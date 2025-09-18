package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/studio-b12/gowebdav"
	tea "github.com/charmbracelet/bubbletea"
)

func InitializePostgrest() (string, map[string]string) {

	// Get PostgREST URL from configs
	postgrestURL := viper.GetString("url.postgrest")
	if postgrestURL == "" {
		log.Fatal("Error: url.postgrest is not defined in config file")
		return "", nil
	}

	// Initialize headers
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Prefer"] = "return=representation"
	headers["accept"] = "application/json"
	headers["CF-Access-Client-Id"] = viper.GetString("cloudflare.CF_ACCESS_CLIENT_ID")
	headers["CF-Access-Client-Secret"] = viper.GetString("cloudflare.CF_ACCESS_CLIENT_SECRET")

	return postgrestURL, headers
}

func InitializeApi() (string, map[string]string) {

	// Get API URL from config
	apiURL := viper.GetString("url.apis")
	if apiURL == "" {
		log.Fatal("Error: url.apis is not defined in config file")
		return "", nil
	}

	// Initialize headers
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["accept"] = "application/json"
	headers["CF-Access-Client-Id"] = viper.GetString("cloudflare.CF_ACCESS_CLIENT_ID")
	headers["CF-Access-Client-Secret"] = viper.GetString("cloudflare.CF_ACCESS_CLIENT_SECRET")

	return apiURL, headers
}

func InitializeNextcloud() *gowebdav.Client {

	// Get Nextcloud URL from config
	nextcloudURL := viper.GetString("nextcloud.url")
	if nextcloudURL == "" {
		log.Fatal("Error: nextcloud.url is not defined in config file")
		return nil
	}
	username := viper.GetString("nextcloud.username")
	password := viper.GetString("nextcloud.password")

	nextcloudURL = nextcloudURL + "/remote.php/dav/files/" + username

	client := gowebdav.NewClient(nextcloudURL, username, password)

	// if err := client.Connect(); err != nil {
	// 	log.Fatal("Error connecting to Nextcloud:", err)
	// 	return nil
	// }

	return client
}

// getAuthToken retrieves a token from the secret worker endpoint
func getAuthToken(key string) (string, error) {
	url := fmt.Sprintf("https://secret.or-gm.com/?key=%s", key)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}

	token := string(body)
	if token == "" {
		return "", fmt.Errorf("received empty token")
	}

	return token, nil
}

// downloadConfigFile downloads a config file using the provided token
func downloadConfigFile(token, filename string) error {
	ctx := context.Background()
	baseURL := "https://r2.or-gm.com/orgm-privado"
	url := fmt.Sprintf("%s/%s", baseURL, filename)

	data, err := r2HTTPGet(ctx, url, token)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", filename, err)
	}

	// Save to config directory
	configDir, err := resolveConfigDir()
	if err != nil {
		return fmt.Errorf("failed to resolve config directory: %w", err)
	}

	filePath := fmt.Sprintf("%s/%s", configDir, filename)
	if err := SaveBytes(filePath, data); err != nil {
		return fmt.Errorf("failed to save %s: %w", filename, err)
	}

	fmt.Printf("%s downloaded successfully\n", filename)
	return nil
}

// InitCmd represents the init command
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ORGM CLI with authentication",
	Long:  `Initialize the ORGM CLI by providing an authentication key to download configuration files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Prompt for authentication key
		model := inputs.TextInput("Enter your authentication key:", "")
		p := tea.NewProgram(model, tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			return fmt.Errorf("error running input program: %w", err)
		}

		var key string
		if m, ok := finalModel.(inputs.TextInputModel); ok {
			key = m.TextInput.Value()
		} else {
			return fmt.Errorf("could not cast final model to TextInputModel")
		}

		if key == "" {
			return fmt.Errorf("authentication key cannot be empty")
		}

		// Get token from secret worker
		fmt.Println("Retrieving authentication token...")
		token, err := getAuthToken(key)
		if err != nil {
			return fmt.Errorf("failed to get authentication token: %w", err)
		}

		fmt.Println("Token retrieved successfully")

		// Download all config files
		configFiles := []string{"keys.toml", "links.toml", "config.toml"}

		for _, filename := range configFiles {
			fmt.Printf("Downloading %s...\n", filename)
			if err := downloadConfigFile(token, filename); err != nil {
				return err
			}
		}

		// Reload viper configuration to include the new keys.toml
		if err := viper.ReadInConfig(); err != nil {
			fmt.Printf("Warning: Could not reload main config: %v\n", err)
		}

		// Load additional configs (this will include the newly downloaded keys.toml)
		loadAdditionalConfigs()

		fmt.Printf("%s\n", inputs.SuccessStyle.Render("ORGM CLI initialized successfully!"))
		fmt.Println("You can now use 'orgm keys update', 'orgm links update', and 'orgm config update' to upload files.")

		return nil
	},
}
