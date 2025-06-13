package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"io"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/osmargm1202/orgm/inputs"
)

var GobuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build and release the application",
	Long:  `Build the application for different platforms and manage GitHub releases.`,
}

// Helper function to build for Linux
func BuildLinux() error {
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
func BuildWindows() error {
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

var GobuildExeCmd = &cobra.Command{
	Use:   "exe",
	Short: "Build the application",
	Run: func(cmd *cobra.Command, args []string) {
		// Build for Linux
		if err := BuildLinux(); err != nil {
			os.Exit(1)
		}

		// Build for Windows
		if err := BuildWindows(); err != nil {
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
		if err := BuildLinux(); err != nil {
			os.Exit(1)
		}

		// Build for Windows
		if err := BuildWindows(); err != nil {
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
	GobuildCmd.AddCommand(GobuildExeCmd)
	GobuildCmd.AddCommand(GobuildFullCmd)
	GobuildCmd.AddCommand(GhbuildTagCmd)
	GobuildCmd.AddCommand(GhbuildUploadCmd)
	GobuildCmd.AddCommand(InstallCmd)
	GobuildCmd.AddCommand(UpdateCmd)
}



func installFunc() {
	fmt.Printf("%s\n", inputs.TitleStyle.Render("Installing the application in dev Linux"))


	BuildFunc()
	
	cmd := exec.Command("cp", "orgm", filepath.Join(os.Getenv("HOME"), "Nextcloud", "Apps", "bin", "orgm"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error during go install: %v\n", err)
		return
	}

	cmd = exec.Command("cp", "orgm.exe", filepath.Join(os.Getenv("HOME"), "Nextcloud", "Apps", "bin", "orgm.exe"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error during go install: %v\n", err)
		return
	}
	fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ Build and install completed successfully"))

}


func BuildFunc() {
	fmt.Printf("%s\n", inputs.TitleStyle.Render("Building the application"))

	BuildLinux()
	BuildWindows()

}

func updateFunc() {
	fmt.Printf("%s\n", inputs.TitleStyle.Render("Updating the application"))

	// Determine the download URL and installation path based on OS
	var downloadURL, installPath string

	switch runtime.GOOS {
	case "windows":
		downloadURL = "https://github.com/osmargm1202/orgm/releases/latest/download/orgm.exe"
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error getting home directory: %v\n", err)
			return
		}
		installPath = filepath.Join(homeDir, ".config", "orgm", "orgm.exe")
	case "linux":
		downloadURL = "https://github.com/osmargm1202/orgm/releases/latest/download/orgm"
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error getting home directory: %v\n", err)
			return
		}
		installPath = filepath.Join(homeDir, ".local", "bin", "orgm")
	default:
		fmt.Printf("Unsupported operating system: %s\n", runtime.GOOS)
		return
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(installPath), 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	// Download the latest version
	fmt.Printf("%s\n", inputs.InfoStyle.Render("Downloading latest version..."))

	resp, err := http.Get(downloadURL)
	if err != nil {
		fmt.Printf("Error downloading file: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: HTTP %d when downloading %s\n", resp.StatusCode, downloadURL)
		return
	}

	// Create temporary file
	tempFile := installPath + ".tmp"
	out, err := os.Create(tempFile)
	if err != nil {
		fmt.Printf("Error creating temporary file: %v\n", err)
		return
	}
	defer out.Close()

	// Copy downloaded content to temporary file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Remove(tempFile)
		return
	}

	// Close the file before moving it
	out.Close()

	// Remove existing file if it exists and replace with new one
	if _, err := os.Stat(installPath); err == nil {
		if err := os.Remove(installPath); err != nil {
			fmt.Printf("Error removing old file: %v\n", err)
			os.Remove(tempFile)
			return
		}
	}

	// Move temporary file to final location
	if err := os.Rename(tempFile, installPath); err != nil {
		fmt.Printf("Error moving file to final location: %v\n", err)
		os.Remove(tempFile)
		return
	}

	// Set executable permissions on Linux
	if runtime.GOOS == "linux" {
		if err := os.Chmod(installPath, 0755); err != nil {
			fmt.Printf("Error setting executable permissions: %v\n", err)
			return
		}
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ Application updated successfully"))
	fmt.Printf("%s %s\n", inputs.InfoStyle.Render("Updated executable location:"), installPath)
}