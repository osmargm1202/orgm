package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// viperCmd represents the viper command
var viperCmd = &cobra.Command{
	Use:   "viper",
	Short: "Show all loaded configuration variables from config files",
	Long:  `Display all configuration variables loaded from config.toml, links.toml, and keys.toml files.`,
	Run: func(cmd *cobra.Command, args []string) {
		showViperConfig()
	},
}

func showViperConfig() {
	fmt.Printf("%s\n", inputs.TitleStyle.Render("ORGM - Configuration Variables"))
	fmt.Println()

	// Show config file locations
	fmt.Printf("%s\n", inputs.SubtitleStyle.Render("Configuration Files:"))
	
	configFile := viper.ConfigFileUsed()
	if configFile != "" {
		fmt.Printf("  ✓ config.toml: %s\n", configFile)
	} else {
		fmt.Printf("  ✗ config.toml: Not found\n")
	}

    // Deprecated: links.toml and keys.toml are no longer used

	fmt.Println()

	// Get all keys from viper
	allKeys := viper.AllKeys()
	if len(allKeys) == 0 {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("No configuration variables found"))
		return
	}

	// Sort keys for better readability
	sort.Strings(allKeys)

	fmt.Printf("%s\n", inputs.SubtitleStyle.Render("Loaded Variables:"))
	fmt.Printf("Total variables: %d\n\n", len(allKeys))

	// Group variables by category
	categories := make(map[string][]string)
	
	for _, key := range allKeys {
		parts := strings.Split(key, ".")
		if len(parts) > 0 {
			category := parts[0]
			categories[category] = append(categories[category], key)
		} else {
			categories["root"] = append(categories["root"], key)
		}
	}

	// Display variables by category
	for category, keys := range categories {
		if category == "root" {
			fmt.Printf("%s\n", inputs.InfoStyle.Render("Root Variables:"))
		} else {
			fmt.Printf("%s\n", inputs.InfoStyle.Render(strings.Title(category) + " Variables:"))
		}
		
		for _, key := range keys {
			value := viper.Get(key)
			displayValue := formatValue(value)
			fmt.Printf("  %s = %s\n", key, displayValue)
		}
		fmt.Println()
	}
}

func formatValue(value interface{}) string {
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
			items[i] = formatValue(item)
		}
		return fmt.Sprintf("[%s]", strings.Join(items, ", "))
	case map[string]interface{}:
		if len(v) == 0 {
			return "{}"
		}
		items := make([]string, 0, len(v))
		for k, val := range v {
			items = append(items, fmt.Sprintf("%s: %s", k, formatValue(val)))
		}
		return fmt.Sprintf("{%s}", strings.Join(items, ", "))
	default:
		return fmt.Sprintf("%v", v)
	}
}

func init() {
	RootCmd.AddCommand(viperCmd)
}
