package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/osmargm1202/orgm/inputs"
	"github.com/osmargm1202/orgm/cmd/adm"
	"github.com/osmargm1202/orgm/cmd/misc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version = "v0.132"

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
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func initConfig() {

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.ReadInConfig()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error al obtener el directorio home: %v", err)
		return
	}

	configPath := filepath.Join(homeDir, ".config", "orgm")
	viper.SetDefault("config_path", configPath)

	// Configuración de Viper
	viper.AddConfigPath(configPath) // agregar la ruta al directorio de configuración

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
