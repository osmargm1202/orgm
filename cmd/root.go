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
	"time"

	"github.com/osmargm1202/orgm/cmd/adm"
	"github.com/osmargm1202/orgm/cmd/misc"
	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version = "v0.134"

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update ORGM to the latest version",
	Long:  `Downloads the latest ORGM installer and updates the binary automatically.`,
	Run: func(cmd *cobra.Command, args []string) {
		updateFunc()
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
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(adm.AdmCmd)
	RootCmd.AddCommand(misc.MiscCmd)
	RootCmd.AddCommand(InitCmd)
	RootCmd.AddCommand(UpdateCmd)
	RootCmd.AddCommand(PropCmd)
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

	// Attempt to read the main configuration file
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
		fmt.Println("Loaded config.toml")
	}

}


func updateFunc() {
	fmt.Printf("%s\n", inputs.TitleStyle.Render("🚀 Updating ORGM to latest version"))

	var installerURL, installerName string

	switch runtime.GOOS {
	case "windows":
		installerURL = "https://raw.githubusercontent.com/osmargm1202/orgm/main/install.bat"
		installerName = "install.bat"
	case "linux", "darwin":
		installerURL = "https://raw.githubusercontent.com/osmargm1202/orgm/main/install.sh"
		installerName = "install.sh"
	default:
		fmt.Printf("❌ Unsupported operating system: %s\n", runtime.GOOS)
		return
	}

	// Download the installer
	fmt.Printf("%s\n", inputs.InfoStyle.Render("📥 Downloading installer..."))

	resp, err := http.Get(installerURL)
	if err != nil {
		fmt.Printf("❌ Error downloading installer: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Error: HTTP %d when downloading installer\n", resp.StatusCode)
		return
	}

	// Create temporary installer file
	tempFile := filepath.Join(os.TempDir(), installerName)
	out, err := os.Create(tempFile)
	if err != nil {
		fmt.Printf("❌ Error creating temporary file: %v\n", err)
		return
	}

	// Copy installer content to temporary file
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		fmt.Printf("❌ Error writing installer: %v\n", err)
		os.Remove(tempFile)
		return
	}

	// Make installer executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tempFile, 0755); err != nil {
			fmt.Printf("❌ Error setting executable permissions: %v\n", err)
			os.Remove(tempFile)
			return
		}
	}

	// Special handling for Windows: cannot replace running executable
	if runtime.GOOS == "windows" {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("⚠️  On Windows, the updater cannot replace the running executable."))
		fmt.Printf("%s\n", inputs.InfoStyle.Render("   The installer will open in a new window. Please CLOSE this terminal or any running orgm.exe before continuing the update."))
		fmt.Printf("%s\n", inputs.InfoStyle.Render("   Press ENTER to continue and launch the installer..."))
		fmt.Scanln()
		// Start installer in a new window and exit this process
		cmd := exec.Command("cmd", "/C", "start", "", tempFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err = cmd.Start()
		// Clean up temporary file after a short delay (let installer copy itself if needed)
		go func(f string) {
			time.Sleep(30 * time.Second)
			os.Remove(f)
		}(tempFile)
		if err != nil {
			fmt.Printf("❌ Error launching installer: %v\n", err)
			return
		}
		fmt.Printf("%s\n", inputs.SuccessStyle.Render("✅ Installer launched. Please follow the instructions in the new window."))
		return
	}

	// For Linux/macOS, run installer directly
	fmt.Printf("%s\n", inputs.InfoStyle.Render("🔧 Running installer..."))

	cmd := exec.Command("bash", tempFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err = cmd.Run()

	// Clean up temporary file
	os.Remove(tempFile)

	if err != nil {
		fmt.Printf("❌ Error running installer: %v\n", err)
		return
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("✅ ORGM updated successfully!"))
	fmt.Printf("%s\n", inputs.InfoStyle.Render("💡 If this is your first time, you may need to open a new terminal or run:"))

	switch runtime.GOOS {
	case "windows":
		fmt.Printf("%s\n", inputs.InfoStyle.Render("   • Open a new Command Prompt or PowerShell"))
	case "linux", "darwin":
		fmt.Printf("%s\n", inputs.InfoStyle.Render("   • source ~/.bashrc (or ~/.zshrc)"))
		fmt.Printf("%s\n", inputs.InfoStyle.Render("   • Or open a new terminal"))
	}
}
