package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/osmargm1202/orgm/inputs"
	"github.com/osmargm1202/orgm/pkg/propapi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// PropCmd represents the prop command
var PropCmd = &cobra.Command{
	Use:   "prop",
	Short: "Gestión de propuestas con API",
	Long: `Comando para crear, modificar y gestionar propuestas usando la API de propuestas.

Subcomandos disponibles:
  new     Crear nueva propuesta con interfaz gráfica
  mod     Modificar propuesta existente con interfaz gráfica
  view    Ver y descargar propuestas con interfaz gráfica`,
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // Ensure token for all subcommands
        if _, err := EnsureGCloudIDToken(); err != nil {
            return fmt.Errorf("error obteniendo token: %w", err)
        }
        return nil
    },
	Run: func(cmd *cobra.Command, args []string) {
		showMainProposalMenu()
	},
}

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Crear nueva propuesta con interfaz gráfica",
	Long:  `Crea una nueva propuesta usando interfaz gráfica con yad`,
	Run: func(cmd *cobra.Command, args []string) {
		createNewProposalGUI()
	},
}


// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Instalar aplicación de escritorio",
	Long:  `Crea un archivo .desktop para acceder a la aplicación desde el menú de aplicaciones`,
	Run: func(cmd *cobra.Command, args []string) {
		installDesktopApp()
	},
}

// fyneCmd represents the fyne command
var fyneCmd = &cobra.Command{
	Use:   "fyne",
	Short: "Gestión de propuestas con interfaz Fyne",
	Long:  `Comando para crear, modificar y gestionar propuestas usando la API de propuestas con interfaz gráfica Fyne`,
	Run: func(cmd *cobra.Command, args []string) {
		showMainProposalMenuFyne()
	},
}




func init() {
	// Add subcommands to prop
	PropCmd.AddCommand(newCmd)
	PropCmd.AddCommand(installCmd)
	PropCmd.AddCommand(fyneCmd)
}

func getBaseURL() (string, error) {
	// Try viper first (CLI context)
	baseURL := viper.GetString("url.propuestas_api")
	if baseURL != "" {
		return strings.TrimSuffix(baseURL, "/"), nil
	}
	
	// Fall back to shared config helper
	return propapi.GetBaseURL()
}






func downloadSpecificProposal(client *propapi.Client, proposal propapi.Proposal) {
	// Get download directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al obtener directorio home: "+err.Error()))
		return
	}
	downloadDir := filepath.Join(homeDir, "Downloads")

	// Create download directory if it doesn't exist
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al crear directorio de descarga: "+err.Error()))
		return
	}

	fmt.Printf("%s\n", inputs.InfoStyle.Render("Descargando archivos..."))

	// Download MD file (always try since MD is always generated)
	if err := client.DownloadProposalFile(proposal.ID, "md", filepath.Join(downloadDir, proposal.ID+".md")); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al descargar MD: "+err.Error()))
	} else {
		fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ MD descargado"))
	}

    // Download HTML file: intentar aunque el API no reporte URL (404 si no existe)
		if err := client.DownloadProposalFile(proposal.ID, "html", filepath.Join(downloadDir, proposal.ID+".html")); err != nil {
        fmt.Printf("%s\n", inputs.InfoStyle.Render("HTML no disponible aún"))
		} else {
			fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ HTML descargado"))
	}

    // Download PDF file: intentar aunque el API no reporte URL (404 si no existe)
		if err := client.DownloadProposalFile(proposal.ID, "pdf", filepath.Join(downloadDir, proposal.ID+".pdf")); err != nil {
        fmt.Printf("%s\n", inputs.InfoStyle.Render("PDF no disponible aún"))
		} else {
			fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ PDF descargado"))
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("Descarga completada en: " + downloadDir))

	// Open download directory
	openDirectory(downloadDir)
}


func openDirectory(path string) {
	var cmd *exec.Cmd

	switch {
	case isCommandAvailable("xdg-open"):
		cmd = exec.Command("xdg-open", path)
	case isCommandAvailable("open"):
		cmd = exec.Command("open", path)
	case isCommandAvailable("explorer"):
		cmd = exec.Command("explorer", path)
	default:
		fmt.Printf("%s\n", inputs.InfoStyle.Render("No se pudo abrir el directorio automáticamente"))
		return
	}

	// Start the command in background without waiting
	if err := cmd.Start(); err != nil {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("No se pudo abrir el directorio automáticamente"))
	}
}

func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// showNotification shows a dunst notification
func showNotification(title, message string) {
	if isCommandAvailable("dunstify") {
		exec.Command("dunstify", "--appname=Propuestas", title, message).Run()
	} else if isCommandAvailable("notify-send") {
		exec.Command("notify-send", "--app-name=Propuestas", title, message).Run()
	}
}

func generateHTMLAndPDF(client *propapi.Client, proposalID string) {
	// Generate HTML
	fmt.Printf("%s\n", inputs.InfoStyle.Render("Generando HTML..."))
	showNotification("Generando HTML", "Iniciando generación de HTML...")
    stop := startYadProgress("Manejando Solicitud", "Conectando con la API...\\nProcesando...")
	
	htmlResponse, err := client.GenerateHTML(proposalID)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al generar HTML: "+err.Error()))
		showNotification("Error HTML", "Error al generar HTML")
        stop()
        return
	}
	
	fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ HTML generado: " + htmlResponse.HTMLURL))
	showNotification("HTML Listo", "HTML generado exitosamente")

    // Generate PDF
	fmt.Printf("%s\n", inputs.InfoStyle.Render("Generando PDF..."))
	showNotification("Generando PDF", "Iniciando generación de PDF...")
	
	pdfResponse, err := client.GeneratePDF(proposalID, "normal")
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al generar PDF: "+err.Error()))
		showNotification("Error PDF", "Error al generar PDF")
        stop()
        return
	}
	
	fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ PDF generado: " + pdfResponse.PDFURL))
	showNotification("PDF Listo", "PDF generado exitosamente")
    stop()
}

// createClient creates a propapi client with authentication
func createClient() (*propapi.Client, error) {
	baseURL, err := propapi.GetBaseURL()
	if err != nil {
		return nil, err
	}
	
	// Create auth function that uses EnsureGCloudIDToken
	authFunc := func(req *http.Request) {
		token, err := EnsureGCloudIDToken()
		if err != nil || token == "" {
			fmt.Printf("Warning: No se pudo obtener token de autenticación: %v\n", err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	return propapi.NewClient(baseURL, authFunc), nil
}

// startYadProgress launches a pulsating YAD progress dialog and returns a stop function.
func startYadProgress(title, text string) func() {
    cmd := exec.Command(
        "yad",
        "--progress",
        "--title="+title,
        "--text="+text,
        "--progress-text=",
        "--pulsate",
        "--width=400",
        "--no-buttons",
    )
    // Detach stdio so it doesn't block
    cmd.Stdout = nil
    cmd.Stderr = nil
    _ = cmd.Start()
    return func() {
        if cmd.Process != nil {
            _ = cmd.Process.Kill()
        }
    }
}

// showMainProposalMenu shows the main unified proposal menu
func showMainProposalMenu() {
	for {
		// Show main menu as a list like the action menu
		cmd := exec.Command("yad",
			"--list",
			"--title=📋 Gestor de Propuestas",
			"--text=Selecciona una opción:",
			"--column=Opción",
			"--width=400",
			"--height=200",
			"--print-column=1",
			"--single-click",
			"--separator=|")

		// Add menu items
		cmd.Args = append(cmd.Args, "🆕 Nueva Propuesta")
		cmd.Args = append(cmd.Args, "📁 Propuesta Existente")

		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error ejecutando yad: "+err.Error()))
			return
		}

		result := strings.TrimSpace(string(output))
		result = strings.TrimSuffix(result, "|")
		if result == "" {
			return
		}

		// Handle selected action
		switch result {
		case "🆕 Nueva Propuesta":
			createNewProposalFlow()
		case "📁 Propuesta Existente":
			showExistingProposalFlow()
		}
	}
}

// createNewProposalFlow handles the complete flow for creating new proposals
func createNewProposalFlow() {
	// Show form to create new proposal
	cmd := exec.Command("yad",
		"--form",
		"--title=🆕 Nueva Propuesta",
		"--text=Completa los datos de la nueva propuesta:",
		`--field=Título:TXT`, "",
		`--field=Subtítulo:TXT`, "",
		`--field=Prompt:TXT`, "",
		"--width=600",
		"--height=400")

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error ejecutando yad: "+err.Error()))
		return
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return
	}

	// Parse form result (yad form returns values separated by |)
	parts := strings.Split(result, "|")
	if len(parts) < 3 {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Formato de respuesta inválido"))
		return
	}

	title := strings.TrimSpace(parts[0])
	subtitle := strings.TrimSpace(parts[1])
	prompt := strings.TrimSpace(parts[2])

	if prompt == "" {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("El prompt no puede estar vacío"))
		return
	}

	// Create request
	request := propapi.TextGenerationRequest{
		Title:    title,
		Subtitle: subtitle,
		Prompt:   prompt,
		Model:    "gpt-5-chat-latest",
	}

	// Show generation menu
	showGenerationMenu(request)
}

