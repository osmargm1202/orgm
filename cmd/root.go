package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/osmargm1202/orgm/cmd/adm"
	"github.com/osmargm1202/orgm/cmd/misc"
	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version = "v0.132"

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the application",
	Run: func(cmd *cobra.Command, args []string) {
		updateFunc()
	},
}

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the application in dev Linux",
	Run: func(cmd *cobra.Command, args []string) {
		installFunc()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the version of the application",
	Run: func(cmd *cobra.Command, args []string) {
		versionFunc()
	},
}

func versionFunc() {
	fmt.Println(inputs.InfoStyle.Render("Versión: " + version))
}

var RootCmd = &cobra.Command{
	Use:   "orgm",
	Short: "CLI de ORGM para funciones de la organizacion",
	Long:  `Herramientas de la organizacion ORGM.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	// Check for -v or --version before executing the command tree

	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.AddCommand(InstallCmd)
	RootCmd.AddCommand(UpdateCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(adm.AdmCmd)
	RootCmd.AddCommand(misc.MiscCmd)
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	// First check if config.toml exists in current directory
	if _, err := os.Stat("config.toml"); err == nil {
		viper.AddConfigPath(".") // Path: current directory
	} else {
		// If not found in current directory, use home directory config
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Error al obtener el directorio home: %v", err)
		}

		configPath := filepath.Join(homeDir, ".config", "orgm")
		viper.Set("config_path", configPath)
		viper.AddConfigPath(configPath) // Path: ~/.config/orgm
		viper.AddConfigPath(".")        // Path: current directory as fallback
	}

	viper.AutomaticEnv() // Read in environment variables that match

	// Attempt to read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; Viper will rely on env vars or defaults if any.
			// This might be acceptable for some commands.
			fmt.Fprintln(os.Stderr, "Warning: Config file not found. Proceeding without it or using environment variables/defaults.")
		} else {
			// Other error reading config file
			fmt.Fprintf(os.Stderr, "Warning: Error reading config file: %v\n", err)
		}
	} else {
		// Config file found and successfully parsed
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func installFunc() {
	fmt.Printf("%s\n", inputs.TitleStyle.Render("Installing the application in dev Linux"))

	cmd := exec.Command("go", "build", "-o", "orgm", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error during go build: %v\n", err)
		return
	}

	cmd = exec.Command("go", "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error during go install: %v\n", err)
		return
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ Build and install completed successfully"))

}

func updateFunc() {
	fmt.Printf("%s\n", inputs.TitleStyle.Render("Updating the application"))

	// Determine the download URL and installation path based on OS
	var downloadURL, installPath string

	switch runtime.GOOS {
	case "windows":
		downloadURL = "https://github.com/osmargm1202/orgm/releases/latest/download/orgm.exe"
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error getting home directory: %v\n", err)
			return
		}
		installPath = filepath.Join(homeDir, ".config", "orgm", "orgm.exe")
	case "linux":
		downloadURL = "https://github.com/osmargm1202/orgm/releases/latest/download/orgm"
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error getting home directory: %v\n", err)
			return
		}
		installPath = filepath.Join(homeDir, ".local", "bin", "orgm")
	default:
		fmt.Printf("Unsupported operating system: %s\n", runtime.GOOS)
		return
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(installPath), 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	// Download the latest version
	fmt.Printf("%s\n", inputs.InfoStyle.Render("Downloading latest version..."))

	resp, err := http.Get(downloadURL)
	if err != nil {
		fmt.Printf("Error downloading file: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: HTTP %d when downloading %s\n", resp.StatusCode, downloadURL)
		return
	}

	// Create temporary file
	tempFile := installPath + ".tmp"
	out, err := os.Create(tempFile)
	if err != nil {
		fmt.Printf("Error creating temporary file: %v\n", err)
		return
	}
	defer out.Close()

	// Copy downloaded content to temporary file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Remove(tempFile)
		return
	}

	// Close the file before moving it
	out.Close()

	// Remove existing file if it exists and replace with new one
	if _, err := os.Stat(installPath); err == nil {
		if err := os.Remove(installPath); err != nil {
			fmt.Printf("Error removing old file: %v\n", err)
			os.Remove(tempFile)
			return
		}
	}

	// Move temporary file to final location
	if err := os.Rename(tempFile, installPath); err != nil {
		fmt.Printf("Error moving file to final location: %v\n", err)
		os.Remove(tempFile)
		return
	}

	// Set executable permissions on Linux
	if runtime.GOOS == "linux" {
		if err := os.Chmod(installPath, 0755); err != nil {
			fmt.Printf("Error setting executable permissions: %v\n", err)
			return
		}
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ Application updated successfully"))
	fmt.Printf("%s %s\n", inputs.InfoStyle.Render("Updated executable location:"), installPath)
}
