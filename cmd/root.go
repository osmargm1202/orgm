package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/osmargm1202/orgm/cmd/adm"
	"github.com/osmargm1202/orgm/cmd/apps"
	"github.com/osmargm1202/orgm/cmd/misc"
	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version = "v0.134"

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
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(adm.AdmCmd)
	RootCmd.AddCommand(apps.AppsCmd)
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