// showGenerationMenu shows options for generating documents after creating proposal
func showGenerationMenu(request propapi.TextGenerationRequest) {
	for {
		cmd := exec.Command("yad",
			"--form",
			"--title=📄 Generar Documentos",
			"--text=¿Qué documento quieres generar?",
			"--field=Acción:CB", "📝 Solo Texto (MD)!🌐 Generar HTML!📄 Generar PDF!🏠 Volver al Menú Principal",
			"--width=400",
			"--height=250")

		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error ejecutando yad: "+err.Error()))
			return
		}

		result := strings.TrimSpace(string(output))
		if result == "" {
			return
		}

		action := strings.TrimSpace(strings.Split(result, "|")[0])
		if action == "" {
			return
		}

		switch action {
        case "📝 Solo Texto (MD)":
            // Create proposal and download MD, open folder
            client, err := createClient()
            if err != nil {
                fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creando cliente: "+err.Error()))
                continue
            }
            response, err := client.CreateProposal(request)
            if err != nil {
                fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creando propuesta: "+err.Error()))
                continue
            }
            mdPath := getDownloadPath(response.ID + ".md")
            _ = client.DownloadProposalFile(response.ID, "md", mdPath)
            homeDir, _ := os.UserHomeDir()
            openDirectory(filepath.Join(homeDir, "Downloads"))
        case "🌐 Generar HTML":
            // Create proposal and generate HTML, then PDF too per request
            client, err := createClient()
            if err != nil {
                fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creando cliente: "+err.Error()))
                continue
            }
            response, err := client.CreateProposal(request)
            if err != nil {
                fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creando propuesta: "+err.Error()))
                continue
            }
            generateHTMLAndPDF(client, response.ID)
        case "📄 Generar PDF":
            // Create proposal and generate HTML+PDF immediately
            client, err := createClient()
            if err != nil {
                fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creando cliente: "+err.Error()))
                continue
            }
            response, err := client.CreateProposal(request)
            if err != nil {
                fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creando propuesta: "+err.Error()))
                continue
            }
            generateHTMLAndPDF(client, response.ID)
		case "🏠 Volver al Menú Principal":
			return
		}
	}
}

// showExistingProposalFlow handles the complete flow for existing proposals
func showExistingProposalFlow() {
    // Show loading while fetching proposals
    stop := startYadProgress("Manejando Solicitud", "Conectando con la API...\\nProcesando...")
    client, err := createClient()
    if err != nil {
        fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creando cliente: "+err.Error()))
        return
    }
    proposals, err := client.GetProposals()
    stop()
    if err != nil {
        fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error obteniendo propuestas: "+err.Error()))
		return
	}

    if len(proposals) == 0 {
        fmt.Printf("%s\n", inputs.InfoStyle.Render("No hay propuestas disponibles"))
        return
    }

    // Show list with all proposals directly
    cmd := exec.Command("yad", 
        "--list",
        "--title=📁 Propuestas Existentes",
        "--text=Selecciona una propuesta:",
        "--column=ID",
        "--column=Título",
        "--column=Subtítulo", 
        "--column=Creada",
        "--width=900",
        "--height=480",
        "--search-column=2", // quick search on title
        "--print-column=1",   // Return only ID
        "--single-click",
        "--separator=|")

    // Add proposal items as arguments
    for _, prop := range proposals {
        cmd.Args = append(cmd.Args, prop.ID)
        cmd.Args = append(cmd.Args, prop.Title)
        cmd.Args = append(cmd.Args, prop.Subtitle)
        cmd.Args = append(cmd.Args, prop.CreatedAt.Format("2006-01-02 15:04"))
    }

    output, err := cmd.Output()
    if err != nil {
        fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error ejecutando yad: "+err.Error()))
        return
    }

    selectedID := strings.TrimSpace(string(output))
    selectedID = strings.TrimSuffix(selectedID, "|")
    if selectedID == "" {
        return
    }

    // Find selected proposal
    var selectedProposal *propapi.Proposal
    for _, prop := range proposals {
        if prop.ID == selectedID {
            selectedProposal = &prop
            break
        }
    }

    if selectedProposal == nil {
        fmt.Printf("%s\n", inputs.ErrorStyle.Render("Propuesta no encontrada"))
        return
    }

    // Show proposal management menu (stays in loop until user exits)
    showProposalManagementLoop(client, *selectedProposal)
}

// showProposalManagementLoop shows the management menu for a specific proposal in a loop
func showProposalManagementLoop(client *propapi.Client, proposal propapi.Proposal) {
	for {
		// Create menu options based on available documents
        menuItems := []string{
            "📝 Ver propuesta (MD)",
            "🛠️ Regenerar (título/subtítulo/prompt)",
            "✍️ Cambiar solo título/subtítulo",
        }
		
		// Add conditional buttons based on document availability
        if proposal.HTMLURL != "" {
            menuItems = append(menuItems, "🌐 Ver HTML")
            menuItems = append(menuItems, "🔁 Regenerar HTML")
		} else {
			menuItems = append(menuItems, "🌐 Generar HTML")
		}
		
        if proposal.PDFURL != "" {
            menuItems = append(menuItems, "📄 Ver PDF")
            menuItems = append(menuItems, "🔁 Regenerar PDF")
		} else {
			menuItems = append(menuItems, "📄 Generar PDF")
		}
		
		menuItems = append(menuItems, "✏️ Modificar propuesta")
		menuItems = append(menuItems, "📥 Descargar archivos")
		menuItems = append(menuItems, "🏠 Volver al Menú Principal")

		// Show yad menu dialog
		cmd := exec.Command("yad", 
			"--list",
			"--title=📋 Gestionar: "+proposal.Title,
			"--text=Selecciona una acción:",
			"--column=Acción",
			"--width=400",
			"--height=350",
			"--print-column=1",
			"--single-click",
			"--separator=|")

		// Add menu items
		cmd.Args = append(cmd.Args, menuItems...)

		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error ejecutando yad: "+err.Error()))
			return
		}

        selectedAction := strings.TrimSpace(string(output))
        // Remove trailing separator if present
        selectedAction = strings.TrimSuffix(selectedAction, "|")
		if selectedAction == "" {
			return
		}

		// Handle selected action
		switch selectedAction {
        case "📝 Ver propuesta (MD)":
            // Descargar MD y abrir carpeta de descargas
            mdPath := getDownloadPath(proposal.ID + ".md")
            if err := client.DownloadProposalFile(proposal.ID, "md", mdPath); err != nil {
                exec.Command("yad", "--error", "--title=Descarga MD", "--text=MD no disponible aún").Run()
            } else {
                exec.Command("yad", "--info", "--title=Descarga MD", "--text=MD descargado en carpeta de Descargas").Run()
            }
            // abrir carpeta
            homeDir, _ := os.UserHomeDir()
            openDirectory(filepath.Join(homeDir, "Downloads"))
        case "🌐 Ver HTML":
            // Descargar HTML y abrir carpeta
            downloadPath := getDownloadPath(proposal.ID + ".html")
            if err := client.DownloadProposalFile(proposal.ID, "html", downloadPath); err != nil {
                exec.Command("yad", "--error", "--title=Descarga HTML", "--text=HTML no disponible aún").Run()
            } else {
                exec.Command("yad", "--info", "--title=Descarga HTML", "--text=HTML descargado en carpeta de Descargas").Run()
            }
            homeDir, _ := os.UserHomeDir()
            openDirectory(filepath.Join(homeDir, "Downloads"))
		case "🌐 Generar HTML":
			generateHTMLGUI(client, proposal.ID)
		case "🔁 Regenerar HTML":
			// Fuerza nueva generación de HTML según API: POST /generate-html con proposal_id y model por defecto
			generateHTMLGUI(client, proposal.ID)
			// Después de regenerar HTML, regenerar PDF automáticamente
			generatePDFGUI(client, proposal.ID)
		case "🔁 Regenerar PDF":
			// Volver a generar PDF (llama a generatePDFGUI que invoca POST /generate-pdf)
			generatePDFGUI(client, proposal.ID)
        case "📄 Ver PDF":
            // Descargar PDF y abrir carpeta
            downloadPath := getDownloadPath(proposal.ID + ".pdf")
            if err := client.DownloadProposalFile(proposal.ID, "pdf", downloadPath); err != nil {
                exec.Command("yad", "--error", "--title=Descarga PDF", "--text=PDF no disponible aún").Run()
            } else {
                exec.Command("yad", "--info", "--title=Descarga PDF", "--text=PDF descargado en carpeta de Descargas").Run()
            }
            homeDir, _ := os.UserHomeDir()
            openDirectory(filepath.Join(homeDir, "Downloads"))
		case "📄 Generar PDF":
			generatePDFGUI(client, proposal.ID)
		case "✏️ Modificar propuesta":
			modifyProposalGUI(client, proposal)
		case "📥 Descargar archivos":
			downloadSpecificProposal(client, proposal)
		case "🏠 Volver al Menú Principal":
			return
        case "🛠️ Regenerar (título/subtítulo/prompt)":
            regenerateProposalGUI(client, &proposal)
        case "✍️ Cambiar solo título/subtítulo":
            updateTitleSubtitleGUI(client, &proposal)
		}
	}
}

