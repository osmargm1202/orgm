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

var GobuildLinuxCmd = &cobra.Command{
	Use:   "linux",
	Short: "Build the application for Linux",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Building for Linux...")
		goCmd := exec.Command("go", "build", "-o", "orgm", ".")
		goCmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")
		output, err := goCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error building for Linux: %s \n %s", err, output)
			os.Exit(1)
		}
		fmt.Printf("Successfully built for Linux: %s \n %s", "orgm", output)
	},
}

var GobuildWindowsCmd = &cobra.Command{
	Use:   "windows",
	Short: "Build the application for Windows",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Building for Windows...")
		goCmd := exec.Command("go", "build", "-o", "orgm.exe", ".")
		goCmd.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64")
		output, err := goCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error building for Windows: %s \n %s", err, output)
			os.Exit(1)
		}
		fmt.Printf("Successfully built for Windows: %s \n %s", "orgm.exe", output)
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

var GhbuildTagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Create a new GitHub release tag",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Creating new GitHub release tag...")
		nextTag, err := getNextTag()
		if err != nil {
			fmt.Printf("Error getting next tag: %s \n", err)
			os.Exit(1)
		}

		fmt.Printf("Next tag will be: %s \n", nextTag)

		title := fmt.Sprintf("%s (beta)", nextTag)
		notes := "ORGM AI, SYSTEM, TOOL, CLI FOR GENERAL TASK"

		ghCmd := exec.Command("gh", "release", "create", nextTag, "--title", title, "--notes", notes)
		output, err := ghCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error creating GitHub release: %s \n %s", err, output)
			os.Exit(1)
		}
		fmt.Printf("Successfully created GitHub release: %s \n %s", nextTag, output)
	},
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

var GhbuildUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload build artifacts to the latest GitHub release",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Uploading build artifacts...")

		linuxArtifact := "orgm"
		windowsArtifact := "orgm.exe"

		if _, err := os.Stat(linuxArtifact); os.IsNotExist(err) {
			fmt.Printf("Error: Linux artifact '%s' not found. Please build it first.\n", linuxArtifact)
			os.Exit(1)
		}
		if _, err := os.Stat(windowsArtifact); os.IsNotExist(err) {
			fmt.Printf("Error: Windows artifact '%s' not found. Please build it first.\n", windowsArtifact)
			os.Exit(1)
		}

		latestTag, err := getLatestTag()
		if err != nil {
			fmt.Printf("Error getting latest tag: %s \n", err)
			os.Exit(1)
		}

		deleteAsset(latestTag, linuxArtifact)
		deleteAsset(latestTag, windowsArtifact)

		fmt.Printf("Uploading artifacts to tag: %s \n", latestTag)

		ghCmd := exec.Command("gh", "release", "upload", latestTag, linuxArtifact, windowsArtifact)
		output, err := ghCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error uploading artifacts: %s \n %s", err, output)
			os.Exit(1)
		}
		fmt.Printf("Successfully uploaded artifacts to %s \n %s", latestTag, output)
	},
}

func deleteAsset(tag, asset string) {
	ghCmd := exec.Command("gh", "release", "delete-asset", tag, asset)
	output, err := ghCmd.CombinedOutput()
	if err != nil {
		return
	}
	fmt.Printf("Successfully deleted asset: %s \n %s", asset, output)

}

func init() {
	RootCmd.AddCommand(GobuildCmd)
	GobuildCmd.AddCommand(GobuildLinuxCmd)
	GobuildCmd.AddCommand(GobuildWindowsCmd)
	GobuildCmd.AddCommand(GhbuildTagCmd)
	GobuildCmd.AddCommand(GhbuildUploadCmd)
}
