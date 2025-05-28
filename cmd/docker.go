package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
)

// Helper function to load environment variables from .env file
func loadLocalEnv() error {
	dotenvPath := filepath.Join(".", ".env")
	if _, err := os.Stat(dotenvPath); os.IsNotExist(err) {
		return fmt.Errorf("%s", inputs.ErrorStyle.Render(".env file not found in the current directory"))
	}

	fmt.Printf("%s\n", inputs.InfoStyle.Render("Loading environment from .env file"))
	return godotenv.Load(dotenvPath)
}

// Helper function to check if required environment variables are set
func requireVars(vars []string) error {
	var missing []string

	for _, v := range vars {
		if os.Getenv(v) == "" {
			missing = append(missing, v)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("%s: %s",
			inputs.ErrorStyle.Render("Missing required environment variables"),
			inputs.WarningStyle.Render(strings.Join(missing, ", ")))
	}

	return nil
}

// Helper function to execute docker commands
func dockerCmd(args []string, inputText string) error {
	cmd := exec.Command(args[0], args[1:]...)

	if inputText != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}

		go func() {
			defer stdin.Close()
			fmt.Fprintln(stdin, inputText)
		}()
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func DbuildCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "Build Docker image",
		Long:  "Build Docker image using cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadLocalEnv(); err != nil {
				return err
			}

			if err := requireVars([]string{"DOCKER_IMAGE_NAME", "DOCKER_IMAGE_TAG", "DOCKER_USER"}); err != nil {
				return err
			}

			tag := os.Getenv("DOCKER_IMAGE_TAG")
			image := fmt.Sprintf("%s/%s:%s", os.Getenv("DOCKER_USER"), os.Getenv("DOCKER_IMAGE_NAME"), tag)

			fmt.Printf("%s\n", inputs.InfoStyle.Render("Building image: "+image))
			return dockerCmd([]string{"docker", "build", "-t", image, "."}, "")
		},
	}
}

func DbuildNoCacheCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "build-no-cache",
		Short: "Build Docker image without cache",
		Long:  "Build Docker image without using cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadLocalEnv(); err != nil {
				return err
			}

			if err := requireVars([]string{"DOCKER_IMAGE_NAME", "DOCKER_IMAGE_TAG", "DOCKER_USER"}); err != nil {
				return err
			}

			tag := os.Getenv("DOCKER_IMAGE_TAG")
			image := fmt.Sprintf("%s/%s:%s", os.Getenv("DOCKER_USER"), os.Getenv("DOCKER_IMAGE_NAME"), tag)

			fmt.Printf("%s\n", inputs.InfoStyle.Render("Building image without cache: "+image))
			return dockerCmd([]string{"docker", "build", "--no-cache", "-t", image, "."}, "")
		},
	}
}

func DsaveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "save",
		Short: "Save Docker image to file",
		Long:  "Save Docker image to a tar file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadLocalEnv(); err != nil {
				return err
			}

			if err := requireVars([]string{
				"DOCKER_IMAGE_NAME",
				"DOCKER_IMAGE_TAG",
				"DOCKER_SAVE_FILE",
				"DOCKER_FOLDER_SAVE",
				"DOCKER_USER",
			}); err != nil {
				return err
			}

			tag := os.Getenv("DOCKER_IMAGE_TAG")
			image := fmt.Sprintf("%s/%s:%s", os.Getenv("DOCKER_USER"), os.Getenv("DOCKER_IMAGE_NAME"), tag)
			savePath := filepath.Join(os.Getenv("DOCKER_FOLDER_SAVE"), os.Getenv("DOCKER_SAVE_FILE"))

			// Create directory if it doesn't exist
			if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
				return err
			}

			fmt.Printf("%s\n", inputs.InfoStyle.Render("Saving image to: "+savePath))
			return dockerCmd([]string{"docker", "save", "-o", savePath, image}, "")
		},
	}
}

func DpushCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "Push Docker image",
		Long:  "Push Docker image to registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadLocalEnv(); err != nil {
				return err
			}

			if err := requireVars([]string{"DOCKER_IMAGE_NAME", "DOCKER_IMAGE_TAG", "DOCKER_USER", "DOCKER_URL"}); err != nil {
				return err
			}

			tag := os.Getenv("DOCKER_IMAGE_TAG")
			image := fmt.Sprintf("%s/%s/%s:%s", os.Getenv("DOCKER_URL"), os.Getenv("DOCKER_USER"), os.Getenv("DOCKER_IMAGE_NAME"), tag)

			fmt.Printf("%s\n", inputs.InfoStyle.Render("Pushing image: "+image))
			return dockerCmd([]string{"docker", "push", image}, "")
		},
	}
}

func DtagCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tag",
		Short: "Tag Docker image",
		Long:  "Tag Docker image with latest tag in registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadLocalEnv(); err != nil {
				return err
			}

			if err := requireVars([]string{"DOCKER_IMAGE_NAME", "DOCKER_IMAGE_TAG", "DOCKER_USER", "DOCKER_URL"}); err != nil {
				return err
			}

			current := fmt.Sprintf("%s/%s:%s", os.Getenv("DOCKER_USER"), os.Getenv("DOCKER_IMAGE_NAME"), os.Getenv("DOCKER_IMAGE_TAG"))
			target := fmt.Sprintf("%s/%s/%s:latest", os.Getenv("DOCKER_URL"), os.Getenv("DOCKER_USER"), os.Getenv("DOCKER_IMAGE_NAME"))

			fmt.Printf("%s\n", inputs.InfoStyle.Render("Tagging image: "+current+" → "+target))
			return dockerCmd([]string{"docker", "tag", current, target}, "")
		},
	}
}

func DcreateProdContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create-prod-context",
		Short: "Create prod Docker context",
		Long:  "Create a Docker context named 'prod'",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadLocalEnv(); err != nil {
				return err
			}

			if err := requireVars([]string{"DOCKER_HOST_USER", "DOCKER_HOST_IP"}); err != nil {
				return err
			}

			hostStr := fmt.Sprintf("ssh://%s@%s", os.Getenv("DOCKER_HOST_USER"), os.Getenv("DOCKER_HOST_IP"))

			fmt.Printf("%s\n", inputs.InfoStyle.Render("Creating prod context: "+hostStr))
			return dockerCmd([]string{"docker", "context", "create", "prod", "--docker", fmt.Sprintf("host=%s", hostStr)}, "")
		},
	}
}

func DremoveProdContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove-prod-context",
		Short: "Remove prod Docker context",
		Long:  "Remove the Docker context named 'prod'",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadLocalEnv(); err != nil {
				return err
			}

			fmt.Printf("%s\n", inputs.InfoStyle.Render("Removing prod context..."))
			return dockerCmd([]string{"docker", "context", "rm", "prod"}, "")
		},
	}
}

func DdeployCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deploy",
		Short: "Deploy application",
		Long:  "Deploy application to prod context using docker compose",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadLocalEnv(); err != nil {
				return err
			}

			if err := requireVars([]string{"DOCKER_IMAGE_NAME", "DOCKER_USER", "DOCKER_URL"}); err != nil {
				return err
			}

			image := fmt.Sprintf("%s/%s/%s:latest", os.Getenv("DOCKER_URL"), os.Getenv("DOCKER_USER"), os.Getenv("DOCKER_IMAGE_NAME"))

			fmt.Printf("%s\n", inputs.InfoStyle.Render("Deploying to prod context..."))

			// Check if prod context exists
			checkCmd := exec.Command("docker", "context", "inspect", "prod")
			if err := checkCmd.Run(); err != nil {
				fmt.Printf("%s\n", inputs.WarningStyle.Render("Prod context doesn't exist. Creating it..."))

				if err := requireVars([]string{"DOCKER_HOST_USER", "DOCKER_HOST_IP"}); err != nil {
					return fmt.Errorf("could not create prod context: %v", err)
				}

				hostStr := fmt.Sprintf("ssh://%s@%s", os.Getenv("DOCKER_HOST_USER"), os.Getenv("DOCKER_HOST_IP"))
				if err := dockerCmd([]string{"docker", "context", "create", "prod", "--docker", fmt.Sprintf("host=%s", hostStr)}, ""); err != nil {
					return err
				}
			}

			// Pull the image
			fmt.Printf("%s\n", inputs.InfoStyle.Render("Pulling image: "+image))
			if err := dockerCmd([]string{"docker", "--context", "prod", "pull", image}, ""); err != nil {
				return err
			}

			// Deploy with docker compose
			fmt.Printf("%s\n", inputs.InfoStyle.Render("Starting containers with docker compose..."))
			return dockerCmd([]string{"docker", "--context", "prod", "compose", "up", "-d", "--remove-orphans"}, "")
		},
	}
}

func DloginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Login to Docker registry",
		Long:  "Login to Docker registry using credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireVars([]string{"DOCKER_URL", "DOCKER_USER"}); err != nil {
				return err
			}

			dockerHubUrl := os.Getenv("DOCKER_URL")
			dockerHubUser := os.Getenv("DOCKER_USER")

			fmt.Printf("%s ", inputs.InfoStyle.Render("Enter Docker Hub password:"))
			reader := bufio.NewReader(os.Stdin)
			password, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			password = strings.TrimSpace(password)

			if password == "" {
				return fmt.Errorf("a password is required to continue")
			}

			fmt.Printf("%s\n", inputs.InfoStyle.Render("Logging in to "+dockerHubUrl+"..."))
			return dockerCmd([]string{"docker", "login", dockerHubUrl, "-u", dockerHubUser, "--password-stdin"}, password)
		},
	}
}