// regenerateProposalGUI: POST /proposal/{id}/regenerate with title/subtitle/prompt
func regenerateProposalGUI(client *propapi.Client, proposal *propapi.Proposal) {
    // Form with current values
    cmd := exec.Command("yad",
        "--form",
        "--title=🛠️ Regenerar Propuesta: "+proposal.Title,
        "--text=Edita los campos para regenerar el contenido (MD se reemplaza)",
        `--field=Título:TXT`, proposal.Title,
        `--field=Subtítulo:TXT`, proposal.Subtitle,
        `--field=Prompt:TXT`, proposal.Prompt,
        "--width=700",
        "--height=420")

    output, err := cmd.Output()
    if err != nil {
        fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error ejecutando yad: "+err.Error()))
        return
    }
    res := strings.TrimSpace(string(output))
    if res == "" { return }
    parts := strings.Split(res, "|")
    if len(parts) < 3 { exec.Command("yad","--error","--text=Entrada inválida").Run(); return }
    newTitle := strings.TrimSpace(parts[0])
    newSubtitle := strings.TrimSpace(parts[1])
    newPrompt := strings.TrimSpace(parts[2])

    err = client.RegenerateProposal(proposal.ID, newTitle, newSubtitle, newPrompt)
    if err != nil { 
        exec.Command("yad","--error","--text=Fallo al regenerar: "+err.Error()).Run(); 
        return 
    }

    // Update local proposal state and clear HTML/PDF
    proposal.Title = newTitle
    proposal.Subtitle = newSubtitle
    proposal.Prompt = newPrompt
    proposal.HTMLURL = ""
    proposal.PDFURL = ""

    exec.Command("yad","--info","--text=Texto regenerado. Generando HTML y PDF...").Run()
    generateHTMLAndPDF(client, proposal.ID)
}

// updateTitleSubtitleGUI: PATCH /proposal/{id}/title-subtitle
func updateTitleSubtitleGUI(client *propapi.Client, proposal *propapi.Proposal) {
    cmd := exec.Command("yad",
        "--form",
        "--title=✍️ Actualizar Título/Subtítulo: "+proposal.Title,
        `--field=Título:TXT`, proposal.Title,
        `--field=Subtítulo:TXT`, proposal.Subtitle,
        "--width=600",
        "--height=260")
    output, err := cmd.Output()
    if err != nil { fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error ejecutando yad: "+err.Error())); return }
    res := strings.TrimSpace(string(output))
    if res == "" { return }
    parts := strings.Split(res, "|")
    if len(parts) < 2 { exec.Command("yad","--error","--text=Entrada inválida").Run(); return }
    newTitle := strings.TrimSpace(parts[0])
    newSubtitle := strings.TrimSpace(parts[1])

    err = client.UpdateTitleSubtitle(proposal.ID, newTitle, newSubtitle)
    if err != nil { 
        exec.Command("yad","--error","--text=Fallo al actualizar: "+err.Error()).Run(); 
        return 
    }

    proposal.Title = newTitle
    proposal.Subtitle = newSubtitle
    exec.Command("yad","--info","--text=Título/Subtítulo actualizados. Si deseas verlos en HTML/PDF, vuelve a generarlos.").Run()
}


// modifyProposalGUI shows a dialog to modify a proposal
func modifyProposalGUI(client *propapi.Client, proposal propapi.Proposal) {
	// Show yad text entry dialog for prompt
	cmd := exec.Command("yad",
		"--form",
		"--title=Modificar Propuesta: "+proposal.Title,
		"--text=Ingresa el nuevo prompt:",
		`--field=Título:TXT`, proposal.Title,
		`--field=Subtítulo:TXT`, proposal.Subtitle,
		`--field=Prompt:TXT`, proposal.Title+" modificada",
		"--width=600",
		"--height=400")

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error ejecutando yad: "+err.Error()))
		return
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return
	}

	// Parse form result (yad form returns values separated by |)
	parts := strings.Split(result, "|")
	if len(parts) < 3 {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Formato de respuesta inválido"))
		return
	}

	title := strings.TrimSpace(parts[0])
	subtitle := strings.TrimSpace(parts[1])
	prompt := strings.TrimSpace(parts[2])

	if prompt == "" {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("El prompt no puede estar vacío"))
		return
	}

	// Send modification request
	fmt.Printf("%s\n", inputs.InfoStyle.Render("Modificando propuesta..."))

	response, err := client.ModifyProposal(proposal.ID, title, subtitle, prompt)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al modificar propuesta: "+err.Error()))
		return
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("¡Propuesta modificada exitosamente!"))
	fmt.Printf("ID: %s\n", response.ID)

// Generar HTML y PDF automáticamente tras modificar
generateHTMLAndPDF(client, proposal.ID)

// Show success dialog
exec.Command("yad", "--info", "--title=Éxito", "--text=Propuesta modificada y documentos generados").Run()
}

// showProposalContent shows the proposal content in a yad dialog
func showProposalContent(client *propapi.Client, proposal propapi.Proposal) {
	// Download MD content
	mdPath := filepath.Join(os.TempDir(), proposal.ID+".md")
	if err := client.DownloadProposalFile(proposal.ID, "md", mdPath); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error descargando contenido: "+err.Error()))
		return
	}
	defer os.Remove(mdPath)

	// Read content
	content, err := os.ReadFile(mdPath)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error leyendo contenido: "+err.Error()))
		return
	}

	// Create temporary file with content
	tempFile, err := os.CreateTemp("", "proposal_content_*.md")
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creando archivo temporal: "+err.Error()))
		return
	}
	defer os.Remove(tempFile.Name())

	tempFile.Write(content)
	tempFile.Close()

	// Show content in yad text dialog with buttons
	cmd := exec.Command("yad",
		"--text-info",
		"--title=Propuesta: "+proposal.Title,
		"--filename="+tempFile.Name(),
		"--width=800",
		"--height=600",
		"--button=Generar HTML:1",
		"--button=Generar PDF:2",
		"--button=Cerrar:0")

	_, err = cmd.Output()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error ejecutando yad: "+err.Error()))
		return
	}

	exitCode := cmd.ProcessState.ExitCode()
	switch exitCode {
	case 1: // Generate HTML
		generateHTMLGUI(client, proposal.ID)
	case 2: // Generate PDF
		generatePDFGUI(client, proposal.ID)
	}
}

// generateHTMLGUI generates HTML and shows success message
func generateHTMLGUI(client *propapi.Client, proposalID string) {
	fmt.Printf("%s\n", inputs.InfoStyle.Render("Generando HTML..."))
	showNotification("Generando HTML", "Iniciando generación de HTML...")
	stop := startYadProgress("Manejando Solicitud", "Conectando con la API...\\nProcesando...")

	_, err := client.GenerateHTML(proposalID)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al generar HTML: "+err.Error()))
		showNotification("Error HTML", "Error al generar HTML")
		stop()
		return
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ HTML generado exitosamente"))
	showNotification("HTML Listo", "HTML generado exitosamente")

	// Download HTML file
	client.DownloadProposalFile(proposalID, "html", getDownloadPath(proposalID+".html"))

	// Show success dialog
	exec.Command("yad", "--info", "--title=Éxito", "--text=HTML generado y descargado exitosamente").Run()
	stop()
}

