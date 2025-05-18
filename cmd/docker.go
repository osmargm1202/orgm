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
	"github.com/spf13/cobra"
)

// Estilos para el menú
var (
	titleStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4")).Padding(0, 1)
	subtitleStyle     = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#ABABAB"))
	cursorStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	checkedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#73F59F"))
	uncheckedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F27878"))
	itemStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#74ACDF"))
	descriptionStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#6A6A6A")).Italic(true)

	// Estilos para los mensajes de salida
	successStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#73F59F"))
	errorStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F27878"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAB26"))
	commandStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#3B9FEF"))
)

// Helper function to load environment variables from .env file
func loadLocalEnv() error {
	dotenvPath := filepath.Join(".", ".env")
	if _, err := os.Stat(dotenvPath); os.IsNotExist(err) {
		return fmt.Errorf("%s", errorStyle.Render(".env file not found in the current directory"))
	}

	fmt.Printf("%s\n", infoStyle.Render("Loading environment from .env file"))
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
			errorStyle.Render("Missing required environment variables"),
			warningStyle.Render(strings.Join(missing, ", ")))
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

func buildCmd() *cobra.Command {
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

			fmt.Printf("%s\n", infoStyle.Render("Building image: "+image))
			return dockerCmd([]string{"docker", "build", "-t", image, "."}, "")
		},
	}
}

func buildNoCacheCmd() *cobra.Command {
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

			fmt.Printf("%s\n", infoStyle.Render("Building image without cache: "+image))
			return dockerCmd([]string{"docker", "build", "--no-cache", "-t", image, "."}, "")
		},
	}
}

func saveCmd() *cobra.Command {
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

			fmt.Printf("%s\n", infoStyle.Render("Saving image to: "+savePath))
			return dockerCmd([]string{"docker", "save", "-o", savePath, image}, "")
		},
	}
}

func pushCmd() *cobra.Command {
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

			fmt.Printf("%s\n", infoStyle.Render("Pushing image: "+image))
			return dockerCmd([]string{"docker", "push", image}, "")
		},
	}
}

func tagCmd() *cobra.Command {
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

			fmt.Printf("%s\n", infoStyle.Render("Tagging image: "+current+" → "+target))
			return dockerCmd([]string{"docker", "tag", current, target}, "")
		},
	}
}

func createProdContextCmd() *cobra.Command {
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

			fmt.Printf("%s\n", infoStyle.Render("Creating prod context: "+hostStr))
			return dockerCmd([]string{"docker", "context", "create", "prod", "--docker", fmt.Sprintf("host=%s", hostStr)}, "")
		},
	}
}

func removeProdContextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove-prod-context",
		Short: "Remove prod Docker context",
		Long:  "Remove the Docker context named 'prod'",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loadLocalEnv(); err != nil {
				return err
			}

			fmt.Printf("%s\n", infoStyle.Render("Removing prod context..."))
			return dockerCmd([]string{"docker", "context", "rm", "prod"}, "")
		},
	}
}

func deployCmd() *cobra.Command {
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

			fmt.Printf("%s\n", infoStyle.Render("Deploying to prod context..."))

			// Check if prod context exists
			checkCmd := exec.Command("docker", "context", "inspect", "prod")
			if err := checkCmd.Run(); err != nil {
				fmt.Printf("%s\n", warningStyle.Render("Prod context doesn't exist. Creating it..."))

				if err := requireVars([]string{"DOCKER_HOST_USER", "DOCKER_HOST_IP"}); err != nil {
					return fmt.Errorf("could not create prod context: %v", err)
				}

				hostStr := fmt.Sprintf("ssh://%s@%s", os.Getenv("DOCKER_HOST_USER"), os.Getenv("DOCKER_HOST_IP"))
				if err := dockerCmd([]string{"docker", "context", "create", "prod", "--docker", fmt.Sprintf("host=%s", hostStr)}, ""); err != nil {
					return err
				}
			}

			// Pull the image
			fmt.Printf("%s\n", infoStyle.Render("Pulling image: "+image))
			if err := dockerCmd([]string{"docker", "--context", "prod", "pull", image}, ""); err != nil {
				return err
			}

			// Deploy with docker compose
			fmt.Printf("%s\n", infoStyle.Render("Starting containers with docker compose..."))
			return dockerCmd([]string{"docker", "--context", "prod", "compose", "up", "-d", "--remove-orphans"}, "")
		},
	}
}

func loginCmd() *cobra.Command {
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

			fmt.Printf("%s ", infoStyle.Render("Enter Docker Hub password:"))
			reader := bufio.NewReader(os.Stdin)
			password, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			password = strings.TrimSpace(password)

			if password == "" {
				return fmt.Errorf("a password is required to continue")
			}

			fmt.Printf("%s\n", infoStyle.Render("Logging in to "+dockerHubUrl+"..."))
			return dockerCmd([]string{"docker", "login", dockerHubUrl, "-u", dockerHubUser, "--password-stdin"}, password)
		},
	}
}

// Bubble tea model for Docker menu
type item struct {
	title       string
	description string
	value       string
	checked     bool
}

type model struct {
	choices  []item
	cursor   int
	selected map[int]struct{}
	quitting bool
}

