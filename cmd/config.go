/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Edit Config commands with nano or editor choice, or set config file path.",
	Long:  `Edit Config commands with nano or editor choice (e.g., orgm config nano). If no editor is provided, it prompts to set the configuration file path via a text input.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			filePath, err := runInputAndGetPath()
			if err != nil {
				log.Fatalf("Error getting input: %v", err)
			}

			if filePath == "" {
				log.Println("No file path entered or input was cancelled.")
				return
			}

			err = savePath(filePath)
			if err != nil {
				log.Fatalf("Error saving file path: %v", err)
			}
		} else {
			EditConfig(args[0])
		}
	},
}

func runInputAndGetPath() (string, error) {
	model := inputs.TextInput("Enter the path to your configuration file:", "/path/to/your/file.toml")
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running bubbletea program: %w", err)
	}

	if m, ok := finalModel.(inputs.TextInputModel); ok {
		filePath := m.TextInput.Value()
		// Remove surrounding quotes if present
		if len(filePath) >= 2 {
			if (filePath[0] == '\'' && filePath[len(filePath)-1] == '\'') ||
				(filePath[0] == '"' && filePath[len(filePath)-1] == '"') {
				filePath = filePath[1 : len(filePath)-1]
			}
		}
		return filePath, nil
	}
	return "", fmt.Errorf("could not cast final model to TextInputModel")
}

func savePath(filePath string) error {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Use default config path: ~/.config/orgm/config.toml
	configDir := filepath.Join(homeDir, ".config", "orgm")

	// Ensure the source file exists and is readable
	sourceFileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", filePath, err)
	}

	// Ensure the target directory exists
	err = os.MkdirAll(configDir, 0755) // 0755 gives rwx for owner, rx for group/others
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", configDir, err)
	}

	targetFile := filepath.Join(configDir, "config.toml")

	// Write the content of sourceFileContent to targetFile
	// This will create the file if it doesn't exist, or truncate and overwrite if it does.
	err = os.WriteFile(targetFile, sourceFileContent, 0644) // 0644 gives rw for owner, r for group/others
	if err != nil {
		return fmt.Errorf("failed to write to %s: %w", targetFile, err)
	}
	log.Printf("Content of '%s' saved to %s", filePath, targetFile)
	return nil
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

func init() {
	RootCmd.AddCommand(configCmd)

}