// generatePDFGUI generates PDF and opens it
func generatePDFGUI(client *propapi.Client, proposalID string) {
	// Show dialog to select PDF mode
	cmd := exec.Command("yad",
		"--form",
		"--title=Generar PDF",
		"--text=Selecciona el modo de impresión del PDF:",
		"--field=Modo:CB", "normal!dapec!oscuro",
		"--width=400",
		"--height=200")

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error ejecutando yad: "+err.Error()))
		return
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return
	}

	// Parse mode selection
	mode := strings.TrimSpace(strings.Split(result, "|")[0])
	if mode == "" {
		mode = "normal"
	}

    fmt.Printf("%s\n", inputs.InfoStyle.Render("Generando PDF en modo: "+mode+"..."))
    showNotification("Generando PDF", "Iniciando generación de PDF en modo "+mode+"...")
    stop := startYadProgress("Manejando Solicitud", "Generando PDF...")

	pdfResponse, err := client.GeneratePDF(proposalID, mode)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al generar PDF: "+err.Error()))
		showNotification("Error PDF", "Error al generar PDF")
		stop()
		return
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ PDF generado: " + pdfResponse.PDFURL))
	showNotification("PDF Listo", "PDF generado exitosamente")

	// Download PDF file
	filepath := getDownloadPath(proposalID + ".pdf")
	if err := client.DownloadProposalFile(proposalID, "pdf", filepath); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al descargar PDF: "+err.Error()))
		showNotification("Error PDF", "Error al descargar PDF")
		return
	}

	// Open PDF file
    openFile(filepath)
    stop()
}

// createNewProposalGUI creates a new proposal using GUI
func createNewProposalGUI() {
	// Show yad form dialog for new proposal


	cmd := exec.Command("yad",
    "--form",
    "--title=Nueva Propuesta",
    "--text=Ingresa los datos de la nueva propuesta:",
    `--field=Título:TXT`, `Propuesta de Servicios`,
    "--field=Subtítulo:TXT",
    "--field=Prompt:TXT",
    "--width=600",
    "--height=400")

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error ejecutando yad: "+err.Error()))
		return
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return
	}

	// Parse form result
	parts := strings.Split(result, "|")
	if len(parts) < 3 {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Formato de respuesta inválido"))
		return
	}

	title := strings.TrimSpace(parts[0])
	subtitle := strings.TrimSpace(parts[1])
	prompt := strings.TrimSpace(parts[2])

	if prompt == "" || prompt == "Escribe aquí tu prompt..." {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("El prompt no puede estar vacío"))
		return
	}

	// Create request
	request := propapi.TextGenerationRequest{
		Title:    title,
		Subtitle: subtitle,
		Prompt:   prompt,
		Model:    "gpt-5-chat-latest",
	}

	// Send request
	fmt.Printf("%s\n", inputs.InfoStyle.Render("Creando propuesta..."))
	stop := startYadProgress("Manejando Solicitud", "Conectando con la API...\\nProcesando...")

	// Create client
	client, err := createClient()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creando cliente: "+err.Error()))
		stop()
		return
	}

	response, err := client.CreateProposal(request)

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("¡Propuesta creada exitosamente!"))
	fmt.Printf("ID: %s\n", response.ID)

	// Download MD file
	filepath := getDownloadPath(response.ID + ".md")
	if err := client.DownloadProposalFile(response.ID, "md", filepath); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al descargar MD: "+err.Error()))
		return
	}

	// Show success dialog with options
	cmd = exec.Command("yad",
		"--question",
		"--title=Propuesta Creada",
		"--text=Propuesta creada exitosamente.\n¿Qué deseas hacer?",
		"--button=Ver Propuesta:1",
		"--button=Generar HTML:2",
		"--button=Generar PDF:3",
		"--button=Cerrar:0")

	cmd.Run()
	exitCode := cmd.ProcessState.ExitCode()

	switch exitCode {
	case 1: // View proposal
		// Create a temporary proposal object for viewing
		tempProposal := propapi.Proposal{
			ID:        response.ID,
			Title:     title,
			Subtitle:  subtitle,
			CreatedAt: response.CreatedAt,
		}
		showProposalContent(client, tempProposal)
	case 2: // Generate HTML
		generateHTMLGUI(client, response.ID)
	case 3: // Generate PDF
		generatePDFGUI(client, response.ID)
	}
}

// Helper functions
func getDownloadPath(filename string) string {
	homeDir, _ := os.UserHomeDir()
	downloadDir := filepath.Join(homeDir, "Downloads")
	os.MkdirAll(downloadDir, 0755)
	return filepath.Join(downloadDir, filename)
}

func openFile(filepath string) {
	var cmd *exec.Cmd
	switch {
	case isCommandAvailable("xdg-open"):
		cmd = exec.Command("xdg-open", filepath)
	case isCommandAvailable("open"):
		cmd = exec.Command("open", filepath)
	case isCommandAvailable("explorer"):
		cmd = exec.Command("explorer", filepath)
	default:
		fmt.Printf("%s\n", inputs.InfoStyle.Render("No se pudo abrir el archivo automáticamente"))
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("No se pudo abrir el archivo automáticamente"))
	}
}


// installDesktopApp creates a .desktop file for the application
func installDesktopApp() {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error obteniendo ruta del ejecutable: "+err.Error()))
		return
	}

	// Get user home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error obteniendo directorio home: "+err.Error()))
		return
	}

	// Create .desktop file content
	desktopContent := fmt.Sprintf(`[Desktop Entry]
Version=1.0
Type=Application
Name=📋 Gestor de Propuestas
Comment=Gestiona propuestas comerciales con interfaz gráfica
Exec=%s prop
Icon=applications-office
Terminal=false
Categories=Office;Documentation;
Keywords=propuestas;documentos;pdf;html;
StartupNotify=true
`, execPath)

	// Create applications directory if it doesn't exist
	applicationsDir := filepath.Join(homeDir, ".local", "share", "applications")
	if err := os.MkdirAll(applicationsDir, 0755); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creando directorio de aplicaciones: "+err.Error()))
		return
	}

	// Write .desktop file
	desktopPath := filepath.Join(applicationsDir, "propuestas.desktop")
	if err := os.WriteFile(desktopPath, []byte(desktopContent), 0755); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error escribiendo archivo .desktop: "+err.Error()))
		return
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("✅ Aplicación instalada exitosamente!"))
	fmt.Printf("%s\n", inputs.InfoStyle.Render("📁 Archivo creado en: "+desktopPath))
	fmt.Printf("%s\n", inputs.InfoStyle.Render("🔍 Busca 'Propuestas' en el menú de aplicaciones"))
	fmt.Printf("%s\n", inputs.InfoStyle.Render("💡 También puedes ejecutar: "+execPath+" prop"))
}

// ==============================================
// FUNCIONES FYNE GUI
// ==============================================

// showMainProposalMenuFyne shows the main proposal menu using Fyne
func showMainProposalMenuFyne() {
	myApp := app.NewWithID("orgm.propuestas")

	myWindow := myApp.NewWindow("Gestor de Propuestas")
	myWindow.Resize(fyne.NewSize(1400, 900))
	myWindow.CenterOnScreen()

	// Create main interface content
	content := createMainInterfaceContent(myApp, myWindow)
	myWindow.SetContent(content)

	myWindow.ShowAndRun()
}

// ProposalManager holds the state for the proposal manager
type ProposalManager struct {
	proposals           []propapi.Proposal
	filteredProposals   []propapi.Proposal
	selectedProposal    *propapi.Proposal
	proposalsList       *widget.List
	selectedLabel       *widget.Label
	searchEntry         *widget.Entry
	client              *propapi.Client
	window              fyne.Window
	app                 fyne.App
}

// Global variable to store the current proposal manager
var currentProposalManager *ProposalManager

