package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/osmargm1202/orgm/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage orgm configuration",
	Long:  `Manage orgm configuration files and paths.`,
}

func init() {
	// Add subcommands
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configAppPathCmd)
	configCmd.AddCommand(configGetCmd)
}

// configInitCmd initializes the configuration directory and files
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize orgm configuration directories and files",
	Long:  `Creates the ~/.config/orgm/ directory structure and initial config files if they don't exist.`,
	Run: func(cmd *cobra.Command, args []string) {
		paths, err := config.GetPaths()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Ensure global config exists
		if err := config.EnsureGlobalConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Configuration initialized successfully\n")
		fmt.Printf("  Config directory: %s\n", paths.ConfigDir)
		fmt.Printf("  Global config:    %s\n", paths.GlobalConfig)
	},
}

// configPathCmd shows configuration paths
var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show orgm configuration paths",
	Long:  `Displays the paths used by orgm for configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		config.ShowPaths()
	},
}

// configAppPathCmd shows app configuration paths
var configAppPathCmd = &cobra.Command{
	Use:   "app-path [app-name]",
	Short: "Show configuration path for a specific app",
	Long:  `Displays the config path for a specific app (e.g., 'orgm config app-path org').`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config.ShowAppPath(args[0])
	},
}

// configShowCmd shows current configuration
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Displays the current configuration values loaded from config files.`,
	Run: func(cmd *cobra.Command, args []string) {
		showConfig()
	},
}

// configGetCmd gets a specific configuration value
var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a specific configuration value",
	Long:  `Get a specific configuration value by key (e.g., 'orgm config get version').`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		getConfigValue(args[0])
	},
}

func showConfig() {
	fmt.Println("ORGM Configuration")
	fmt.Println()

	// Show paths
	fmt.Println("Configuration Paths:")
	config.ShowPaths()
	fmt.Println()

	// Show if global config exists
	paths, err := config.GetPaths()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	if _, err := os.Stat(paths.GlobalConfig); err == nil {
		fmt.Printf("Global config exists: yes\n")
	} else {
		fmt.Printf("Global config exists: no (run 'orgm config init' to create)\n")
	}

	fmt.Println()

	// Show loaded configuration values
	allKeys := viper.AllKeys()
	if len(allKeys) == 0 {
		fmt.Println("No configuration values loaded")
		return
	}

	sort.Strings(allKeys)
	fmt.Printf("Loaded values (%d):\n", len(allKeys))

	for _, key := range allKeys {
		value := viper.Get(key)
		fmt.Printf("  %s = %s\n", key, formatConfigValue(value))
	}
}

func formatConfigValue(value interface{}) string {
	if value == nil {
		return "nil"
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return `""`
		}
		return fmt.Sprintf(`"%s"`, v)
	case bool:
		return fmt.Sprintf("%t", v)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%.2f", v)
	case []interface{}:
		if len(v) == 0 {
			return "[]"
		}
		items := make([]string, len(v))
		for i, item := range v {
			items[i] = formatConfigValue(item)
		}
		return fmt.Sprintf("[%s]", strings.Join(items, ", "))
	case map[string]interface{}:
		if len(v) == 0 {
			return "{}"
		}
		items := make([]string, 0, len(v))
		for k, val := range v {
			items = append(items, fmt.Sprintf("%s: %s", k, formatConfigValue(val)))
		}
		return fmt.Sprintf("{%s}", strings.Join(items, ", "))
	default:
		return fmt.Sprintf("%v", v)
	}
}

func getConfigValue(key string) {
	if viper.IsSet(key) {
		value := viper.GetString(key)
		fmt.Print(value)
		return
	}

	fmt.Fprintf(os.Stderr, "Error: Config key '%s' not found\n", key)
	os.Exit(1)
}