var choices = []inputs.ItemMS{
	{Title: " Build", Description: "Build Docker image using cache", Value: "build", Checked: false},
	{Title: " Build (sin cache)", Description: "Build Docker image without cache", Value: "build_no_cache", Checked: false},
	{Title: " Tag", Description: "Tag Docker image with latest tag", Value: "tag", Checked: false},
	{Title: " Save", Description: "Save Docker image to a tar file", Value: "save", Checked: false},
	{Title: " Push", Description: "Push Docker image to registry", Value: "push", Checked: false},
	{Title: " Deploy", Description: "Deploy application to prod", Value: "deploy", Checked: false},
	{Title: " Create prod context", Description: "Create a Docker context named 'prod'", Value: "create_prod_context", Checked: false},
	{Title: " Remove prod context", Description: "Remove the Docker context named 'prod'", Value: "remove_prod_context", Checked: false},
	{Title: " Login", Description: "Login to Docker registry", Value: "login", Checked: false},
	{Title: " Ayuda", Description: "Show Docker help", Value: "docker -h", Checked: false},
	{Title: " Salir", Description: "Exit the menu", Value: "exit", Checked: false},
}

func runSelectedCommands(selected []inputs.ItemMS) {
	for _, choice := range selected {
		fmt.Printf("\n%s\n", inputs.CommandStyle.Render("► Executing: "+choice.Title))

		var err error

		switch choice.Value {
		case "login":
			err = DloginCmd().RunE(nil, nil)
		case "build_no_cache":
			err = DbuildNoCacheCmd().RunE(nil, nil)
		case "build":
			err = DbuildCmd().RunE(nil, nil)
		case "tag":
			err = DtagCmd().RunE(nil, nil)
		case "save":
			err = DsaveCmd().RunE(nil, nil)
		case "push":
			err = DpushCmd().RunE(nil, nil)
		case "deploy":
			err = DdeployCmd().RunE(nil, nil)
		case "create_prod_context":
			err = DcreateProdContextCmd().RunE(nil, nil)
		case "remove_prod_context":
			err = DremoveProdContextCmd().RunE(nil, nil)
		case "docker -h":
			dockerCmd([]string{"orgm", "docker", "-h"}, "")
		case "exit":
			return
		}

		if err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("✘ Error: "+err.Error()))
		} else {
			fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ Completed successfully"))
		}
	}
}

func DmenuCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "menu",
		Short: "Interactive Docker menu",
		Long:  "Interactive menu for Docker commands",
		Run: func(cmd *cobra.Command, args []string) {
			// Banner de bienvenida
			banner := `
			ORGM DOCKER MENU`
			bannerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true)
			fmt.Println(bannerStyle.Render(banner))

			// Iniciar la aplicación BubbleTea con alt screen
			p := tea.NewProgram(inputs.InitialModelMS(choices), tea.WithAltScreen())
			m, err := p.Run()
			if err != nil {
				fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running menu: "+err.Error()))
				return
			}

			// Get the final model
			model, ok := m.(inputs.ModelMS)
			if !ok {
				fmt.Printf("%s\n", inputs.ErrorStyle.Render("Could not get the model"))
				return
			}

			if model.Quitting {
				return
			}

			// Get selected items
			var selected []inputs.ItemMS
			for i, choice := range model.Choices {
				if _, ok := model.Selected[i]; ok {
					selected = append(selected, choice)
				}
			}

			if len(selected) == 0 {
				fmt.Printf("%s\n", inputs.WarningStyle.Render("No operations selected."))
				return
			}

			// Mostrar resumen de operaciones seleccionadas
			fmt.Printf("\n%s\n", inputs.TitleStyle.Render("Selected Operations"))
			for i, choice := range selected {
				fmt.Printf("%d. %s\n", i+1, inputs.CommandStyle.Render(choice.Title))
			}
			fmt.Println()

			// Run selected commands
			runSelectedCommands(selected)
		},
	}
}

func DockerCmd() *cobra.Command {
	dockerCmd := &cobra.Command{
		Use:   "docker",
		Short: "Docker commands",
		Long:  `Commands for Docker operations like build, tag, push, deploy, etc.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Si se ejecuta sin argumentos, mostrar el menú interactivo
			DmenuCmd().Run(cmd, args)
		},
	}

	// Add subcommands
	dockerCmd.AddCommand(
		DbuildCmd(),
		DbuildNoCacheCmd(),
		DsaveCmd(),
		DpushCmd(),
		DtagCmd(),
		DcreateProdContextCmd(),
		DremoveProdContextCmd(),
		DdeployCmd(),
		DloginCmd(),
		DmenuCmd(),
	)

	return dockerCmd
}

func init() {

	RootCmd.AddCommand(DockerCmd())
}