// createMainInterfaceContent creates the unified main interface
func createMainInterfaceContent(app fyne.App, window fyne.Window) *fyne.Container {
	// Title
	title := widget.NewLabel("Gestor de Propuestas")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	// Search/filter bar
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Buscar por título o subtítulo...")

	// Proposals list
	proposalsList := widget.NewList(
		func() int {
			if currentProposalManager != nil {
				return len(currentProposalManager.filteredProposals)
			}
			return 0
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("ID"),
				widget.NewLabel("Título"),
				widget.NewLabel("Subtítulo"),
				widget.NewLabel("Creada"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if currentProposalManager != nil && id < len(currentProposalManager.filteredProposals) {
				proposal := currentProposalManager.filteredProposals[id]
				container := obj.(*fyne.Container)
				container.Objects[0].(*widget.Label).SetText(proposal.ID)
				container.Objects[1].(*widget.Label).SetText(proposal.Title)
				container.Objects[2].(*widget.Label).SetText(proposal.Subtitle)
				container.Objects[3].(*widget.Label).SetText(proposal.CreatedAt.Format("2006-01-02 15:04"))
			}
		},
	)

	proposalsList.Resize(fyne.NewSize(200, 500))

	// List selection handler
	proposalsList.OnSelected = func(id widget.ListItemID) {
		if currentProposalManager != nil && id < len(currentProposalManager.filteredProposals) {
			proposal := currentProposalManager.filteredProposals[id]
			currentProposalManager.selectedProposal = &proposal
			currentProposalManager.selectedLabel.SetText(fmt.Sprintf("Seleccionada: %s\n%s", proposal.Title, proposal.Subtitle))
		}
	}

	// Selected proposal info
	selectedProposalLabel := widget.NewLabel("Ninguna propuesta seleccionada")
	selectedProposalLabel.Alignment = fyne.TextAlignCenter

	// Create client
	client, err := createClient()
	if err != nil {
		// Show error dialog
		dialog.ShowError(fmt.Errorf("Error creando cliente: %v", err), window)
		return container.NewVBox(widget.NewLabel("Error: No se pudo conectar con la API"))
	}

	// Create proposal manager
	manager := &ProposalManager{
		proposals:         []propapi.Proposal{},
		filteredProposals: []propapi.Proposal{},
		selectedProposal:  nil,
		proposalsList:     proposalsList,
		selectedLabel:     selectedProposalLabel,
		searchEntry:       searchEntry,
		client:            client,
		window:            window,
		app:               app,
	}

	// Store manager in global variable for list callbacks
	currentProposalManager = manager

	// Action buttons
	buttonsContainer := manager.createActionButtons()

	// Layout - New proposal button first, then list, then buttons
	// New proposal button at the top
	newProposalBtn := widget.NewButton("🆕 Nueva Propuesta", func() {
		createNewProposalFlowFyne(app, window)
	})
	newProposalBtn.Resize(fyne.NewSize(200, 40))

	// List panel
	listPanel := container.NewVBox(
		widget.NewLabel("Propuestas Existentes:"),
		searchEntry,
		proposalsList,
	)

	listPanel.Resize(fyne.NewSize(200, 500))


	// Buttons panel - 2 rows of 5 buttons
	buttonsPanel := container.NewVBox(
		selectedProposalLabel,
		widget.NewSeparator(),
		buttonsContainer,
	)

	// Vertical split layout
	split := container.NewVSplit(listPanel, buttonsPanel)
	split.Offset = 0.7 // 70% for list, 30% for buttons

	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		container.NewCenter(newProposalBtn),
		widget.NewSeparator(),
		split,
	)

	// Load proposals data
	manager.loadProposalsData()

	return container.NewPadded(content)
}

// createActionButtons creates all action buttons in 2 rows of 5
func (pm *ProposalManager) createActionButtons() *fyne.Container {
	// Action buttons
	viewMDBtn := widget.NewButton("📝 Ver MD", func() {
		if pm.selectedProposal == nil {
			dialog.ShowInformation("Seleccionar Propuesta", "Por favor selecciona una propuesta primero", pm.window)
			return
		}
		mdPath := getDownloadPath(pm.selectedProposal.ID + ".md")
		if err := pm.client.DownloadProposalFile(pm.selectedProposal.ID, "md", mdPath); err != nil {
			dialog.ShowError(fmt.Errorf("MD no disponible aún"), pm.window)
		} else {
			dialog.ShowInformation("Descarga MD", "MD descargado en carpeta de Descargas", pm.window)
		}
		homeDir, _ := os.UserHomeDir()
		openDirectory(filepath.Join(homeDir, "Downloads"))
	})

	regenerateBtn := widget.NewButton("🛠️ Regenerar", func() {
		if pm.selectedProposal == nil {
			dialog.ShowInformation("Seleccionar Propuesta", "Por favor selecciona una propuesta primero", pm.window)
			return
		}
		regenerateProposalFyne(pm.app, pm.window, pm.client, pm.selectedProposal)
	})

	updateTitleBtn := widget.NewButton("✍️ Actualizar Título", func() {
		if pm.selectedProposal == nil {
			dialog.ShowInformation("Seleccionar Propuesta", "Por favor selecciona una propuesta primero", pm.window)
			return
		}
		updateTitleSubtitleFyne(pm.app, pm.window, pm.client, pm.selectedProposal)
	})

	viewHTMLBtn := widget.NewButton("🌐 Ver HTML", func() {
		if pm.selectedProposal == nil {
			dialog.ShowInformation("Seleccionar Propuesta", "Por favor selecciona una propuesta primero", pm.window)
			return
		}
		downloadPath := getDownloadPath(pm.selectedProposal.ID + ".html")
		if err := pm.client.DownloadProposalFile(pm.selectedProposal.ID, "html", downloadPath); err != nil {
			dialog.ShowError(fmt.Errorf("HTML no disponible aún"), pm.window)
		} else {
			dialog.ShowInformation("Descarga HTML", "HTML descargado en carpeta de Descargas", pm.window)
		}
		homeDir, _ := os.UserHomeDir()
		openDirectory(filepath.Join(homeDir, "Downloads"))
	})

	generateHTMLBtn := widget.NewButton("🌐 Generar HTML", func() {
		if pm.selectedProposal == nil {
			dialog.ShowInformation("Seleccionar Propuesta", "Por favor selecciona una propuesta primero", pm.window)
			return
		}
		generateHTMLFyne(pm.app, pm.window, pm.client, pm.selectedProposal.ID)
	})

	viewPDFBtn := widget.NewButton("📄 Ver PDF", func() {
		if pm.selectedProposal == nil {
			dialog.ShowInformation("Seleccionar Propuesta", "Por favor selecciona una propuesta primero", pm.window)
			return
		}
		downloadPath := getDownloadPath(pm.selectedProposal.ID + ".pdf")
		if err := pm.client.DownloadProposalFile(pm.selectedProposal.ID, "pdf", downloadPath); err != nil {
			dialog.ShowError(fmt.Errorf("PDF no disponible aún"), pm.window)
		} else {
			dialog.ShowInformation("Descarga PDF", "PDF descargado en carpeta de Descargas", pm.window)
		}
		homeDir, _ := os.UserHomeDir()
		openDirectory(filepath.Join(homeDir, "Downloads"))
	})

	generatePDFBtn := widget.NewButton("📄 Generar PDF", func() {
		if pm.selectedProposal == nil {
			dialog.ShowInformation("Seleccionar Propuesta", "Por favor selecciona una propuesta primero", pm.window)
			return
		}
		generatePDFFyne(pm.app, pm.window, pm.client, pm.selectedProposal.ID)
	})

	modifyBtn := widget.NewButton("✏️ Modificar", func() {
		if pm.selectedProposal == nil {
			dialog.ShowInformation("Seleccionar Propuesta", "Por favor selecciona una propuesta primero", pm.window)
			return
		}
		modifyProposalFyne(pm.app, pm.window, pm.client, *pm.selectedProposal)
	})

	downloadAllBtn := widget.NewButton("📥 Descargar Todo", func() {
		if pm.selectedProposal == nil {
			dialog.ShowInformation("Seleccionar Propuesta", "Por favor selecciona una propuesta primero", pm.window)
			return
		}
		downloadSpecificProposal(pm.client, *pm.selectedProposal)
		dialog.ShowInformation("Descarga", "Archivos descargados en carpeta de Descargas", pm.window)
	})

	refreshBtn := widget.NewButton("🔄 Actualizar Lista", func() {
		pm.loadProposalsData()
	})

	// First row of 5 buttons
	row1 := container.NewHBox(
		viewMDBtn,
		regenerateBtn,
		updateTitleBtn,
		viewHTMLBtn,
		generateHTMLBtn,
	)

	// Second row of 5 buttons
	row2 := container.NewHBox(
		viewPDFBtn,
		generatePDFBtn,
		modifyBtn,
		downloadAllBtn,
		refreshBtn,
	)

	return container.NewVBox(
		widget.NewLabel("Acciones con propuesta seleccionada:"),
		row1,
		row2,
	)
}

