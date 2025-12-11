package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	calcapi "github.com/osmargm1202/orgm/pkg/calc-api-management"
	"github.com/spf13/cobra"
)

// CalcAPICmd represents the calc-api command
var CalcAPICmd = &cobra.Command{
	Use:   "calc-api",
	Short: "Gestión de empresas, proyectos e ingenieros",
	Long:  `Gestión de empresas, proyectos e ingenieros con interfaz interactiva usando huh.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Ensure token for all subcommands using api_calc_management
		if _, err := EnsureGCloudIDTokenForAPI("api_calc_management"); err != nil {
			return fmt.Errorf("error obteniendo token: %w", err)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		runCalcAPITUI()
	},
}

func init() {
	RootCmd.AddCommand(CalcAPICmd)
}

// Colors: azul cielo (#87CEEB), azul oscuro (#00008B), verde (#00FF00)
var (
	calcTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00008B")). // azul oscuro
		Padding(1)
	calcTableHeaderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00008B")). // azul oscuro
		Bold(true)
	calcTableRowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	calcSuccessStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")). // verde
		Bold(true)
	calcErrorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)
)

// isCancelled checks if the error is a cancellation (Esc key or "q")
func isCancelled(err error) bool {
	if err == nil {
		return false
	}
	// Check for tea.ErrProgramKilled which is returned when Esc is pressed
	return errors.Is(err, tea.ErrProgramKilled)
}

// runFormWithCancel runs a form and returns true if cancelled (Esc or "q" pressed)
// It also checks for "q" keypress by running the form in a custom way
func runFormWithCancel(form *huh.Form) bool {
	// Create a custom program to intercept "q" key
	model := &formModel{
		form: form,
	}
	
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		if isCancelled(err) {
			return true
		}
		fmt.Printf("%s Error: %v\n", calcErrorStyle.Render("Error:"), err)
		return true
	}
	
	// Check if "q" was pressed or form was cancelled
	if finalModel != nil {
		if fm, ok := finalModel.(*formModel); ok {
			if fm.cancelled {
				return true
			}
			// Check if form completed successfully
			if fm.form.State == huh.StateCompleted {
				return false // Form completed successfully
			}
		}
	}
	
	// If we get here and form is not completed, it was cancelled
	if model.form.State != huh.StateCompleted {
		return true
	}
	
	return false
}

// formModel wraps huh.Form to intercept "q" key
type formModel struct {
	form      *huh.Form
	cancelled bool
}

func (m *formModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m *formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// First check for cancel keys before delegating to form
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Intercept "q" key to cancel (case insensitive)
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			key := strings.ToLower(string(msg.Runes[0]))
			if key == "q" {
				m.cancelled = true
				return m, tea.Quit
			}
		}
		// Also check for Esc key
		if msg.Type == tea.KeyEsc {
			m.cancelled = true
			return m, tea.Quit
		}
	}
	
	// Delegate to form - huh.Form implements tea.Model as pointer
	var cmd tea.Cmd
	updatedForm, cmd := m.form.Update(msg)
	
	// Check if form is a pointer to huh.Form
	if formPtr, ok := updatedForm.(*huh.Form); ok {
		m.form = formPtr
		// Check if form completed successfully
		if m.form.State == huh.StateCompleted {
			// Form completed, quit the program
			return m, tea.Quit
		}
	}
	
	return m, cmd
}

func (m *formModel) View() string {
	view := m.form.View()
	// Replace the default help text with our custom one that includes "q"
	// Find and replace the help text line
	lines := strings.Split(view, "\n")
	for i, line := range lines {
		if strings.Contains(line, "↑ up • ↓ down • / filter • enter submit") {
			// Replace with our custom help text
			lines[i] = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Render("↑ up • ↓ down • / filter • enter submit • q/Esc cancel")
			break
		}
	}
	return strings.Join(lines, "\n")
}

func runCalcAPITUI() {
	// Create client
	client, err := createCalcAPIClient()
	if err != nil {
		fmt.Printf("%s Error creando cliente: %v\n", calcErrorStyle.Render("Error:"), err)
		os.Exit(1)
	}

	for {
		// Main menu
		var action string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("🔢 Gestión de Datos").
				Options(
					huh.NewOption("➕ Crear", "create"),
					huh.NewOption("✏️ Editar", "edit"),
					huh.NewOption("🗑️ Eliminar", "delete"),
					huh.NewOption("📄 Crear datos.json", "create-datos"),
					huh.NewOption("🚪 Salir", "exit"),
				).
					Value(&action),
			),
		)

		// Use runFormWithCancel to handle Esc and "q" keys
		if runFormWithCancel(form) {
			// If Esc or "q" is pressed in main menu, exit program
			return
		}

		if action == "exit" {
			return
		}

		if action == "create-datos" {
			createDatosJSON(client)
			continue
		}

		// Submenu: select entity type
		var entityType string
		form = huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Seleccionar Tipo").
					Options(
						huh.NewOption("🏢 Empresas", "empresas"),
						huh.NewOption("📁 Proyectos", "proyectos"),
						huh.NewOption("👤 Ingenieros", "ingenieros"),
					).
					Value(&entityType),
			),
		)

		// Use runFormWithCancel to handle Esc and "q" keys
		if runFormWithCancel(form) {
			// If Esc or "q" is pressed in submenu, return to main menu
			continue
		}

		switch action {
		case "create":
			createFlow(entityType, client)
		case "edit":
			editFlow(entityType, client)
		case "delete":
			deleteFlow(entityType, client)
		}
	}
}

func createCalcAPIClient() (*calcapi.Client, error) {
	baseURL, err := calcapi.GetBaseURL()
	if err != nil {
		return nil, err
	}

	// Create auth function that uses EnsureGCloudIDTokenForAPI
	authFunc := func(req *http.Request) {
		token, err := EnsureGCloudIDTokenForAPI("api_calc_management")
		if err != nil || token == "" {
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return calcapi.NewClient(baseURL, authFunc), nil
}

// Exported functions for other modules
func GetEmpresas(client *calcapi.Client) ([]calcapi.Empresa, error) {
	return client.GetEmpresas()
}

func GetProyectos(client *calcapi.Client) ([]calcapi.Proyecto, error) {
	return client.GetProyectos()
}

func GetIngenieros(client *calcapi.Client) ([]calcapi.Ingeniero, error) {
	return client.GetIngenieros()
}

// Create flow
func createFlow(entityType string, client *calcapi.Client) {
	switch entityType {
	case "empresas":
		createEmpresa(client)
	case "proyectos":
		createProyecto(client)
	case "ingenieros":
		createIngeniero(client)
	}
}

func createEmpresa(client *calcapi.Client) {
	var nombre string
	var logoURL string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("➕ Crear Empresa").
				Prompt("Nombre:").
				Value(&nombre),
			huh.NewInput().
				Title("URL del Logo (opcional)").
				Prompt("URL:").
				Value(&logoURL),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		fmt.Printf("%s El nombre es requerido\n", calcErrorStyle.Render("Error:"))
		return
	}

	var logoPtr *string
	if strings.TrimSpace(logoURL) != "" {
		logoPtr = &logoURL
	}

	// Show summary
	fmt.Println("\n" + calcTitleStyle.Render("📊 Resumen"))
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("%s %s\n", calcTableHeaderStyle.Render("Nombre:"), nombre)
	if logoPtr != nil {
		fmt.Printf("%s %s\n", calcTableHeaderStyle.Render("Logo URL:"), *logoPtr)
	} else {
		fmt.Printf("%s %s\n", calcTableHeaderStyle.Render("Logo URL:"), "N/A")
	}
	fmt.Println(strings.Repeat("═", 60) + "\n")

	var confirmar bool
	form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Confirmar Creación").
				Description("¿Crear empresa con estos datos?").
				Value(&confirmar),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	if !confirmar {
		fmt.Println("Creación cancelada.")
		return
	}

	empresa, err := client.CreateEmpresa(calcapi.CreateEmpresaRequest{
		Nombre:  nombre,
		URLLogo: logoPtr,
	})
	if err != nil {
		fmt.Printf("%s Error al crear empresa: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	fmt.Printf("%s Empresa creada exitosamente! ID: %d\n", calcSuccessStyle.Render("✓"), empresa.ID)
}

func createProyecto(client *calcapi.Client) {
	// First, show empresas table and select empresa
	empresas, err := client.GetEmpresas()
	if err != nil {
		fmt.Printf("%s Error al cargar empresas: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	if len(empresas) == 0 {
		fmt.Printf("%s No hay empresas disponibles. Crea una empresa primero.\n", calcErrorStyle.Render("Error:"))
		return
	}

	showEmpresasTable(empresas)

	var empresaOption string
	options := []huh.Option[string]{}
	for _, emp := range empresas {
		options = append(options, huh.NewOption(fmt.Sprintf("%s (ID: %d)", emp.Nombre, emp.ID), fmt.Sprintf("%d", emp.ID)))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Seleccionar Empresa").
				Options(options...).
				Value(&empresaOption),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	empresaID, _ := strconv.Atoi(empresaOption)

	// Form for proyecto data
	var nombre string
	var cliente string
	var ubicacion string
	var urlLogo string

	form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("➕ Crear Proyecto").
				Prompt("Nombre:").
				Value(&nombre),
			huh.NewInput().
				Prompt("Cliente:").
				Value(&cliente),
			huh.NewInput().
				Prompt("Ubicación:").
				Value(&ubicacion),
			huh.NewInput().
				Title("URL del Logo del Cliente (opcional)").
				Prompt("URL:").
				Value(&urlLogo),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	nombre = strings.TrimSpace(nombre)
	cliente = strings.TrimSpace(cliente)
	ubicacion = strings.TrimSpace(ubicacion)
	urlLogo = strings.TrimSpace(urlLogo)

	if nombre == "" || cliente == "" || ubicacion == "" {
		fmt.Printf("%s Todos los campos son requeridos\n", calcErrorStyle.Render("Error:"))
		return
	}

	var urlLogoPtr *string
	if urlLogo != "" {
		urlLogoPtr = &urlLogo
	}

	// Get empresa name for summary
	var empresaNombre string
	for _, emp := range empresas {
		if emp.ID == empresaID {
			empresaNombre = emp.Nombre
			break
		}
	}

	// Show summary
	fmt.Println("\n" + calcTitleStyle.Render("📊 Resumen"))
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("%s %s (ID: %d)\n", calcTableHeaderStyle.Render("Empresa:"), empresaNombre, empresaID)
	fmt.Printf("%s %s\n", calcTableHeaderStyle.Render("Nombre:"), nombre)
	fmt.Printf("%s %s\n", calcTableHeaderStyle.Render("Cliente:"), cliente)
	fmt.Printf("%s %s\n", calcTableHeaderStyle.Render("Ubicación:"), ubicacion)
	if urlLogoPtr != nil {
		fmt.Printf("%s %s\n", calcTableHeaderStyle.Render("URL Logo Cliente:"), *urlLogoPtr)
	} else {
		fmt.Printf("%s %s\n", calcTableHeaderStyle.Render("URL Logo Cliente:"), "N/A")
	}
	fmt.Println(strings.Repeat("═", 60) + "\n")

	var confirmar bool
	form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Confirmar Creación").
				Description("¿Crear proyecto con estos datos?").
				Value(&confirmar),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	if !confirmar {
		fmt.Println("Creación cancelada.")
		return
	}

	proyecto, err := client.CreateProyecto(calcapi.CreateProyectoRequest{
		Nombre:    nombre,
		Cliente:   cliente,
		Ubicacion: ubicacion,
		URLLogo:   urlLogoPtr,
		EmpresaID: empresaID,
	})
	if err != nil {
		fmt.Printf("%s Error al crear proyecto: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	fmt.Printf("%s Proyecto creado exitosamente! ID: %d\n", calcSuccessStyle.Render("✓"), proyecto.ID)
}

func createIngeniero(client *calcapi.Client) {
	var nombre string
	var profesion string
	var codia string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("➕ Crear Ingeniero").
				Prompt("Nombre:").
				Value(&nombre),
			huh.NewInput().
				Prompt("Profesión:").
				Value(&profesion),
			huh.NewInput().
				Prompt("CODIA:").
				Value(&codia),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	nombre = strings.TrimSpace(nombre)
	profesion = strings.TrimSpace(profesion)
	codia = strings.TrimSpace(codia)

	if nombre == "" || profesion == "" || codia == "" {
		fmt.Printf("%s Todos los campos son requeridos\n", calcErrorStyle.Render("Error:"))
		return
	}

	// Show summary
	fmt.Println("\n" + calcTitleStyle.Render("📊 Resumen"))
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("%s %s\n", calcTableHeaderStyle.Render("Nombre:"), nombre)
	fmt.Printf("%s %s\n", calcTableHeaderStyle.Render("Profesión:"), profesion)
	fmt.Printf("%s %s\n", calcTableHeaderStyle.Render("CODIA:"), codia)
	fmt.Println(strings.Repeat("═", 60) + "\n")

	var confirmar bool
	form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Confirmar Creación").
				Description("¿Crear ingeniero con estos datos?").
				Value(&confirmar),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	if !confirmar {
		fmt.Println("Creación cancelada.")
		return
	}

	ingeniero, err := client.CreateIngeniero(calcapi.CreateIngenieroRequest{
		Nombre:    nombre,
		Profesion: profesion,
		CODIA:     codia,
	})
	if err != nil {
		fmt.Printf("%s Error al crear ingeniero: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	fmt.Printf("%s Ingeniero creado exitosamente! ID: %d\n", calcSuccessStyle.Render("✓"), ingeniero.ID)
}

// Edit flow
func editFlow(entityType string, client *calcapi.Client) {
	switch entityType {
	case "empresas":
		editEmpresa(client)
	case "proyectos":
		editProyecto(client)
	case "ingenieros":
		editIngeniero(client)
	}
}

func editEmpresa(client *calcapi.Client) {
	empresas, err := client.GetEmpresas()
	if err != nil {
		fmt.Printf("%s Error al cargar empresas: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	if len(empresas) == 0 {
		fmt.Printf("%s No hay empresas disponibles\n", calcErrorStyle.Render("Error:"))
		return
	}

	showEmpresasTable(empresas)

	var empresaOption string
	options := []huh.Option[string]{}
	for _, emp := range empresas {
		options = append(options, huh.NewOption(fmt.Sprintf("%s (ID: %d)", emp.Nombre, emp.ID), fmt.Sprintf("%d", emp.ID)))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Seleccionar Empresa a Editar").
				Options(options...).
				Value(&empresaOption),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	empresaID, _ := strconv.Atoi(empresaOption)

	// Get current empresa
	var currentEmpresa *calcapi.Empresa
	for i := range empresas {
		if empresas[i].ID == empresaID {
			currentEmpresa = &empresas[i]
			break
		}
	}

	if currentEmpresa == nil {
		fmt.Printf("%s Empresa no encontrada\n", calcErrorStyle.Render("Error:"))
		return
	}

	// Form with current values
	nombre := currentEmpresa.Nombre
	logoURL := ""
	if currentEmpresa.URLLogo != nil {
		logoURL = *currentEmpresa.URLLogo
	}

	form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("✏️ Editar Empresa").
				Prompt("Nombre:").
				Value(&nombre),
			huh.NewInput().
				Title("URL del Logo (opcional)").
				Prompt("URL:").
				Value(&logoURL),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		fmt.Printf("%s El nombre es requerido\n", calcErrorStyle.Render("Error:"))
		return
	}

	var logoPtr *string
	if strings.TrimSpace(logoURL) != "" {
		logoPtr = &logoURL
	}

	// Show summary of changes
	fmt.Println("\n" + calcTitleStyle.Render("📊 Resumen de Cambios"))
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("%s %s → %s\n", calcTableHeaderStyle.Render("Nombre:"), currentEmpresa.Nombre, nombre)
	oldLogo := "N/A"
	if currentEmpresa.URLLogo != nil {
		oldLogo = *currentEmpresa.URLLogo
	}
	newLogo := "N/A"
	if logoPtr != nil {
		newLogo = *logoPtr
	}
	fmt.Printf("%s %s → %s\n", calcTableHeaderStyle.Render("Logo URL:"), oldLogo, newLogo)
	fmt.Println(strings.Repeat("═", 60) + "\n")

	var confirmar bool
	form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Confirmar Edición").
				Description("¿Actualizar empresa con estos datos?").
				Value(&confirmar),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	if !confirmar {
		fmt.Println("Edición cancelada.")
		return
	}

	empresa, err := client.UpdateEmpresa(empresaID, calcapi.CreateEmpresaRequest{
		Nombre:  nombre,
		URLLogo: logoPtr,
	})
	if err != nil {
		fmt.Printf("%s Error al actualizar empresa: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	fmt.Printf("%s Empresa actualizada exitosamente! ID: %d\n", calcSuccessStyle.Render("✓"), empresa.ID)
}

func editProyecto(client *calcapi.Client) {
	proyectos, err := client.GetProyectos()
	if err != nil {
		fmt.Printf("%s Error al cargar proyectos: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	if len(proyectos) == 0 {
		fmt.Printf("%s No hay proyectos disponibles\n", calcErrorStyle.Render("Error:"))
		return
	}

	showProyectosTable(proyectos)

	var proyectoOption string
	options := []huh.Option[string]{}
	for _, proj := range proyectos {
		options = append(options, huh.NewOption(fmt.Sprintf("%s - %s (ID: %d)", proj.Nombre, proj.Cliente, proj.ID), fmt.Sprintf("%d", proj.ID)))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Seleccionar Proyecto a Editar").
				Options(options...).
				Value(&proyectoOption),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	proyectoID, _ := strconv.Atoi(proyectoOption)

	// Get current proyecto
	proyecto, err := client.GetProyecto(proyectoID)
	if err != nil {
		fmt.Printf("%s Error al cargar proyecto: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	// Load empresas for selection
	empresas, err := client.GetEmpresas()
	if err != nil {
		fmt.Printf("%s Error al cargar empresas: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	showEmpresasTable(empresas)

	// Select empresa
	var empresaOption string
	empresaOptions := []huh.Option[string]{}
	for _, emp := range empresas {
		empresaOptions = append(empresaOptions, huh.NewOption(fmt.Sprintf("%s (ID: %d)", emp.Nombre, emp.ID), fmt.Sprintf("%d", emp.ID)))
	}

	// Preselect current empresa
	empresaOption = fmt.Sprintf("%d", proyecto.EmpresaID)

	form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Seleccionar Empresa").
				Options(empresaOptions...).
				Value(&empresaOption),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	empresaID, _ := strconv.Atoi(empresaOption)

	// Form with current values
	nombre := proyecto.Nombre
	cliente := proyecto.Cliente
	ubicacion := proyecto.Ubicacion
	urlLogo := ""
	if proyecto.URLLogo != nil {
		urlLogo = *proyecto.URLLogo
	}

	form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("✏️ Editar Proyecto").
				Prompt("Nombre:").
				Value(&nombre),
			huh.NewInput().
				Prompt("Cliente:").
				Value(&cliente),
			huh.NewInput().
				Prompt("Ubicación:").
				Value(&ubicacion),
			huh.NewInput().
				Title("URL del Logo del Cliente (opcional)").
				Prompt("URL:").
				Value(&urlLogo),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	nombre = strings.TrimSpace(nombre)
	cliente = strings.TrimSpace(cliente)
	ubicacion = strings.TrimSpace(ubicacion)
	urlLogo = strings.TrimSpace(urlLogo)

	if nombre == "" || cliente == "" || ubicacion == "" {
		fmt.Printf("%s Todos los campos son requeridos\n", calcErrorStyle.Render("Error:"))
		return
	}

	var urlLogoPtr *string
	if urlLogo != "" {
		urlLogoPtr = &urlLogo
	}

	// Get empresa name for summary
	var empresaNombre string
	for _, emp := range empresas {
		if emp.ID == empresaID {
			empresaNombre = emp.Nombre
			break
		}
	}

	// Show summary of changes
	fmt.Println("\n" + calcTitleStyle.Render("📊 Resumen de Cambios"))
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("%s %s (ID: %d) → %s (ID: %d)\n", calcTableHeaderStyle.Render("Empresa:"), proyecto.Empresa.Nombre, proyecto.EmpresaID, empresaNombre, empresaID)
	fmt.Printf("%s %s → %s\n", calcTableHeaderStyle.Render("Nombre:"), proyecto.Nombre, nombre)
	fmt.Printf("%s %s → %s\n", calcTableHeaderStyle.Render("Cliente:"), proyecto.Cliente, cliente)
	fmt.Printf("%s %s → %s\n", calcTableHeaderStyle.Render("Ubicación:"), proyecto.Ubicacion, ubicacion)
	oldLogo := "N/A"
	if proyecto.URLLogo != nil {
		oldLogo = *proyecto.URLLogo
	}
	newLogo := "N/A"
	if urlLogoPtr != nil {
		newLogo = *urlLogoPtr
	}
	fmt.Printf("%s %s → %s\n", calcTableHeaderStyle.Render("URL Logo Cliente:"), oldLogo, newLogo)
	fmt.Println(strings.Repeat("═", 60) + "\n")

	var confirmar bool
	form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Confirmar Edición").
				Description("¿Actualizar proyecto con estos datos?").
				Value(&confirmar),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	if !confirmar {
		fmt.Println("Edición cancelada.")
		return
	}

	updatedProyecto, err := client.UpdateProyecto(proyectoID, calcapi.CreateProyectoRequest{
		Nombre:    nombre,
		Cliente:   cliente,
		Ubicacion: ubicacion,
		URLLogo:   urlLogoPtr,
		EmpresaID: empresaID,
	})
	if err != nil {
		fmt.Printf("%s Error al actualizar proyecto: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	fmt.Printf("%s Proyecto actualizado exitosamente! ID: %d\n", calcSuccessStyle.Render("✓"), updatedProyecto.ID)
}

func editIngeniero(client *calcapi.Client) {
	ingenieros, err := client.GetIngenieros()
	if err != nil {
		fmt.Printf("%s Error al cargar ingenieros: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	if len(ingenieros) == 0 {
		fmt.Printf("%s No hay ingenieros disponibles\n", calcErrorStyle.Render("Error:"))
		return
	}

	showIngenierosTable(ingenieros)

	var ingenieroOption string
	options := []huh.Option[string]{}
	for _, ing := range ingenieros {
		options = append(options, huh.NewOption(fmt.Sprintf("%s - %s (ID: %d)", ing.Nombre, ing.Profesion, ing.ID), fmt.Sprintf("%d", ing.ID)))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Seleccionar Ingeniero a Editar").
				Options(options...).
				Value(&ingenieroOption),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	ingenieroID, _ := strconv.Atoi(ingenieroOption)

	// Get current ingeniero
	var currentIngeniero *calcapi.Ingeniero
	for i := range ingenieros {
		if ingenieros[i].ID == ingenieroID {
			currentIngeniero = &ingenieros[i]
			break
		}
	}

	if currentIngeniero == nil {
		fmt.Printf("%s Ingeniero no encontrado\n", calcErrorStyle.Render("Error:"))
		return
	}

	// Form with current values
	nombre := currentIngeniero.Nombre
	profesion := currentIngeniero.Profesion
	codia := currentIngeniero.CODIA

	form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("✏️ Editar Ingeniero").
				Prompt("Nombre:").
				Value(&nombre),
			huh.NewInput().
				Prompt("Profesión:").
				Value(&profesion),
			huh.NewInput().
				Prompt("CODIA:").
				Value(&codia),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	nombre = strings.TrimSpace(nombre)
	profesion = strings.TrimSpace(profesion)
	codia = strings.TrimSpace(codia)

	if nombre == "" || profesion == "" || codia == "" {
		fmt.Printf("%s Todos los campos son requeridos\n", calcErrorStyle.Render("Error:"))
		return
	}

	// Show summary of changes
	fmt.Println("\n" + calcTitleStyle.Render("📊 Resumen de Cambios"))
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("%s %s → %s\n", calcTableHeaderStyle.Render("Nombre:"), currentIngeniero.Nombre, nombre)
	fmt.Printf("%s %s → %s\n", calcTableHeaderStyle.Render("Profesión:"), currentIngeniero.Profesion, profesion)
	fmt.Printf("%s %s → %s\n", calcTableHeaderStyle.Render("CODIA:"), currentIngeniero.CODIA, codia)
	fmt.Println(strings.Repeat("═", 60) + "\n")

	var confirmar bool
	form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Confirmar Edición").
				Description("¿Actualizar ingeniero con estos datos?").
				Value(&confirmar),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	if !confirmar {
		fmt.Println("Edición cancelada.")
		return
	}

	ingeniero, err := client.UpdateIngeniero(ingenieroID, calcapi.CreateIngenieroRequest{
		Nombre:    nombre,
		Profesion: profesion,
		CODIA:     codia,
	})
	if err != nil {
		fmt.Printf("%s Error al actualizar ingeniero: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	fmt.Printf("%s Ingeniero actualizado exitosamente! ID: %d\n", calcSuccessStyle.Render("✓"), ingeniero.ID)
}

// Delete flow
func deleteFlow(entityType string, client *calcapi.Client) {
	switch entityType {
	case "empresas":
		fmt.Printf("%s Las empresas no pueden ser eliminadas por seguridad e integridad de datos.\n", calcErrorStyle.Render("Error:"))
	case "proyectos":
		deleteProyecto(client)
	case "ingenieros":
		fmt.Printf("%s Los ingenieros no pueden ser eliminados por seguridad e integridad de datos.\n", calcErrorStyle.Render("Error:"))
	}
}

func deleteProyecto(client *calcapi.Client) {
	proyectos, err := client.GetProyectos()
	if err != nil {
		fmt.Printf("%s Error al cargar proyectos: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	if len(proyectos) == 0 {
		fmt.Printf("%s No hay proyectos disponibles\n", calcErrorStyle.Render("Error:"))
		return
	}

	showProyectosTable(proyectos)

	var proyectoOption string
	options := []huh.Option[string]{}
	for _, proj := range proyectos {
		options = append(options, huh.NewOption(fmt.Sprintf("%s - %s (ID: %d)", proj.Nombre, proj.Cliente, proj.ID), fmt.Sprintf("%d", proj.ID)))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Seleccionar Proyecto a Eliminar").
				Options(options...).
				Value(&proyectoOption),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	proyectoID, _ := strconv.Atoi(proyectoOption)

	// Get proyecto details
	proyecto, err := client.GetProyecto(proyectoID)
	if err != nil {
		fmt.Printf("%s Error al cargar proyecto: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	// Show proyecto information
	fmt.Println("\n" + calcTitleStyle.Render("🗑️ Información del Proyecto a Eliminar"))
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("%s %d\n", calcTableHeaderStyle.Render("ID:"), proyecto.ID)
	fmt.Printf("%s %s\n", calcTableHeaderStyle.Render("Nombre:"), proyecto.Nombre)
	fmt.Printf("%s %s\n", calcTableHeaderStyle.Render("Cliente:"), proyecto.Cliente)
	fmt.Printf("%s %s\n", calcTableHeaderStyle.Render("Ubicación:"), proyecto.Ubicacion)
	fmt.Printf("%s %s (ID: %d)\n", calcTableHeaderStyle.Render("Empresa:"), proyecto.Empresa.Nombre, proyecto.EmpresaID)
	fmt.Println(strings.Repeat("═", 60) + "\n")

	var confirmar bool
	form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Confirmar Eliminación").
				Description("¿Estás seguro de que deseas eliminar este proyecto? Esta acción no se puede deshacer.").
				Value(&confirmar),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	if !confirmar {
		fmt.Println("Eliminación cancelada.")
		return
	}

	err = client.DeleteProyecto(proyectoID)
	if err != nil {
		fmt.Printf("%s Error al eliminar proyecto: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	fmt.Printf("%s Proyecto eliminado exitosamente! ID: %d\n", calcSuccessStyle.Render("✓"), proyectoID)
}

func createDatosJSON(client *calcapi.Client) {
	// Step 1: Select proyecto
	proyectos, err := client.GetProyectos()
	if err != nil {
		fmt.Printf("%s Error al cargar proyectos: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	if len(proyectos) == 0 {
		fmt.Printf("%s No hay proyectos disponibles\n", calcErrorStyle.Render("Error:"))
		return
	}

	showProyectosTable(proyectos)

	var proyectoOption string
	options := []huh.Option[string]{}
	for _, proj := range proyectos {
		options = append(options, huh.NewOption(fmt.Sprintf("%s - %s (ID: %d)", proj.Nombre, proj.Cliente, proj.ID), fmt.Sprintf("%d", proj.ID)))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("📄 Crear datos.json - Seleccionar Proyecto").
				Options(options...).
				Value(&proyectoOption),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	proyectoID, _ := strconv.Atoi(proyectoOption)

	// Get complete proyecto data (includes empresa)
	proyecto, err := client.GetProyecto(proyectoID)
	if err != nil {
		fmt.Printf("%s Error al cargar proyecto: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	// Step 2: Select ingeniero
	ingenieros, err := client.GetIngenieros()
	if err != nil {
		fmt.Printf("%s Error al cargar ingenieros: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	if len(ingenieros) == 0 {
		fmt.Printf("%s No hay ingenieros disponibles\n", calcErrorStyle.Render("Error:"))
		return
	}

	showIngenierosTable(ingenieros)

	var ingenieroOption string
	ingenieroOptions := []huh.Option[string]{}
	for _, ing := range ingenieros {
		ingenieroOptions = append(ingenieroOptions, huh.NewOption(fmt.Sprintf("%s - %s (ID: %d)", ing.Nombre, ing.Profesion, ing.ID), fmt.Sprintf("%d", ing.ID)))
	}

	form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Seleccionar Ingeniero").
				Options(ingenieroOptions...).
				Value(&ingenieroOption),
		),
	)

	if runFormWithCancel(form) {
		return
	}

	ingenieroID, _ := strconv.Atoi(ingenieroOption)

	// Get complete ingeniero data
	ingeniero, err := client.GetIngeniero(ingenieroID)
	if err != nil {
		fmt.Printf("%s Error al cargar ingeniero: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	// Ensure empresa is loaded
	if proyecto.Empresa == nil {
		empresa, err := client.GetEmpresa(proyecto.EmpresaID)
		if err != nil {
			fmt.Printf("%s Error al cargar empresa: %v\n", calcErrorStyle.Render("Error:"), err)
			return
		}
		proyecto.Empresa = empresa
	}

	// Create JSON structure
	datosJSON := map[string]interface{}{
		"proyecto": map[string]interface{}{
			"id":         proyecto.ID,
			"nombre":     proyecto.Nombre,
			"cliente":    proyecto.Cliente,
			"ubicacion":  proyecto.Ubicacion,
			"url_logo":   proyecto.URLLogo,
			"empresa_id": proyecto.EmpresaID,
			"created_at": proyecto.CreatedAt.Format(time.RFC3339),
			"updated_at": proyecto.UpdatedAt.Format(time.RFC3339),
		},
		"empresa": map[string]interface{}{
			"id":         proyecto.Empresa.ID,
			"nombre":     proyecto.Empresa.Nombre,
			"url_logo":   proyecto.Empresa.URLLogo,
			"created_at": proyecto.Empresa.CreatedAt.Format(time.RFC3339),
			"updated_at": proyecto.Empresa.UpdatedAt.Format(time.RFC3339),
		},
		"ingeniero": map[string]interface{}{
			"id":         ingeniero.ID,
			"nombre":     ingeniero.Nombre,
			"profesion":  ingeniero.Profesion,
			"codia":      ingeniero.CODIA,
			"created_at": ingeniero.CreatedAt.Format(time.RFC3339),
			"updated_at": ingeniero.UpdatedAt.Format(time.RFC3339),
		},
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(datosJSON, "", "  ")
	if err != nil {
		fmt.Printf("%s Error al serializar JSON: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("%s Error al obtener directorio actual: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	// Create file path
	filePath := filepath.Join(cwd, "datos.json")

	// Write file
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		fmt.Printf("%s Error al escribir archivo: %v\n", calcErrorStyle.Render("Error:"), err)
		return
	}

	fmt.Printf("%s Archivo datos.json creado exitosamente en: %s\n", calcSuccessStyle.Render("✓"), filePath)
}

// Table display functions using lipgloss
func showEmpresasTable(empresas []calcapi.Empresa) {
	if len(empresas) == 0 {
		return
	}

	fmt.Println("\n" + calcTitleStyle.Render("📋 Empresas Disponibles"))
	fmt.Println(strings.Repeat("─", 80))
	
	header := fmt.Sprintf("%-5s %-30s %-40s", "ID", "Nombre", "Logo URL")
	fmt.Println(calcTableHeaderStyle.Render(header))
	fmt.Println(strings.Repeat("─", 80))

	for _, emp := range empresas {
		logo := "N/A"
		if emp.URLLogo != nil && *emp.URLLogo != "" {
			logo = *emp.URLLogo
			if len(logo) > 38 {
				logo = logo[:35] + "..."
			}
		}
		row := fmt.Sprintf("%-5d %-30s %-40s", emp.ID, emp.Nombre, logo)
		fmt.Println(calcTableRowStyle.Render(row))
	}
	fmt.Println(strings.Repeat("─", 80) + "\n")
}

func showProyectosTable(proyectos []calcapi.Proyecto) {
	if len(proyectos) == 0 {
		return
	}

	fmt.Println("\n" + calcTitleStyle.Render("📋 Proyectos Disponibles"))
	fmt.Println(strings.Repeat("─", 100))
	
	header := fmt.Sprintf("%-5s %-25s %-20s %-25s %-10s", "ID", "Nombre", "Cliente", "Ubicación", "Empresa ID")
	fmt.Println(calcTableHeaderStyle.Render(header))
	fmt.Println(strings.Repeat("─", 100))

	for _, proj := range proyectos {
		nombre := proj.Nombre
		if len(nombre) > 23 {
			nombre = nombre[:20] + "..."
		}
		cliente := proj.Cliente
		if len(cliente) > 18 {
			cliente = cliente[:15] + "..."
		}
		ubicacion := proj.Ubicacion
		if len(ubicacion) > 23 {
			ubicacion = ubicacion[:20] + "..."
		}
		row := fmt.Sprintf("%-5d %-25s %-20s %-25s %-10d", proj.ID, nombre, cliente, ubicacion, proj.EmpresaID)
		fmt.Println(calcTableRowStyle.Render(row))
	}
	fmt.Println(strings.Repeat("─", 100) + "\n")
}

func showIngenierosTable(ingenieros []calcapi.Ingeniero) {
	if len(ingenieros) == 0 {
		return
	}

	fmt.Println("\n" + calcTitleStyle.Render("📋 Ingenieros Disponibles"))
	fmt.Println(strings.Repeat("─", 80))
	
	header := fmt.Sprintf("%-5s %-30s %-25s %-15s", "ID", "Nombre", "Profesión", "CODIA")
	fmt.Println(calcTableHeaderStyle.Render(header))
	fmt.Println(strings.Repeat("─", 80))

	for _, ing := range ingenieros {
		nombre := ing.Nombre
		if len(nombre) > 28 {
			nombre = nombre[:25] + "..."
		}
		profesion := ing.Profesion
		if len(profesion) > 23 {
			profesion = profesion[:20] + "..."
		}
		row := fmt.Sprintf("%-5d %-30s %-25s %-15s", ing.ID, nombre, profesion, ing.CODIA)
		fmt.Println(calcTableRowStyle.Render(row))
	}
	fmt.Println(strings.Repeat("─", 80) + "\n")
}
