/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/studio-b12/gowebdav"
)

func InitializePostgrest() (string, map[string]string) {

	// Get PostgREST URL from config
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
		log.Fatal("Error: url.api is not defined in config file")
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

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Edit Config commands with nano or editor choice",
	Long:  `Edit Config commands with nano or editor choice for example: orgm config nano or orgm config nvim`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			args = []string{"nano"}
		}
		EditConfig(args[0])
	},
}

func InitializeNextcloud() *gowebdav.Client {

	// Get Nextcloud URL from config
	nextcloudURL := viper.GetString("url.nextcloud")
	if nextcloudURL == "" {
		log.Fatal("Error: url.nextcloud is not defined in config file")
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

func EditConfig(editor string) {

	configPath := viper.GetViper().ConfigFileUsed()

	var cmd *exec.Cmd
	if editor == "" {
		editor = "nano"
	}
	cmd = exec.Command(editor, configPath)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatal("Error running editor:", err)
		return
	}

}

func CopyConfig() {
	// Get config path from viper
	configPath := viper.GetString("config_path")
	if configPath == "" {
		log.Fatal("Error: config_path is not defined")
		return
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configPath, 0755); err != nil {
		log.Fatalf("Error creating config directory: %v", err)
		return
	}

	// Download and copy config files from GitHub
	configFiles := []string{
		"config.toml",
		"config.example.toml",
	}

	baseURL := "https://raw.githubusercontent.com/osmargm1202/orgm/main/configs/"

	for _, file := range configFiles {
		// Download file from GitHub
		resp, err := http.Get(baseURL + file)
		if err != nil {
			log.Printf("Error downloading %s: %v", file, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Error downloading %s: status code %d", file, resp.StatusCode)
			continue
		}

		// Create destination file
		destPath := filepath.Join(configPath, file)
		destFile, err := os.Create(destPath)
		if err != nil {
			log.Printf("Error creating %s: %v", destPath, err)
			continue
		}
		defer destFile.Close()

		// Copy content
		if _, err := io.Copy(destFile, resp.Body); err != nil {
			log.Printf("Error copying %s: %v", file, err)
			continue
		}

		fmt.Printf("Successfully copied %s to %s\n", file, destPath)
	}
}


func init() {
	rootCmd.AddCommand(configCmd)

}