// loadProposalsData loads proposals and sets up the list
func (pm *ProposalManager) loadProposalsData() {
	// Show loading
	pm.selectedLabel.SetText("Cargando propuestas...")

	go func() {
		proposals, err := pm.client.GetProposals()
		if err != nil {
			pm.selectedLabel.SetText("Error cargando propuestas: " + err.Error())
			return
		}

		pm.proposals = proposals
		pm.filteredProposals = proposals

		if len(proposals) == 0 {
			pm.selectedLabel.SetText("No hay propuestas disponibles")
			pm.proposalsList.Refresh()
			return
		}

		// Debug: Log the number of proposals
		fmt.Printf("DEBUG: Cargadas %d propuestas\n", len(proposals))
		for i, prop := range proposals {
			fmt.Printf("DEBUG: Propuesta %d: ID=%s, Title=%s\n", i, prop.ID, prop.Title)
		}

		// Set up search filter
		pm.searchEntry.OnChanged = func(text string) {
			pm.filteredProposals = pm.filterProposals(text)
			pm.proposalsList.Refresh()
		}

		pm.proposalsList.Refresh()
		pm.selectedLabel.SetText(fmt.Sprintf("Cargadas %d propuestas. Selecciona una para ver acciones.", len(proposals)))
	}()
}

// filterProposals filters proposals based on search text
func (pm *ProposalManager) filterProposals(searchText string) []propapi.Proposal {
	if searchText == "" {
		return pm.proposals
	}
	
	searchText = strings.ToLower(searchText)
	var filtered []propapi.Proposal
	
	for _, proposal := range pm.proposals {
		if strings.Contains(strings.ToLower(proposal.Title), searchText) ||
		   strings.Contains(strings.ToLower(proposal.Subtitle), searchText) {
			filtered = append(filtered, proposal)
		}
	}
	
	return filtered
}

// createNewProposalFlowFyne handles the complete flow for creating new proposals with Fyne
func createNewProposalFlowFyne(app fyne.App, parent fyne.Window) {
	// Create new window for form
	formWindow := app.NewWindow("Nueva Propuesta")
	formWindow.Resize(fyne.NewSize(600, 500))
	formWindow.CenterOnScreen()

	// Form fields
	titleEntry := widget.NewEntry()
	titleEntry.SetText("Propuesta de Servicios")
	titleEntry.SetPlaceHolder("Ingresa el título")

	subtitleEntry := widget.NewEntry()
	subtitleEntry.SetPlaceHolder("Ingresa el subtítulo")

	promptEntry := widget.NewMultiLineEntry()
	promptEntry.SetPlaceHolder("Escribe aquí tu prompt...")
	promptEntry.Resize(fyne.NewSize(0, 200))

	// Form layout
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Título:", Widget: titleEntry},
			{Text: "Subtítulo:", Widget: subtitleEntry},
			{Text: "Prompt:", Widget: promptEntry},
		},
	}

	// Buttons
	createBtn := widget.NewButton("Crear Propuesta", func() {
		title := strings.TrimSpace(titleEntry.Text)
		subtitle := strings.TrimSpace(subtitleEntry.Text)
		prompt := strings.TrimSpace(promptEntry.Text)

		if prompt == "" {
			dialog.ShowError(fmt.Errorf("el prompt no puede estar vacío"), formWindow)
			return
		}

		// Create request
		request := propapi.TextGenerationRequest{
			Title:    title,
			Subtitle: subtitle,
			Prompt:   prompt,
			Model:    "gpt-5-chat-latest",
		}

		// Show generation menu
		showGenerationMenuFyne(app, formWindow, request)
	})

	cancelBtn := widget.NewButton("Cancelar", func() {
		formWindow.Close()
	})

	// Layout
	content := container.NewVBox(
		widget.NewLabel("Completa los datos de la nueva propuesta:"),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		container.NewHBox(cancelBtn, createBtn),
	)

	formWindow.SetContent(container.NewPadded(content))
	formWindow.Show()
}

// showGenerationMenuFyne shows options for generating documents after creating proposal
func showGenerationMenuFyne(app fyne.App, parent fyne.Window, request propapi.TextGenerationRequest) {
	genWindow := app.NewWindow("Generar Documentos")
	genWindow.Resize(fyne.NewSize(500, 400))
	genWindow.CenterOnScreen()

	title := widget.NewLabel("¿Qué documento quieres generar?")
	title.Alignment = fyne.TextAlignCenter

	// Buttons
	mdBtn := widget.NewButton("📝 Solo Texto (MD)", func() {
		client, err := createClient()
		if err != nil {
			dialog.ShowError(fmt.Errorf("error creando cliente: %v", err), genWindow)
			return
		}
		response, err := client.CreateProposal(request)
		if err != nil {
			dialog.ShowError(fmt.Errorf("error creando propuesta: %v", err), genWindow)
			return
		}
		mdPath := getDownloadPath(response.ID + ".md")
		_ = client.DownloadProposalFile(response.ID, "md", mdPath)
		homeDir, _ := os.UserHomeDir()
		openDirectory(filepath.Join(homeDir, "Downloads"))
		genWindow.Close()
		parent.Close()
	})

	htmlBtn := widget.NewButton("🌐 Generar HTML", func() {
		client, err := createClient()
		if err != nil {
			dialog.ShowError(fmt.Errorf("error creando cliente: %v", err), genWindow)
			return
		}
		response, err := client.CreateProposal(request)
		if err != nil {
			dialog.ShowError(fmt.Errorf("error creando propuesta: %v", err), genWindow)
			return
		}
		generateHTMLAndPDFFyne(app, genWindow, response.ID)
		genWindow.Close()
		parent.Close()
	})

	pdfBtn := widget.NewButton("📄 Generar PDF", func() {
		client, err := createClient()
		if err != nil {
			dialog.ShowError(fmt.Errorf("error creando cliente: %v", err), genWindow)
			return
		}
		response, err := client.CreateProposal(request)
		if err != nil {
			dialog.ShowError(fmt.Errorf("error creando propuesta: %v", err), genWindow)
			return
		}
		generateHTMLAndPDFFyne(app, genWindow, response.ID)
		genWindow.Close()
		parent.Close()
	})

	cancelBtn := widget.NewButton("🏠 Volver al Menú Principal", func() {
		genWindow.Close()
		parent.Close()
	})

	// Layout
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		container.NewCenter(container.NewVBox(mdBtn, htmlBtn, pdfBtn)),
		widget.NewSeparator(),
		container.NewCenter(cancelBtn),
	)

	genWindow.SetContent(container.NewPadded(content))
	genWindow.Show()
}

// showExistingProposalFlowFyne handles the complete flow for existing proposals with Fyne
func showExistingProposalFlowFyne(app fyne.App, parent fyne.Window, client *propapi.Client) {
	// Show loading dialog
	loadingDialog := dialog.NewProgressInfinite("Cargando", "Obteniendo propuestas...", parent)
	loadingDialog.Show()

	// Fetch proposals in goroutine
	go func() {
		proposals, err := client.GetProposals()
		loadingDialog.Hide()

		if err != nil {
			dialog.ShowError(fmt.Errorf("error obteniendo propuestas: %v", err), parent)
			return
		}

		if len(proposals) == 0 {
			dialog.ShowInformation("Sin propuestas", "No hay propuestas disponibles", parent)
			return
		}

		// Show proposals list
		showProposalsListFyne(app, parent, client, proposals)
	}()
}

