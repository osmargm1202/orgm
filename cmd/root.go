package cmd

import (
	"fmt"
	"os"

	"github.com/osmargm1202/orgm/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version = "1"

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "orgm",
	Short: "CLI de ORGM para funciones de la organizacion",
	Long:  `Herramientas de la organizacion ORGM.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Add version command
	RootCmd.AddCommand(versionCmd)

	// Add config management command
	RootCmd.AddCommand(configCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the version of the application",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("orgm version %s\n", version)
	},
}

func initConfig() {
	// Initialize viper configuration
	paths, err := config.GetPaths()
	if err == nil {
		viper.SetConfigFile(paths.GlobalConfig)
		viper.Set("config_path", paths.ConfigDir)
	} else {
		// Fallback to current directory
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv()

	// Attempt to read the main configuration file
	// Config file not found is OK - we'll create it as needed
	_ = viper.ReadInConfig()
}

// GetConfigPath returns the path to the CLI config directory
func GetConfigPath() (string, error) {
	paths, err := config.GetPaths()
	if err != nil {
		return "", err
	}
	return paths.ConfigDir, nil
}
