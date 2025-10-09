package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var GobuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the application for Windows and Linux",
	Long:  `Build the application for Windows and Linux platforms in the main directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting build process...")

		// Build for Linux
		if err := BuildLinux(); err != nil {
			os.Exit(1)
		}

		// Build for Windows
		if err := BuildWindows(); err != nil {
			os.Exit(1)
		}

		// Build for Prop
		if err := BuildProp(); err != nil {
			os.Exit(1)
		}

		// Build for Prop Windows
		if err := BuildPropWindows(); err != nil {
			os.Exit(1)
		}

		fmt.Println("Build process completed successfully!")
		fmt.Println("Generated files:")
		fmt.Println("  - orgm (Linux binary)")
		fmt.Println("  - orgm.exe (Windows binary)")
	},
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

func BuildProp() error {
	fmt.Println("Building for Prop...")
	goCmd := exec.Command("go", "build", "-o", "orgm-prop", ".")
	goCmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")
	output, err := goCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error building for Prop: %s \n %s", err, output)
		return err
	}

	fmt.Printf("Successfully built for Prop: %s \n %s", "orgm-prop", output)
	return nil
}

func BuildPropWindows() error {
	fmt.Println("Building for Prop Windows...")
	goCmd := exec.Command("go", "build", "-o", "orgm-prop.exe", ".")
	goCmd.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=1", "CC=x86_64-w64-mingw32-gcc")
	output, err := goCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error building for Prop Windows: %s \n %s", err, output)
		return err
	}
	fmt.Printf("Successfully built for Prop Windows: %s \n %s", "orgm-prop.exe", output)
	return nil
}


func init() {
	RootCmd.AddCommand(GobuildCmd)
}