// showProposalsListFyne shows a list of proposals for selection
func showProposalsListFyne(app fyne.App, parent fyne.Window, client *propapi.Client, proposals []propapi.Proposal) {
	listWindow := app.NewWindow("Propuestas Existentes")
	listWindow.Resize(fyne.NewSize(900, 600))
	listWindow.CenterOnScreen()

	// Create list widget
	list := widget.NewList(
		func() int {
			return len(proposals)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("ID"),
				widget.NewLabel("Título"),
				widget.NewLabel("Subtítulo"),
				widget.NewLabel("Creada"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			proposal := proposals[id]
			container := obj.(*fyne.Container)
			container.Objects[0].(*widget.Label).SetText(proposal.ID)
			container.Objects[1].(*widget.Label).SetText(proposal.Title)
			container.Objects[2].(*widget.Label).SetText(proposal.Subtitle)
			container.Objects[3].(*widget.Label).SetText(proposal.CreatedAt.Format("2006-01-02 15:04"))
		},
	)

	// Selection handler
	list.OnSelected = func(id widget.ListItemID) {
		selectedProposal := proposals[id]
		showProposalManagementFyne(app, listWindow, client, selectedProposal)
	}

	// Layout
	content := container.NewVBox(
		widget.NewLabel("Selecciona una propuesta:"),
		widget.NewSeparator(),
		list,
	)

	listWindow.SetContent(container.NewPadded(content))
	listWindow.Show()
}

// showProposalManagementFyne shows the management interface for a specific proposal
func showProposalManagementFyne(app fyne.App, parent fyne.Window, client *propapi.Client, proposal propapi.Proposal) {
	mgmtWindow := app.NewWindow("Gestionar: " + proposal.Title)
	mgmtWindow.Resize(fyne.NewSize(500, 600))
	mgmtWindow.CenterOnScreen()

	title := widget.NewLabel("Selecciona una acción:")
	title.Alignment = fyne.TextAlignCenter

	// Create action buttons based on available documents
	var buttons []fyne.CanvasObject

	// Always available actions
	buttons = append(buttons, widget.NewButton("📝 Ver propuesta (MD)", func() {
		mdPath := getDownloadPath(proposal.ID + ".md")
		if err := client.DownloadProposalFile(proposal.ID, "md", mdPath); err != nil {
			dialog.ShowError(fmt.Errorf("MD no disponible aún"), mgmtWindow)
		} else {
			dialog.ShowInformation("Descarga MD", "MD descargado en carpeta de Descargas", mgmtWindow)
		}
		homeDir, _ := os.UserHomeDir()
		openDirectory(filepath.Join(homeDir, "Downloads"))
	}))

	buttons = append(buttons, widget.NewButton("🛠️ Regenerar (título/subtítulo/prompt)", func() {
		regenerateProposalFyne(app, mgmtWindow, client, &proposal)
	}))

	buttons = append(buttons, widget.NewButton("✍️ Cambiar solo título/subtítulo", func() {
		updateTitleSubtitleFyne(app, mgmtWindow, client, &proposal)
	}))

	// HTML actions
	if proposal.HTMLURL != "" {
		buttons = append(buttons, widget.NewButton("🌐 Ver HTML", func() {
			downloadPath := getDownloadPath(proposal.ID + ".html")
			if err := client.DownloadProposalFile(proposal.ID, "html", downloadPath); err != nil {
				dialog.ShowError(fmt.Errorf("HTML no disponible aún"), mgmtWindow)
			} else {
				dialog.ShowInformation("Descarga HTML", "HTML descargado en carpeta de Descargas", mgmtWindow)
			}
			homeDir, _ := os.UserHomeDir()
			openDirectory(filepath.Join(homeDir, "Downloads"))
		}))
		buttons = append(buttons, widget.NewButton("🔁 Regenerar HTML", func() {
			generateHTMLFyne(app, mgmtWindow, client, proposal.ID)
		}))
	} else {
		buttons = append(buttons, widget.NewButton("🌐 Generar HTML", func() {
			generateHTMLFyne(app, mgmtWindow, client, proposal.ID)
		}))
	}

	// PDF actions
	if proposal.PDFURL != "" {
		buttons = append(buttons, widget.NewButton("📄 Ver PDF", func() {
			downloadPath := getDownloadPath(proposal.ID + ".pdf")
			if err := client.DownloadProposalFile(proposal.ID, "pdf", downloadPath); err != nil {
				dialog.ShowError(fmt.Errorf("PDF no disponible aún"), mgmtWindow)
			} else {
				dialog.ShowInformation("Descarga PDF", "PDF descargado en carpeta de Descargas", mgmtWindow)
			}
			homeDir, _ := os.UserHomeDir()
			openDirectory(filepath.Join(homeDir, "Downloads"))
		}))
		buttons = append(buttons, widget.NewButton("🔁 Regenerar PDF", func() {
			generatePDFFyne(app, mgmtWindow, client, proposal.ID)
		}))
	} else {
		buttons = append(buttons, widget.NewButton("📄 Generar PDF", func() {
			generatePDFFyne(app, mgmtWindow, client, proposal.ID)
		}))
	}

	// Additional actions
	buttons = append(buttons, widget.NewButton("✏️ Modificar propuesta", func() {
		modifyProposalFyne(app, mgmtWindow, client, proposal)
	}))

	buttons = append(buttons, widget.NewButton("📥 Descargar archivos", func() {
		downloadSpecificProposal(client, proposal)
		dialog.ShowInformation("Descarga", "Archivos descargados en carpeta de Descargas", mgmtWindow)
	}))

	buttons = append(buttons, widget.NewButton("🏠 Volver al Menú Principal", func() {
		mgmtWindow.Close()
		parent.Close()
	}))

	// Layout
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		container.NewGridWithColumns(1, buttons...),
	)

	mgmtWindow.SetContent(container.NewPadded(content))
	mgmtWindow.Show()
}

// regenerateProposalFyne shows form to regenerate proposal
func regenerateProposalFyne(app fyne.App, parent fyne.Window, client *propapi.Client, proposal *propapi.Proposal) {
	formWindow := app.NewWindow("Regenerar Propuesta: " + proposal.Title)
	formWindow.Resize(fyne.NewSize(700, 500))
	formWindow.CenterOnScreen()

	// Form fields with current values
	titleEntry := widget.NewEntry()
	titleEntry.SetText(proposal.Title)

	subtitleEntry := widget.NewEntry()
	subtitleEntry.SetText(proposal.Subtitle)

	promptEntry := widget.NewMultiLineEntry()
	promptEntry.SetText(proposal.Prompt)
	promptEntry.Resize(fyne.NewSize(0, 200))

	// Form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Título:", Widget: titleEntry},
			{Text: "Subtítulo:", Widget: subtitleEntry},
			{Text: "Prompt:", Widget: promptEntry},
		},
	}

	// Buttons
	regenerateBtn := widget.NewButton("Regenerar", func() {
		title := strings.TrimSpace(titleEntry.Text)
		subtitle := strings.TrimSpace(subtitleEntry.Text)
		prompt := strings.TrimSpace(promptEntry.Text)

		body := map[string]string{
			"title":   title,
			"subtitle": subtitle,
			"prompt":  prompt,
			"model":   "gpt-5-chat-latest",
		}
		b, _ := json.Marshal(body)
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/proposal/%s/regenerate", client.BaseURL, proposal.ID), bytes.NewBuffer(b))
		if err != nil {
			dialog.ShowError(fmt.Errorf("no se pudo crear la solicitud"), formWindow)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		if client.AuthFunc != nil {
			client.AuthFunc(req)
		}
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			dialog.ShowError(fmt.Errorf("error de red"), formWindow)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			dialog.ShowError(fmt.Errorf("fallo al regenerar: %s", string(bodyBytes)), formWindow)
			return
		}

		// Update local proposal state
		proposal.Title = title
		proposal.Subtitle = subtitle
		proposal.Prompt = prompt
		proposal.HTMLURL = ""
		proposal.PDFURL = ""

		dialog.ShowInformation("Éxito", "Texto regenerado. Generando HTML y PDF...", formWindow)
		generateHTMLAndPDFFyne(app, formWindow, proposal.ID)
		formWindow.Close()
	})

	cancelBtn := widget.NewButton("Cancelar", func() {
		formWindow.Close()
	})

	// Layout
	content := container.NewVBox(
		widget.NewLabel("Edita los campos para regenerar el contenido (MD se reemplaza)"),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		container.NewHBox(cancelBtn, regenerateBtn),
	)

	formWindow.SetContent(container.NewPadded(content))
	formWindow.Show()
}

// updateTitleSubtitleFyne shows form to update title and subtitle
func updateTitleSubtitleFyne(app fyne.App, parent fyne.Window, client *propapi.Client, proposal *propapi.Proposal) {
	formWindow := app.NewWindow("Actualizar Título/Subtítulo: " + proposal.Title)
	formWindow.Resize(fyne.NewSize(600, 300))
	formWindow.CenterOnScreen()

	// Form fields with current values
	titleEntry := widget.NewEntry()
	titleEntry.SetText(proposal.Title)

	subtitleEntry := widget.NewEntry()
	subtitleEntry.SetText(proposal.Subtitle)

	// Form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Título:", Widget: titleEntry},
			{Text: "Subtítulo:", Widget: subtitleEntry},
		},
	}

	// Buttons
	updateBtn := widget.NewButton("Actualizar", func() {
		title := strings.TrimSpace(titleEntry.Text)
		subtitle := strings.TrimSpace(subtitleEntry.Text)

		body := map[string]string{
			"title":    title,
			"subtitle": subtitle,
		}
		b, _ := json.Marshal(body)
		req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/proposal/%s/title-subtitle", client.BaseURL, proposal.ID), bytes.NewBuffer(b))
		if err != nil {
			dialog.ShowError(fmt.Errorf("no se pudo crear la solicitud"), formWindow)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		if client.AuthFunc != nil {
			client.AuthFunc(req)
		}
		resp, err := (&http.Client{}).Do(req)
		if err != nil {
			dialog.ShowError(fmt.Errorf("error de red"), formWindow)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			dialog.ShowError(fmt.Errorf("fallo al actualizar: %s", string(bodyBytes)), formWindow)
			return
		}

		proposal.Title = title
		proposal.Subtitle = subtitle
		dialog.ShowInformation("Éxito", "Título/Subtítulo actualizados. Si deseas verlos en HTML/PDF, vuelve a generarlos.", formWindow)
		formWindow.Close()
	})

	cancelBtn := widget.NewButton("Cancelar", func() {
		formWindow.Close()
	})

	// Layout
	content := container.NewVBox(
		form,
		widget.NewSeparator(),
		container.NewHBox(cancelBtn, updateBtn),
	)

	formWindow.SetContent(container.NewPadded(content))
	formWindow.Show()
}

