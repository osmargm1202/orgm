package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var GobuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build and release the application",
	Long:  `Build the application for different platforms and manage GitHub releases.`,
}

// Helper function to build for Linux
func buildLinux() error {
	fmt.Println("Building for Linux...")
	goCmd := exec.Command("go", "build", "-o", "orgm", ".")
	goCmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")
	output, err := goCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error building for Linux: %s \n %s", err, output)
		return err
	}
	fmt.Printf("Successfully built for Linux: %s \n %s", "orgm", output)
	return nil
}

// Helper function to build for Windows
func buildWindows() error {
	fmt.Println("Building for Windows...")
	goCmd := exec.Command("go", "build", "-o", "orgm.exe", ".")
	goCmd.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=1", "CC=x86_64-w64-mingw32-gcc")

	output, err := goCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error building for Windows: %s \n %s", err, output)
		fmt.Println("\nNote: Windows cross-compilation requires mingw-w64 to be installed.")
		fmt.Println("On Ubuntu/Debian: sudo apt install gcc-mingw-w64-x86-64")
		fmt.Println("On Arch/Manjaro: sudo pacman -S mingw-w64-gcc")
		fmt.Println("On Fedora: sudo dnf install mingw64-gcc")
		return err
	}
	fmt.Printf("Successfully built for Windows: %s \n %s", "orgm.exe", output)
	return nil
}

// Helper function to create tag
func createTag() error {
	fmt.Println("Creating new GitHub release tag...")
	nextTag, err := getNextTag()
	if err != nil {
		fmt.Printf("Error getting next tag: %s \n", err)
		return err
	}

	fmt.Printf("Next tag will be: %s \n", nextTag)

	title := fmt.Sprintf("%s (beta)", nextTag)
	notes := "ORGM AI, SYSTEM, TOOL, CLI FOR GENERAL TASK"

	ghCmd := exec.Command("gh", "release", "create", nextTag, "--title", title, "--notes", notes)
	output, err := ghCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error creating GitHub release: %s \n %s", err, output)
		return err
	}
	fmt.Printf("Successfully created GitHub release: %s \n %s", nextTag, output)
	return nil
}

// Helper function to upload artifacts
func uploadArtifacts() error {
	fmt.Println("Uploading build artifacts...")

	linuxArtifact := "orgm"
	windowsArtifact := "orgm.exe"

	if _, err := os.Stat(linuxArtifact); os.IsNotExist(err) {
		fmt.Printf("Error: Linux artifact '%s' not found. Please build it first.\n", linuxArtifact)
		return err
	}
	if _, err := os.Stat(windowsArtifact); os.IsNotExist(err) {
		fmt.Printf("Error: Windows artifact '%s' not found. Please build it first.\n", windowsArtifact)
		return err
	}

	latestTag, err := getLatestTag()
	if err != nil {
		fmt.Printf("Error getting latest tag: %s \n", err)
		return err
	}

	deleteAsset(latestTag, linuxArtifact)
	deleteAsset(latestTag, windowsArtifact)

	fmt.Printf("Uploading artifacts to tag: %s \n", latestTag)

	ghCmd := exec.Command("gh", "release", "upload", latestTag, linuxArtifact, windowsArtifact)
	output, err := ghCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error uploading artifacts: %s \n %s", err, output)
		return err
	}
	fmt.Printf("Successfully uploaded artifacts to %s \n %s", latestTag, output)
	return nil
}

var GobuildLinuxCmd = &cobra.Command{
	Use:   "l",
	Short: "Build the application for Linux",
	Run: func(cmd *cobra.Command, args []string) {
		if err := buildLinux(); err != nil {
			os.Exit(1)
		}
	},
}

var GobuildWindowsCmd = &cobra.Command{
	Use:   "w",
	Short: "Build the application for Windows",
	Run: func(cmd *cobra.Command, args []string) {
		if err := buildWindows(); err != nil {
			os.Exit(1)
		}
	},
}

var GobuildFullCmd = &cobra.Command{
	Use:   "full",
	Short: "Build for all platforms, create tag and upload",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting full build process...")

		// Build for Linux
		if err := buildLinux(); err != nil {
			os.Exit(1)
		}

		// Build for Windows
		if err := buildWindows(); err != nil {
			os.Exit(1)
		}

		// Create tag
		if err := createTag(); err != nil {
			os.Exit(1)
		}

		// Upload artifacts
		if err := uploadArtifacts(); err != nil {
			os.Exit(1)
		}

		fmt.Println("Full build process completed successfully!")
	},
}

// Helper function to get the latest tag and increment it
func getNextTag() (string, error) {
	latestTag, err := getLatestTag()

	if err != nil {
		return "", err
	}

	// Remove 'v' prefix and convert to float
	numStr := strings.TrimPrefix(latestTag, "v")
	version, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return "", fmt.Errorf("failed to parse version number: %v", err)
	}

	// Add 0.001 to increment version
	nextVersion := version + 0.001

	// Format back to string with 3 decimal places and 'v' prefix
	return fmt.Sprintf("v%.3f", nextVersion), nil
}

// Helper function to get the latest tag
func getLatestTag() (string, error) {
	cmd := exec.Command("gh", "release", "list", "--limit", "1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to list releases: %s %s", err, output)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("no releases found")
	}

	fields := strings.Fields(lines[0])
	var latestTag string
	for _, field := range fields {
		if strings.HasPrefix(field, "v") {
			latestTag = field
			fmt.Println(latestTag)
			break
		}
	}

	if latestTag == "" && len(lines) > 1 && len(strings.Fields(lines[1])) > 1 {
		fields = strings.Fields(lines[1])
		for _, field := range fields {
			if strings.HasPrefix(field, "v") {
				parts := strings.Split(strings.TrimPrefix(field, "v"), ".")
				if len(parts) == 3 {
					latestTag = field
					break
				}
			}
		}
	}

	if latestTag == "" {
		return "", fmt.Errorf("could not determine latest tag from output: %s", string(output))
	}
	return latestTag, nil
}

func deleteAsset(tag, asset string) {
	ghCmd := exec.Command("gh", "release", "delete-asset", tag, asset)
	output, err := ghCmd.CombinedOutput()
	if err != nil {
		return
	}
	fmt.Printf("Successfully deleted asset: %s \n %s", asset, output)
}

var GhbuildTagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Create a new GitHub release tag",
	Run: func(cmd *cobra.Command, args []string) {
		if err := createTag(); err != nil {
			os.Exit(1)
		}
	},
}

var GhbuildUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload build artifacts to the latest GitHub release",
	Run: func(cmd *cobra.Command, args []string) {
		if err := uploadArtifacts(); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(GobuildCmd)
	GobuildCmd.AddCommand(GobuildLinuxCmd)
	GobuildCmd.AddCommand(GobuildWindowsCmd)
	GobuildCmd.AddCommand(GobuildFullCmd)
	GobuildCmd.AddCommand(GhbuildTagCmd)
	GobuildCmd.AddCommand(GhbuildUploadCmd)
}