func initialModel() model {
	choices := []item{
		{title: "📤 Build", description: "Build Docker image using cache", value: "build", checked: false},
		{title: "📤 Build (sin cache)", description: "Build Docker image without cache", value: "build_no_cache", checked: false},
		{title: "📤 Save", description: "Save Docker image to a tar file", value: "save", checked: false},
		{title: "📤 Push", description: "Push Docker image to registry", value: "push", checked: false},
		{title: "📤 Tag", description: "Tag Docker image with latest tag", value: "tag", checked: false},
		{title: "📤 Create prod context", description: "Create a Docker context named 'prod'", value: "create_prod_context", checked: false},
		{title: "📤 Deploy", description: "Deploy application to prod", value: "deploy", checked: false},
		{title: "📤 Remove prod context", description: "Remove the Docker context named 'prod'", value: "remove_prod_context", checked: false},
		{title: "📤 Login", description: "Login to Docker registry", value: "login", checked: false},
		{title: "🔍 Ayuda", description: "Show Docker help", value: "docker -h", checked: false},
		{title: "❌ Salir", description: "Exit the menu", value: "exit", checked: false},
	}

	return model{
		choices:  choices,
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			// Check if "Salir" is selected
			if m.choices[m.cursor].value == "exit" {
				m.quitting = true
				return m, tea.Quit
			}
			// Toggle selection
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor)
				m.choices[m.cursor].checked = false
			} else {
				m.selected[m.cursor] = struct{}{}
				m.choices[m.cursor].checked = true
			}
		case " ":
			// Toggle selection
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor)
				m.choices[m.cursor].checked = false
			} else {
				m.selected[m.cursor] = struct{}{}
				m.choices[m.cursor].checked = true
			}
		case "tab":
			// Run selected commands
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	// Título y subtítulo
	s := titleStyle.Render("ORGM DOCKER MENU") + "\n\n"
	s += subtitleStyle.Render("Select operations to execute (space to select, enter to toggle, tab to run)") + "\n\n"

	for i, choice := range m.choices {
		// Cursor
		cursor := " "
		if m.cursor == i {
			cursor = cursorStyle.Render(">")
		} else {
			cursor = " "
		}

		// Checkbox
		checked := " "
		if choice.checked {
			checked = checkedStyle.Render("✓")
		} else {
			checked = uncheckedStyle.Render(" ")
		}

		// Item title
		itemText := choice.title
		if m.cursor == i {
			itemText = selectedItemStyle.Render(itemText)
		} else {
			itemText = itemStyle.Render(itemText)
		}

		// Description
		desc := descriptionStyle.Render(choice.description)

		// Combine all parts
		s += fmt.Sprintf("%s [%s] %s - %s\n", cursor, checked, itemText, desc)
	}

	s += "\n" + helpStyle.Render("Press q to quit, space to select/deselect, and tab to run selected commands.") + "\n"

	return s
}

func runSelectedCommands(selected []item) {
	for _, choice := range selected {
		fmt.Printf("\n%s\n", commandStyle.Render("► Executing: "+choice.title))

		var err error

		switch choice.value {
		case "login":
			err = loginCmd().RunE(nil, nil)
		case "build_no_cache":
			err = buildNoCacheCmd().RunE(nil, nil)
		case "build":
			err = buildCmd().RunE(nil, nil)
		case "tag":
			err = tagCmd().RunE(nil, nil)
		case "save":
			err = saveCmd().RunE(nil, nil)
		case "push":
			err = pushCmd().RunE(nil, nil)
		case "deploy":
			err = deployCmd().RunE(nil, nil)
		case "create_prod_context":
			err = createProdContextCmd().RunE(nil, nil)
		case "remove_prod_context":
			err = removeProdContextCmd().RunE(nil, nil)
		case "docker -h":
			dockerCmd([]string{"orgm", "docker", "-h"}, "")
		case "exit":
			return
		}

		if err != nil {
			fmt.Printf("%s\n", errorStyle.Render("✘ Error: "+err.Error()))
		} else {
			fmt.Printf("%s\n", successStyle.Render("✓ Completed successfully"))
		}
	}
}

func menuCmd() *cobra.Command {
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

			// Iniciar la aplicación BubbleTea
			p := tea.NewProgram(initialModel())
			m, err := p.Run()
			if err != nil {
				fmt.Printf("%s\n", errorStyle.Render("Error running menu: "+err.Error()))
				return
			}

			// Get the final model
			model, ok := m.(model)
			if !ok {
				fmt.Printf("%s\n", errorStyle.Render("Could not get the model"))
				return
			}

			if model.quitting {
				return
			}

			// Get selected items
			var selected []item
			for i, choice := range model.choices {
				if _, ok := model.selected[i]; ok {
					selected = append(selected, choice)
				}
			}

			if len(selected) == 0 {
				fmt.Printf("%s\n", warningStyle.Render("No operations selected."))
				return
			}

			// Mostrar resumen de operaciones seleccionadas
			fmt.Printf("\n%s\n", titleStyle.Render("Selected Operations"))
			for i, choice := range selected {
				fmt.Printf("%d. %s\n", i+1, commandStyle.Render(choice.title))
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
			menuCmd().Run(cmd, args)
		},
	}

	// Add subcommands
	dockerCmd.AddCommand(
		buildCmd(),
		buildNoCacheCmd(),
		saveCmd(),
		pushCmd(),
		tagCmd(),
		createProdContextCmd(),
		removeProdContextCmd(),
		deployCmd(),
		loginCmd(),
		menuCmd(),
	)

	return dockerCmd
}

func init() {

	rootCmd.AddCommand(DockerCmd())
}