// modifyProposalFyne shows form to modify proposal
func modifyProposalFyne(app fyne.App, parent fyne.Window, client *propapi.Client, proposal propapi.Proposal) {
	formWindow := app.NewWindow("Modificar Propuesta: " + proposal.Title)
	formWindow.Resize(fyne.NewSize(600, 500))
	formWindow.CenterOnScreen()

	// Form fields
	titleEntry := widget.NewEntry()
	titleEntry.SetText(proposal.Title)

	subtitleEntry := widget.NewEntry()
	subtitleEntry.SetText(proposal.Subtitle)

	promptEntry := widget.NewMultiLineEntry()
	promptEntry.SetText(proposal.Title + " modificada")
	promptEntry.Resize(fyne.NewSize(0, 200))

	// Form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Título:", Widget: titleEntry},
			{Text: "Subtítulo:", Widget: subtitleEntry},
			{Text: "Prompt:", Widget: promptEntry},
		},
	}

	// Buttons
	modifyBtn := widget.NewButton("Modificar", func() {
		title := strings.TrimSpace(titleEntry.Text)
		subtitle := strings.TrimSpace(subtitleEntry.Text)
		prompt := strings.TrimSpace(promptEntry.Text)

		if prompt == "" {
			dialog.ShowError(fmt.Errorf("el prompt no puede estar vacío"), formWindow)
			return
		}

		// Create request
		request := propapi.TextGenerationRequest{
			Title:    title,
			Subtitle: subtitle,
			Prompt:   prompt,
			Model:    "gpt-5-chat-latest",
		}

		// Send modification request
		jsonData, err := json.Marshal(request)
		if err != nil {
			dialog.ShowError(fmt.Errorf("error al serializar la solicitud: %v", err), formWindow)
			return
		}

		req, err := http.NewRequest("PUT", client.BaseURL+"/proposal/"+proposal.ID, bytes.NewBuffer(jsonData))
		if err != nil {
			dialog.ShowError(fmt.Errorf("error al crear la solicitud: %v", err), formWindow)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		if client.AuthFunc != nil {
			client.AuthFunc(req)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			dialog.ShowError(fmt.Errorf("error al enviar la solicitud: %v", err), formWindow)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			dialog.ShowError(fmt.Errorf("error del servidor (%d): %s", resp.StatusCode, string(body)), formWindow)
			return
		}

		var response propapi.ModifyProposalResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			dialog.ShowError(fmt.Errorf("error al decodificar la respuesta: %v", err), formWindow)
			return
		}

		dialog.ShowInformation("Éxito", "Propuesta modificada y documentos generados", formWindow)
		generateHTMLAndPDFFyne(app, formWindow, proposal.ID)
		formWindow.Close()
	})

	cancelBtn := widget.NewButton("Cancelar", func() {
		formWindow.Close()
	})

	// Layout
	content := container.NewVBox(
		widget.NewLabel("Ingresa el nuevo prompt:"),
		widget.NewSeparator(),
		form,
		widget.NewSeparator(),
		container.NewHBox(cancelBtn, modifyBtn),
	)

	formWindow.SetContent(container.NewPadded(content))
	formWindow.Show()
}

// generateHTMLFyne generates HTML with Fyne interface
func generateHTMLFyne(app fyne.App, parent fyne.Window, client *propapi.Client, proposalID string) {
	// Show progress dialog
	progressDialog := dialog.NewProgressInfinite("Generando HTML", "Conectando con la API...", parent)
	progressDialog.Show()

	go func() {
		_, err := client.GenerateHTML(proposalID)
		if err != nil {
			progressDialog.Hide()
			dialog.ShowError(fmt.Errorf("error al generar HTML: %v", err), parent)
			return
		}

		progressDialog.Hide()
		dialog.ShowInformation("Éxito", "HTML generado exitosamente", parent)
		
		// Download HTML file
		client.DownloadProposalFile(proposalID, "html", getDownloadPath(proposalID+".html"))
	}()
}

// generatePDFFyne generates PDF with Fyne interface
func generatePDFFyne(app fyne.App, parent fyne.Window, client *propapi.Client, proposalID string) {
	// Create mode selection window
	modeWindow := app.NewWindow("Generar PDF")
	modeWindow.Resize(fyne.NewSize(400, 200))
	modeWindow.CenterOnScreen()
	
	modeSelect := widget.NewSelect([]string{"normal", "dapec", "oscuro"}, func(value string) {
		modeWindow.Close()
		
		// Show progress dialog
		progressDialog := dialog.NewProgressInfinite("Generando PDF", "Generando PDF en modo "+value+"...", parent)
		progressDialog.Show()

		go func() {
			pdfRequest := propapi.PDFGenerationRequest{
				ProposalID: proposalID,
				Modo:       value,
			}
			jsonData, err := json.Marshal(pdfRequest)
			if err != nil {
				progressDialog.Hide()
				dialog.ShowError(fmt.Errorf("error al serializar solicitud PDF: %v", err), parent)
				return
			}

			req, err := http.NewRequest("POST", client.BaseURL+"/generate-pdf", bytes.NewBuffer(jsonData))
			if err != nil {
				progressDialog.Hide()
				dialog.ShowError(fmt.Errorf("error al crear solicitud PDF: %v", err), parent)
				return
			}
			req.Header.Set("Content-Type", "application/json")
			if client.AuthFunc != nil {
			client.AuthFunc(req)
		}

			httpClient := &http.Client{}
			resp, err := httpClient.Do(req)
			if err != nil {
				progressDialog.Hide()
				dialog.ShowError(fmt.Errorf("error al generar PDF: %v", err), parent)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				body, _ := io.ReadAll(resp.Body)
				progressDialog.Hide()
				dialog.ShowError(fmt.Errorf("error del servidor (%d): %s", resp.StatusCode, string(body)), parent)
				return
			}

			var pdfResponse propapi.PDFGenerationResponse
			if err := json.NewDecoder(resp.Body).Decode(&pdfResponse); err != nil {
				progressDialog.Hide()
				dialog.ShowError(fmt.Errorf("error al decodificar respuesta PDF: %v", err), parent)
				return
			}

			progressDialog.Hide()
			dialog.ShowInformation("Éxito", "PDF generado exitosamente", parent)

			// Download PDF file
			filepath := getDownloadPath(proposalID + ".pdf")
			if err := client.DownloadProposalFile(proposalID, "pdf", filepath); err != nil {
				dialog.ShowError(fmt.Errorf("error al descargar PDF: %v", err), parent)
				return
			}

			// Open PDF file
			openFile(filepath)
		}()
	})
	modeSelect.SetSelected("normal")

	content := container.NewVBox(
		widget.NewLabel("Selecciona el modo de impresión del PDF:"),
		modeSelect,
		widget.NewButton("Generar", func() {
			if modeSelect.Selected != "" {
				modeSelect.OnChanged(modeSelect.Selected)
			}
		}),
		widget.NewButton("Cancelar", func() {
			modeWindow.Close()
		}),
	)
	
	modeWindow.SetContent(container.NewPadded(content))
	modeWindow.Show()
}

// generateHTMLAndPDFFyne generates both HTML and PDF
func generateHTMLAndPDFFyne(app fyne.App, parent fyne.Window, proposalID string) {
	// Show progress dialog
	progressDialog := dialog.NewProgressInfinite("Generando Documentos", "Generando HTML y PDF...", parent)
	progressDialog.Show()

	go func() {
		// Create client
		client, err := createClient()
		if err != nil {
			progressDialog.Hide()
			dialog.ShowError(fmt.Errorf("error creando cliente: %v", err), parent)
			return
		}

		// Generate HTML
		_, err = client.GenerateHTML(proposalID)
		if err != nil {
			progressDialog.Hide()
			dialog.ShowError(fmt.Errorf("error al generar HTML: %v", err), parent)
			return
		}
		// Generate PDF
		_, err = client.GeneratePDF(proposalID, "normal")
		if err != nil {
			progressDialog.Hide()
			dialog.ShowError(fmt.Errorf("error al generar PDF: %v", err), parent)
			return
		}

		progressDialog.Hide()
		dialog.ShowInformation("Éxito", "HTML y PDF generados exitosamente", parent)
	}()
}


