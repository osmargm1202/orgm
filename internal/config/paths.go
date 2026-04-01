// Package config provides path resolution and config bootstrapping for orgm.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	AppName = "orgm"
)

// Paths holds all resolved paths for the application
type Paths struct {
	ConfigDir    string
	GlobalConfig string
}

// GetPaths resolves and returns all orgm paths.
// Creates base directory if it doesn't exist.
func GetPaths() (*Paths, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", AppName)

	// Ensure base directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &Paths{
		ConfigDir:    configDir,
		GlobalConfig: filepath.Join(configDir, "config.yaml"),
	}, nil
}

// GetAppConfigPath returns the config path for a specific app.
// Each app gets its own subdirectory: ~/.config/orgm/<app>/config.yaml
func GetAppConfigPath(appName string) (string, error) {
	paths, err := GetPaths()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(paths.ConfigDir, appName)

	// Ensure app directory exists
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create app directory for %s: %w", appName, err)
	}

	return filepath.Join(appDir, "config.yaml"), nil
}

// EnsureGlobalConfig creates the global config file if it doesn't exist.
// Uses minimal YAML that can evolve as needed.
func EnsureGlobalConfig() error {
	paths, err := GetPaths()
	if err != nil {
		return err
	}

	if _, err := os.Stat(paths.GlobalConfig); os.IsNotExist(err) {
		// Create minimal YAML config - schema is intentionally minimal and evolvable
		content := `# orgm global configuration
# This file is managed by the program and can evolve as needed

version: 1
`
		if err := os.WriteFile(paths.GlobalConfig, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create global config: %w", err)
		}
	}
	return nil
}

// EnsureAppConfig creates the app config file if it doesn't exist.
// Uses minimal YAML that can evolve as needed.
func EnsureAppConfig(appName string) error {
	appConfigPath, err := GetAppConfigPath(appName)
	if err != nil {
		return err
	}

	if _, err := os.Stat(appConfigPath); os.IsNotExist(err) {
		// Create minimal YAML config - schema is intentionally minimal and evolvable
		content := fmt.Sprintf(`# orgm %s app configuration
# This file is managed by the program and can evolve as needed

enabled: true
`, appName)
		if err := os.WriteFile(appConfigPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create app config for %s: %w", appName, err)
		}
	}
	return nil
}

// ShowPaths displays all relevant paths
func ShowPaths() {
	paths, err := GetPaths()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	fmt.Printf("Config Directory: %s\n", paths.ConfigDir)
	fmt.Printf("Global Config:    %s\n", paths.GlobalConfig)
}

// ShowAppPath displays the path for a specific app
func ShowAppPath(appName string) {
	appConfigPath, err := GetAppConfigPath(appName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
	fmt.Printf("%s Config: %s\n", appName, appConfigPath)
}
