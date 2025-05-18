package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "orgm",
	Short: "CLI de ORGM para funciones de la organizacion",
	Long:  `Funcionalidades de la organizacion ORGM para manejo de usuarios, roles, permisos, IA, Adminitracion, Calculos, Proyectos, Fichas, Tecnias, Facturas.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	// Run: DockerCmd,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/orgm/config.toml)")
	rootCmd.PersistentFlags().StringP("author", "a", "osmargm1202", "author name for copyright attribution")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	viper.BindPFlag("author", rootCmd.PersistentFlags().Lookup("author"))
	viper.SetDefault("author", "osmargm1202 <osmargm1202@gmail.com>")
	// Add docker command to root command

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