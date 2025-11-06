package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/osmargm1202/orgm/pkg/propapi"
	"github.com/spf13/cobra"
)

// PropCmd represents the prop command
var PropCmd = &cobra.Command{
	Use:   "prop",
	Short: "Gestión de propuestas con TUI",
	Long:  `Gestión de propuestas con interfaz TUI usando Bubbletea.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Ensure token for all subcommands
		if _, err := EnsureGCloudIDToken(); err != nil {
			return fmt.Errorf("error obteniendo token: %w", err)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		runPropTUI()
	},
}

func init() {
	// No subcommands needed - all functionality is in the TUI
}

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Padding(1)
	selectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	normalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)
	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true)
	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("33"))
)

// State type for navigation
type state int

const (
	stateUnified state = iota // Estado unificado - todo en una pantalla
	stateMainMenu
	stateNewProposal
	stateGenerationMenu
	stateProposalsList
	stateProposalManagement
	stateRegenerate
	stateUpdateTitleSubtitle
	stateModifyProposal
	statePDFMode
	stateLoading
)

// Main model that holds all sub-models
type appModel struct {
	state state
	
	// Client
	client *propapi.Client
	
	// Main menu
	mainMenuCursor int
	mainMenuItems  []string
	
	// New proposal form
	titleInput    textarea.Model
	subtitleInput textarea.Model
	promptInput   textarea.Model
	formFocus     int // 0: title, 1: subtitle, 2: prompt
	
	// Generation menu
	genMenuCursor int
	genMenuItems  []string
	genRequest    propapi.TextGenerationRequest
	
	// Proposals list
	proposalsList list.Model
	proposals     []propapi.Proposal
	
	// Proposal management
	selectedProposal *propapi.Proposal
	mgmtMenuCursor   int
	mgmtMenuItems    []string
	
	// Regenerate form
	regTitleInput    textarea.Model
	regSubtitleInput textarea.Model
	regPromptInput   textarea.Model
	regFormFocus     int
	
	// Update title/subtitle form
	updateTitleInput    textarea.Model
	updateSubtitleInput textarea.Model
	updateFormFocus     int
	
	// Modify form
	modTitleInput    textarea.Model
	modSubtitleInput textarea.Model
	modPromptInput   textarea.Model
	modFormFocus     int
	
	// PDF mode
	pdfModeCursor int
	pdfModeItems  []string
	
	// Loading
	loadingSpinner spinner.Model
	loadingMsg     string
	
	// Error/Success messages
	message     string
	messageType string // "error", "success", "info"
	
	// Unified view control
	activeSection  int  // 0: main menu, 1: content area, 2: management menu
	showProposals  bool
	showNewForm    bool
	showManagement bool
	
	width  int
	height int
}

type errMsg error
type proposalsLoadedMsg []propapi.Proposal
type proposalCreatedMsg *propapi.TextGenerationResponse
type htmlGeneratedMsg *propapi.HTMLGenerationResponse
type pdfGeneratedMsg *propapi.PDFGenerationResponse
type operationCompletedMsg struct {
	message string
	nextState state
}

// addSpaceAfterEmoji agrega un espacio después del primer emoji en un string
func addSpaceAfterEmoji(s string) string {
	// Convertir string a runes para manejar correctamente emojis (que pueden ocupar múltiples bytes)
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	
	// Si el primer carácter ya tiene un espacio después, no hacer nada
	if len(runes) > 1 && runes[1] == ' ' {
		return s
	}
	
	// Agregar espacio después del primer carácter (emoji)
	if len(runes) > 1 {
		return string(runes[0]) + " " + string(runes[1:])
	}
	return s
}

func runPropTUI() {
	// Create client
	client, err := createClient()
	if err != nil {
		fmt.Printf("Error creando cliente: %v\n", err)
		os.Exit(1)
	}
	
	// Initialize model
	model := initialModel(client)
	
	// Run program
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error ejecutando TUI: %v\n", err)
		os.Exit(1)
	}
}

func initialModel(client *propapi.Client) appModel {
	// Main menu
	mainMenuItems := []string{
		"🆕 Nueva Propuesta",
		"📁 Propuesta Existente",
	}
	
	// Generation menu
	genMenuItems := []string{
		"📝 Solo Texto (MD)",
		"🌐 Generar HTML",
		"📄 Generar PDF",
		"🏠 Volver al Menú Principal",
	}
	
	// PDF mode
	pdfModeItems := []string{"normal", "dapec", "oscuro"}
	
	// Title input - cambiado a textarea multilínea
	ti := textarea.New()
	ti.Placeholder = "Ingresa el título"
	ti.CharLimit = 500
	ti.SetWidth(100)
	ti.SetHeight(3)
	ti.Focus()
	
	// Subtitle input - cambiado a textarea multilínea
	si := textarea.New()
	si.Placeholder = "Ingresa el subtítulo"
	si.CharLimit = 500
	si.SetWidth(100)
	si.SetHeight(3)
	
	// Prompt textarea
	ta := textarea.New()
	ta.Placeholder = "Escribe aquí tu prompt..."
	ta.CharLimit = 5000
	ta.SetWidth(100)
	ta.SetHeight(12)
	ta.Focus()
	
	// Regenerate inputs - título y subtítulo multilínea
	rti := textarea.New()
	rti.CharLimit = 500
	rti.SetWidth(100)
	rti.SetHeight(3)
	
	rsi := textarea.New()
	rsi.CharLimit = 500
	rsi.SetWidth(100)
	rsi.SetHeight(3)
	
	rta := textarea.New()
	rta.CharLimit = 5000
	rta.SetWidth(100)
	rta.SetHeight(12)
	
	// Update inputs - título y subtítulo multilínea
	uti := textarea.New()
	uti.CharLimit = 500
	uti.SetWidth(100)
	uti.SetHeight(3)
	
	usi := textarea.New()
	usi.CharLimit = 500
	usi.SetWidth(100)
	usi.SetHeight(3)
	
	// Modify inputs - título y subtítulo multilínea
	mti := textarea.New()
	mti.CharLimit = 500
	mti.SetWidth(100)
	mti.SetHeight(3)
	
	msi := textarea.New()
	msi.CharLimit = 500
	msi.SetWidth(100)
	msi.SetHeight(3)
	
	mta := textarea.New()
	mta.CharLimit = 5000
	mta.SetWidth(100)
	mta.SetHeight(12)
	
	// Spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	
	// Proposals list - inicializar con altura mínima para 10 items
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(lipgloss.Color("39")).Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(lipgloss.Color("39"))
	proposalsList := list.New([]list.Item{}, delegate, 0, 0)
	proposalsList.Title = "Propuestas"
	proposalsList.SetShowStatusBar(false)
	proposalsList.SetFilteringEnabled(true)
	proposalsList.Styles.Title = titleStyle
	
	return appModel{
		state:            stateUnified,
		client:           client,
		mainMenuCursor:   0,
		mainMenuItems:    mainMenuItems,
		activeSection:    0,
		showProposals:    false,
		showNewForm:      false,
		showManagement:   false,
		titleInput:       ti,
		subtitleInput:    si,
		promptInput:      ta,
		formFocus:        0,
		genMenuCursor:    0,
		genMenuItems:     genMenuItems,
		proposalsList:    proposalsList,
		proposals:        []propapi.Proposal{},
		mgmtMenuCursor:   0,
		mgmtMenuItems:    []string{}, // Inicializar vacío, se llenará cuando se seleccione una propuesta
		regTitleInput:    rti,
		regSubtitleInput: rsi,
		regPromptInput:   rta,
		regFormFocus:     0,
		updateTitleInput: uti,
		updateSubtitleInput: usi,
		updateFormFocus:  0,
		modTitleInput:    mti,
		modSubtitleInput: msi,
		modPromptInput:   mta,
		modFormFocus:     0,
		pdfModeCursor:    0,
		pdfModeItems:     pdfModeItems,
		loadingSpinner:   s,
		width:            120,
		height:           24,
	}
}

func (m appModel) Init() tea.Cmd {
	return tea.Batch(
		// titleInput es textarea ahora, Focus() no aplica
		m.loadingSpinner.Tick,
	)
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Ajustar ancho de lista con más margen
		listWidth := msg.Width - 8
		if listWidth < 100 {
			listWidth = 100
		}
		m.proposalsList.SetWidth(listWidth)
		// Calcular altura dinámicamente según el espacio disponible
		// Reservar espacio para: título (2), menú principal (3), separador (1), título contenido (1), instrucciones (2), mensajes (2) = ~11 líneas
		// Si hay menú de gestión, reservar más espacio
		headerLines := 11
		if m.showManagement {
			headerLines += 15 // espacio para menú de gestión
		}
		if m.showNewForm {
			headerLines += 20 // espacio para formulario
		}
		calculatedHeight := msg.Height - headerLines
		// Altura mínima para mostrar al menos 5 propuestas (10 líneas)
		minHeight := 10
		if calculatedHeight < minHeight {
			calculatedHeight = minHeight
		}
		// Altura máxima razonable
		maxHeight := 30
		if calculatedHeight > maxHeight {
			calculatedHeight = maxHeight
		}
		m.proposalsList.SetHeight(calculatedHeight)
		// Ajustar ancho de textareas
		textareaWidth := msg.Width - 8
		if textareaWidth < 80 {
			textareaWidth = 80
		}
		if textareaWidth > 100 {
			textareaWidth = 100
		}
		m.titleInput.SetWidth(textareaWidth)
		m.subtitleInput.SetWidth(textareaWidth)
		m.promptInput.SetWidth(textareaWidth)
		m.regTitleInput.SetWidth(textareaWidth)
		m.regSubtitleInput.SetWidth(textareaWidth)
		m.regPromptInput.SetWidth(textareaWidth)
		m.updateTitleInput.SetWidth(textareaWidth)
		m.updateSubtitleInput.SetWidth(textareaWidth)
		m.modTitleInput.SetWidth(textareaWidth)
		m.modSubtitleInput.SetWidth(textareaWidth)
		m.modPromptInput.SetWidth(textareaWidth)
		return m, nil
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
		
		// Handle state-specific updates
		switch m.state {
		case stateUnified:
			return (&m).updateUnified(msg)
		case stateMainMenu:
			return m.updateMainMenu(msg)
		case stateNewProposal:
			return m.updateNewProposal(msg)
		case stateGenerationMenu:
			return m.updateGenerationMenu(msg)
		case stateProposalsList:
			// Usar puntero para actualizar el modelo correctamente
			return (&m).updateProposalsList(msg)
		case stateProposalManagement:
			// Usar puntero para actualizar el modelo correctamente
			return (&m).updateProposalManagement(msg)
		case stateRegenerate:
			return m.updateRegenerate(msg)
		case stateUpdateTitleSubtitle:
			return m.updateUpdateTitleSubtitle(msg)
		case stateModifyProposal:
			return m.updateModifyProposal(msg)
		case statePDFMode:
			return m.updatePDFMode(msg)
		case stateLoading:
			var cmd tea.Cmd
			m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
			return m, cmd
		}
		
	case errMsg:
		m.state = stateUnified
		m.message = msg.Error()
		m.messageType = "error"
		return m, nil
		
	case proposalsLoadedMsg:
		m.proposals = []propapi.Proposal(msg)
		m.state = stateUnified
		m.showProposals = true
		m.updateProposalsListItems()
		return m, nil
		
	case proposalCreatedMsg:
		m.state = stateUnified
		m.showNewForm = false
		m.showProposals = false
		m.genRequest = propapi.TextGenerationRequest{}
		m.message = fmt.Sprintf("Propuesta creada exitosamente! ID: %s", msg.ID)
		m.messageType = "success"
		m.activeSection = 0
		return m, nil
		
	case htmlGeneratedMsg:
		if m.selectedProposal != nil {
			m.selectedProposal.HTMLURL = msg.HTMLURL
			m.updateManagementMenu()
		}
		m.message = fmt.Sprintf("HTML generado exitosamente: %s", msg.HTMLURL)
		m.messageType = "success"
		m.state = stateUnified
		m.showManagement = true
		m.activeSection = 2
		return m, nil
		
	case pdfGeneratedMsg:
		if m.selectedProposal != nil {
			m.selectedProposal.PDFURL = msg.PDFURL
			m.updateManagementMenu()
		}
		m.message = fmt.Sprintf("PDF generado exitosamente: %s", msg.PDFURL)
		m.messageType = "success"
		m.state = stateUnified
		m.showManagement = true
		m.activeSection = 2
		return m, nil
		
	case operationCompletedMsg:
		m.message = msg.message
		m.messageType = "success"
		// Si el nextState es stateProposalManagement, mantener en unified
		if msg.nextState == stateProposalManagement {
			m.state = stateUnified
			m.showManagement = true
			m.activeSection = 2
		} else if msg.nextState == stateMainMenu {
			m.state = stateUnified
			m.activeSection = 0
		} else {
			m.state = msg.nextState
		}
		return m, nil
		
	case htmlPDFRegeneratedMsg:
		if m.selectedProposal != nil {
			m.selectedProposal.HTMLURL = msg.htmlURL
			m.selectedProposal.PDFURL = msg.pdfURL
			m.updateManagementMenu()
		}
		m.message = fmt.Sprintf("HTML y PDF regenerados. HTML: %s, PDF: %s", msg.htmlURL, msg.pdfURL)
		m.messageType = "success"
		m.state = stateUnified
		m.showManagement = true
		m.activeSection = 2
		return m, nil
		
	case proposalRegeneratedMsg:
		if m.selectedProposal != nil {
			m.selectedProposal.Title = msg.title
			m.selectedProposal.Subtitle = msg.subtitle
			m.selectedProposal.Prompt = msg.prompt
			m.selectedProposal.HTMLURL = msg.htmlURL
			m.selectedProposal.PDFURL = msg.pdfURL
			m.updateManagementMenu()
		}
		m.message = "Propuesta regenerada, HTML y PDF generados exitosamente"
		m.messageType = "success"
		m.state = stateUnified
		m.showManagement = true
		m.activeSection = 2
		return m, nil
		
	case titleSubtitleUpdatedMsg:
		if m.selectedProposal != nil {
			m.selectedProposal.Title = msg.title
			m.selectedProposal.Subtitle = msg.subtitle
			m.updateManagementMenu()
		}
		m.message = "Título y subtítulo actualizados. Si deseas verlos en HTML/PDF, vuelve a generarlos."
		m.messageType = "success"
		m.state = stateUnified
		m.showManagement = true
		m.activeSection = 2
		return m, nil
		
	case proposalModifiedMsg:
		if m.selectedProposal != nil {
			m.selectedProposal.HTMLURL = msg.htmlURL
			m.selectedProposal.PDFURL = msg.pdfURL
			m.updateManagementMenu()
		}
		m.message = "Propuesta modificada y documentos generados exitosamente"
		m.messageType = "success"
		m.state = stateUnified
		m.showManagement = true
		m.activeSection = 2
		return m, nil
	}
	
	// Handle loading spinner
	if m.state == stateLoading {
		m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
		cmds = append(cmds, cmd)
	}
	
	return m, tea.Batch(cmds...)
}

func (m appModel) View() string {
	switch m.state {
	case stateUnified:
		return m.viewUnified()
	case stateMainMenu:
		return m.viewMainMenu()
	case stateNewProposal:
		return m.viewNewProposal()
	case stateGenerationMenu:
		return m.viewGenerationMenu()
	case stateProposalsList:
		return m.viewProposalsList()
	case stateProposalManagement:
		return m.viewProposalManagement()
	case stateRegenerate:
		return m.viewRegenerate()
	case stateUpdateTitleSubtitle:
		return m.viewUpdateTitleSubtitle()
	case stateModifyProposal:
		return m.viewModifyProposal()
	case statePDFMode:
		return m.viewPDFMode()
	case stateLoading:
		return m.viewLoading()
	default:
		return "Estado desconocido"
	}
}

// Unified view - everything in one screen
func (m appModel) viewUnified() string {
	var b strings.Builder
	
	// Calcular ancho de separador basado en el ancho de ventana
	separatorWidth := 60
	if m.width > 0 {
		separatorWidth = m.width - 8
		if separatorWidth < 40 {
			separatorWidth = 40
		}
		if separatorWidth > 80 {
			separatorWidth = 80
		}
	}
	
	// 1. Título principal (solo si no hay contenido activo o si estamos en el menú principal)
	showTitle := m.activeSection == 0 || (!m.showProposals && !m.showNewForm && !m.showManagement)
	if showTitle {
		b.WriteString(titleStyle.Render("📋 Gestor de Propuestas"))
		b.WriteString("\n")
	}
	
	// 2. Menú principal (solo visible cuando activeSection == 0 o cuando no hay contenido activo)
	showMainMenu := m.activeSection == 0 || (!m.showProposals && !m.showNewForm && !m.showManagement)
	if showMainMenu {
		if !showTitle {
			b.WriteString("\n")
		}
		b.WriteString("Selecciona una opción:\n")
		for i, item := range m.mainMenuItems {
			cursor := " "
			if m.activeSection == 0 && m.mainMenuCursor == i {
				cursor = ">"
				formattedItem := addSpaceAfterEmoji(item)
				b.WriteString(selectedStyle.Render(fmt.Sprintf("%s %s", cursor, formattedItem)))
			} else {
				formattedItem := addSpaceAfterEmoji(item)
				b.WriteString(normalStyle.Render(fmt.Sprintf("%s %s", cursor, formattedItem)))
			}
			b.WriteString("\n")
		}
	}
	
	// Separador solo si hay contenido después del menú principal
	if showMainMenu && (m.showProposals || m.showNewForm || m.showManagement) {
		b.WriteString("\n")
		b.WriteString(strings.Repeat("─", separatorWidth))
		b.WriteString("\n")
	}
	
	// 3. Área de contenido dinámico
	if m.showNewForm {
		// Mostrar formulario de nueva propuesta
		if !showMainMenu {
			b.WriteString("\n")
		}
		if m.activeSection == 1 {
			b.WriteString(selectedStyle.Render("🆕 Nueva Propuesta"))
		} else {
			b.WriteString(normalStyle.Render("🆕 Nueva Propuesta"))
		}
		b.WriteString("\n")
		
		// Title
		b.WriteString("Título:\n")
		if m.activeSection == 1 && m.formFocus == 0 {
			b.WriteString(selectedStyle.Render(m.titleInput.View()))
		} else {
			b.WriteString(m.titleInput.View())
		}
		b.WriteString("\n")
		
		// Subtitle
		b.WriteString("Subtítulo:\n")
		if m.activeSection == 1 && m.formFocus == 1 {
			b.WriteString(selectedStyle.Render(m.subtitleInput.View()))
		} else {
			b.WriteString(m.subtitleInput.View())
		}
		b.WriteString("\n")
		
		// Prompt
		b.WriteString("Prompt:\n")
		if m.activeSection == 1 && m.formFocus == 2 {
			b.WriteString(selectedStyle.Render(m.promptInput.View()))
		} else {
			b.WriteString(m.promptInput.View())
		}
		
	} else if m.showProposals {
		// Mostrar lista de propuestas
		if !showMainMenu {
			b.WriteString("\n")
		}
		if m.activeSection == 1 {
			b.WriteString(selectedStyle.Render("📁 Propuestas Existentes"))
		} else {
			b.WriteString(normalStyle.Render("📁 Propuestas Existentes"))
		}
		b.WriteString("\n")
		if len(m.proposals) > 0 {
			b.WriteString(m.proposalsList.View())
		} else {
			b.WriteString(infoStyle.Render("No hay propuestas disponibles"))
		}
	} else if !showMainMenu {
		// Área vacía solo si no hay menú principal
		b.WriteString("\n")
		b.WriteString(infoStyle.Render("Selecciona una opción del menú principal"))
	}
	
	// 4. Menú de gestión (si hay propuesta seleccionada)
	if m.showManagement && m.selectedProposal != nil {
		if m.showProposals || m.showNewForm || showMainMenu {
			b.WriteString("\n")
			b.WriteString(strings.Repeat("─", separatorWidth))
		}
		b.WriteString("\n")
		if m.activeSection == 2 {
			b.WriteString(selectedStyle.Render(fmt.Sprintf("📋 Gestionar: %s", m.selectedProposal.Title)))
		} else {
			b.WriteString(normalStyle.Render(fmt.Sprintf("📋 Gestionar: %s", m.selectedProposal.Title)))
		}
		b.WriteString("\n")
		if len(m.mgmtMenuItems) > 0 {
			for i, item := range m.mgmtMenuItems {
				cursor := " "
				if m.activeSection == 2 && m.mgmtMenuCursor == i {
					cursor = ">"
					formattedItem := addSpaceAfterEmoji(item)
					b.WriteString(selectedStyle.Render(fmt.Sprintf("%s %s", cursor, formattedItem)))
				} else {
					formattedItem := addSpaceAfterEmoji(item)
					b.WriteString(normalStyle.Render(fmt.Sprintf("%s %s", cursor, formattedItem)))
				}
				b.WriteString("\n")
			}
		} else {
			b.WriteString(infoStyle.Render("Cargando opciones..."))
		}
	}
	
	// 5. Instrucciones
	b.WriteString("\n")
	b.WriteString(normalStyle.Render("Tab: cambiar sección | ↑↓: navegar | Enter: seleccionar | Esc: limpiar/salir"))
	
	// 6. Mensajes
	if m.message != "" {
		b.WriteString("\n")
		if m.messageType == "error" {
			b.WriteString(errorStyle.Render(m.message))
		} else if m.messageType == "success" {
			b.WriteString(successStyle.Render(m.message))
		} else {
			b.WriteString(infoStyle.Render(m.message))
		}
	}
	
	return docStyle.Render(b.String())
}

func (m *appModel) updateUnified(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			// Cambiar entre secciones
			maxSection := 0
			if m.showProposals || m.showNewForm {
				maxSection = 1
			}
			if m.showManagement {
				maxSection = 2
			}
			m.activeSection = (m.activeSection + 1) % (maxSection + 1)
			// Limpiar mensajes al cambiar de sección
			m.message = ""
			return m, nil
			
		case "shift+tab":
			// Cambiar hacia atrás
			maxSection := 0
			if m.showProposals || m.showNewForm {
				maxSection = 1
			}
			if m.showManagement {
				maxSection = 2
			}
			m.activeSection--
			if m.activeSection < 0 {
				m.activeSection = maxSection
			}
			// Limpiar mensajes al cambiar de sección
			m.message = ""
			return m, nil
			
		case "up", "k":
			// Navegar dentro de la sección activa
			switch m.activeSection {
			case 0: // Main menu
				if m.mainMenuCursor > 0 {
					m.mainMenuCursor--
				}
				return m, nil
			case 1: // Content area
				if m.showProposals {
					// La lista maneja sus propias teclas
					var cmd tea.Cmd
					m.proposalsList, cmd = m.proposalsList.Update(msg)
					return m, cmd
				} else if m.showNewForm {
					// Cambiar entre campos del formulario
					if m.formFocus > 0 {
						m.formFocus--
					}
					return m, nil
				}
			case 2: // Management menu
				if m.mgmtMenuCursor > 0 {
					m.mgmtMenuCursor--
				}
				return m, nil
			}
			
		case "down", "j":
			// Navegar dentro de la sección activa
			switch m.activeSection {
			case 0: // Main menu
				if m.mainMenuCursor < len(m.mainMenuItems)-1 {
					m.mainMenuCursor++
				}
				return m, nil
			case 1: // Content area
				if m.showProposals {
					// La lista maneja sus propias teclas
					var cmd tea.Cmd
					m.proposalsList, cmd = m.proposalsList.Update(msg)
					return m, cmd
				} else if m.showNewForm {
					// Cambiar entre campos del formulario
					if m.formFocus < 2 {
						m.formFocus++
					}
					return m, nil
				}
			case 2: // Management menu
				if m.mgmtMenuCursor < len(m.mgmtMenuItems)-1 {
					m.mgmtMenuCursor++
				}
				return m, nil
			}
			
		case "enter":
			// Acción según sección activa
			switch m.activeSection {
			case 0: // Main menu
				return m.handleMainMenuSelectionInUnified()
			case 1: // Content area
				if m.showProposals {
					return m.handleProposalSelectionInUnified()
				} else if m.showNewForm {
					return m.handleNewProposalSubmitInUnified()
				}
			case 2: // Management menu
				return m.handleManagementSelectionInUnified()
			}
			
		case "esc":
			// Limpiar selecciones o salir
			if m.showManagement {
				m.showManagement = false
				m.selectedProposal = nil
				// Si hay propuestas o formulario activo, volver a esa sección, sino al menú principal
				if m.showProposals || m.showNewForm {
					m.activeSection = 1
				} else {
					m.activeSection = 0
				}
				m.message = ""
				return m, nil
			} else if m.showProposals || m.showNewForm {
				m.showProposals = false
				m.showNewForm = false
				m.activeSection = 0
				m.message = ""
				return m, nil
			}
			return m, tea.Quit
		}
	}
	
	// Actualizar componentes activos
	var cmd tea.Cmd
	if m.showNewForm && m.activeSection == 1 {
		// Actualizar textareas del formulario según el focus
		switch m.formFocus {
		case 0:
			m.titleInput, cmd = m.titleInput.Update(msg)
		case 1:
			m.subtitleInput, cmd = m.subtitleInput.Update(msg)
		case 2:
			m.promptInput, cmd = m.promptInput.Update(msg)
		}
	} else if m.showProposals && m.activeSection == 1 {
		m.proposalsList, cmd = m.proposalsList.Update(msg)
	}
	
	return m, cmd
}

func (m *appModel) handleMainMenuSelectionInUnified() (tea.Model, tea.Cmd) {
	switch m.mainMenuItems[m.mainMenuCursor] {
	case "🆕 Nueva Propuesta":
		m.showNewForm = true
		m.showProposals = false
		m.showManagement = false
		m.selectedProposal = nil
		m.activeSection = 1
		m.formFocus = 0
		m.message = ""
		return m, nil
		
	case "📁 Propuesta Existente":
		m.showProposals = true
		m.showNewForm = false
		m.showManagement = false
		m.selectedProposal = nil
		m.activeSection = 1
		m.state = stateLoading
		m.loadingMsg = "Cargando propuestas..."
		return m, tea.Batch(m.loadingSpinner.Tick, m.loadProposals())
	}
	return m, nil
}

func (m *appModel) handleProposalSelectionInUnified() (tea.Model, tea.Cmd) {
	if !m.showProposals {
		return m, nil
	}
	selected := m.proposalsList.SelectedItem()
	if selected != nil {
		item := selected.(proposalItem)
		for i := range m.proposals {
			if m.proposals[i].ID == item.id {
				m.selectedProposal = &m.proposals[i]
				m.showManagement = true
				m.updateManagementMenu()
				m.activeSection = 2
				m.message = ""
				return m, nil
			}
		}
	}
	return m, nil
}

func (m *appModel) handleNewProposalSubmitInUnified() (tea.Model, tea.Cmd) {
	if !m.showNewForm {
		return m, nil
	}
	
	title := strings.TrimSpace(m.titleInput.Value())
	subtitle := strings.TrimSpace(m.subtitleInput.Value())
	prompt := strings.TrimSpace(m.promptInput.Value())
	
	if prompt == "" {
		m.message = "El prompt no puede estar vacío"
		m.messageType = "error"
		return m, nil
	}
	
	m.genRequest = propapi.TextGenerationRequest{
		Title:    title,
		Subtitle: subtitle,
		Prompt:   prompt,
		Model:    "gpt-5-chat-latest",
	}
	
	m.state = stateGenerationMenu
	m.message = ""
	return m, nil
}

func (m *appModel) handleManagementSelectionInUnified() (tea.Model, tea.Cmd) {
	if m.selectedProposal == nil || len(m.mgmtMenuItems) == 0 {
		return m, nil
	}
	
	// Reutilizar la lógica existente de updateProposalManagement
	return m.updateProposalManagement(tea.KeyMsg{Type: tea.KeyEnter})
}

// Main menu views and updates
func (m appModel) updateMainMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.mainMenuCursor > 0 {
				m.mainMenuCursor--
			}
		case "down", "j":
			if m.mainMenuCursor < len(m.mainMenuItems)-1 {
				m.mainMenuCursor++
			}
		case "enter":
			switch m.mainMenuItems[m.mainMenuCursor] {
			case "🆕 Nueva Propuesta":
				m.state = stateNewProposal
				m.formFocus = 0
				m.message = ""
				return m, nil
			case "📁 Propuesta Existente":
				m.state = stateLoading
				m.loadingMsg = "Cargando propuestas..."
				return m, tea.Batch(m.loadingSpinner.Tick, m.loadProposals())
			}
		case "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m appModel) viewMainMenu() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("📋 Gestor de Propuestas"))
	b.WriteString("\n\n")
	
	for i, item := range m.mainMenuItems {
		cursor := " "
		if m.mainMenuCursor == i {
			cursor = ">"
			// Agregar espacio después del emoji
			formattedItem := addSpaceAfterEmoji(item)
			b.WriteString(selectedStyle.Render(fmt.Sprintf("%s %s", cursor, formattedItem)))
		} else {
			// Agregar espacio después del emoji
			formattedItem := addSpaceAfterEmoji(item)
			b.WriteString(normalStyle.Render(fmt.Sprintf("%s %s", cursor, formattedItem)))
		}
		b.WriteString("\n")
	}
	
	b.WriteString("\n")
	b.WriteString(normalStyle.Render("↑↓ para navegar | Enter para seleccionar | q para salir"))
	
	if m.message != "" {
		b.WriteString("\n\n")
		if m.messageType == "error" {
			b.WriteString(errorStyle.Render(m.message))
		} else if m.messageType == "success" {
			b.WriteString(successStyle.Render(m.message))
		} else {
			b.WriteString(infoStyle.Render(m.message))
		}
	}
	
	return docStyle.Render(b.String())
}

// New proposal views and updates
func (m appModel) updateNewProposal(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.formFocus = (m.formFocus + 1) % 3
			switch m.formFocus {
			case 0:
				m.titleInput.Focus()
				m.subtitleInput.Blur()
				m.promptInput.Blur()
				return m, nil
			case 1:
				m.titleInput.Blur()
				m.subtitleInput.Focus()
				m.promptInput.Blur()
				return m, nil
			case 2:
				m.titleInput.Blur()
				m.subtitleInput.Blur()
				m.promptInput.Focus()
				return m, nil
			}
		case "enter":
			if m.formFocus == 2 {
				// Create proposal
				title := strings.TrimSpace(m.titleInput.Value())
				subtitle := strings.TrimSpace(m.subtitleInput.Value())
				prompt := strings.TrimSpace(m.promptInput.Value())
				
				if prompt == "" {
					m.message = "El prompt no puede estar vacío"
					m.messageType = "error"
					return m, nil
				}
				
				m.genRequest = propapi.TextGenerationRequest{
					Title:    title,
					Subtitle: subtitle,
					Prompt:   prompt,
					Model:    "gpt-5-chat-latest",
				}
				
				m.state = stateGenerationMenu
				m.message = ""
				return m, nil
			}
		case "esc":
			m.state = stateMainMenu
			m.message = ""
			return m, nil
		}
	}
	
	// Update focused input
	switch m.formFocus {
	case 0:
		m.titleInput, cmd = m.titleInput.Update(msg)
	case 1:
		m.subtitleInput, cmd = m.subtitleInput.Update(msg)
	case 2:
		m.promptInput, cmd = m.promptInput.Update(msg)
	}
	
	return m, cmd
}

func (m appModel) viewNewProposal() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("🆕 Nueva Propuesta"))
	b.WriteString("\n\n")
	
	// Title
	b.WriteString("Título:")
	b.WriteString("\n")
	if m.formFocus == 0 {
		b.WriteString(selectedStyle.Render(m.titleInput.View()))
	} else {
		b.WriteString(m.titleInput.View())
	}
	b.WriteString("\n\n")
	
	// Subtitle
	b.WriteString("Subtítulo:")
	b.WriteString("\n")
	if m.formFocus == 1 {
		b.WriteString(selectedStyle.Render(m.subtitleInput.View()))
	} else {
		b.WriteString(m.subtitleInput.View())
	}
	b.WriteString("\n\n")
	
	// Prompt
	b.WriteString("Prompt:")
	b.WriteString("\n")
	if m.formFocus == 2 {
		b.WriteString(selectedStyle.Render(m.promptInput.View()))
	} else {
		b.WriteString(m.promptInput.View())
	}
	b.WriteString("\n\n")
	
	b.WriteString(normalStyle.Render("Tab para cambiar campo | Enter en prompt para continuar | Esc para volver"))
	
	if m.message != "" {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.message))
	}
	
	return docStyle.Render(b.String())
}

// Generation menu views and updates
func (m appModel) updateGenerationMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.genMenuCursor > 0 {
				m.genMenuCursor--
			}
		case "down", "j":
			if m.genMenuCursor < len(m.genMenuItems)-1 {
				m.genMenuCursor++
			}
		case "enter":
			switch m.genMenuItems[m.genMenuCursor] {
			case "📝 Solo Texto (MD)":
				m.state = stateLoading
				m.loadingMsg = "Creando propuesta y descargando MD..."
				return m, tea.Batch(m.loadingSpinner.Tick, m.createProposalAndDownloadMD())
			case "🌐 Generar HTML":
				m.state = stateLoading
				m.loadingMsg = "Creando propuesta y generando HTML..."
				return m, tea.Batch(m.loadingSpinner.Tick, m.createProposalAndGenerateHTML())
			case "📄 Generar PDF":
				m.state = stateLoading
				m.loadingMsg = "Creando propuesta y generando PDF..."
				return m, tea.Batch(m.loadingSpinner.Tick, m.createProposalAndGeneratePDF())
			case "🏠 Volver al Menú Principal":
				m.state = stateUnified
				m.showNewForm = false
				m.showProposals = false
				m.activeSection = 0
				m.message = ""
				return m, nil
			}
		case "esc":
			m.state = stateUnified
			m.showNewForm = false
			m.showProposals = false
			m.activeSection = 0
			m.message = ""
			return m, nil
		}
	}
	return m, nil
}

func (m appModel) viewGenerationMenu() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("📄 Generar Documentos"))
	b.WriteString("\n\n")
	b.WriteString("¿Qué documento quieres generar?")
	b.WriteString("\n\n")
	
	for i, item := range m.genMenuItems {
		cursor := " "
		if m.genMenuCursor == i {
			cursor = ">"
			// Agregar espacio después del emoji
			formattedItem := addSpaceAfterEmoji(item)
			b.WriteString(selectedStyle.Render(fmt.Sprintf("%s %s", cursor, formattedItem)))
		} else {
			// Agregar espacio después del emoji
			formattedItem := addSpaceAfterEmoji(item)
			b.WriteString(normalStyle.Render(fmt.Sprintf("%s %s", cursor, formattedItem)))
		}
		b.WriteString("\n")
	}
	
	b.WriteString("\n")
	b.WriteString(normalStyle.Render("↑↓ para navegar | Enter para seleccionar | Esc para volver"))
	
	if m.message != "" {
		b.WriteString("\n\n")
		if m.messageType == "error" {
			b.WriteString(errorStyle.Render(m.message))
		} else if m.messageType == "success" {
			b.WriteString(successStyle.Render(m.message))
		} else {
			b.WriteString(infoStyle.Render(m.message))
		}
	}
	
	return docStyle.Render(b.String())
}

// Proposals list views and updates
func (m *appModel) updateProposalsList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selected := m.proposalsList.SelectedItem()
			if selected != nil {
				item := selected.(proposalItem)
				// Find proposal
				for i := range m.proposals {
					if m.proposals[i].ID == item.id {
						m.selectedProposal = &m.proposals[i]
						m.state = stateProposalManagement
						// CORRECCIÓN: Asegurar que el menú se actualiza correctamente
						m.updateManagementMenu()
						m.message = ""
						return m, nil
					}
				}
			}
		case "esc":
			m.state = stateMainMenu
			m.message = ""
			return m, nil
		}
	}
	
	var cmd tea.Cmd
	m.proposalsList, cmd = m.proposalsList.Update(msg)
	return m, cmd
}

func (m appModel) viewProposalsList() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("📁 Propuestas Existentes"))
	b.WriteString("\n\n")
	
	if len(m.proposals) == 0 {
		b.WriteString(infoStyle.Render("No hay propuestas disponibles"))
		b.WriteString("\n\n")
		b.WriteString(normalStyle.Render("Esc para volver"))
		return docStyle.Render(b.String())
	}
	
	b.WriteString(m.proposalsList.View())
	b.WriteString("\n")
	b.WriteString(normalStyle.Render("↑↓ para navegar | Enter para seleccionar | Esc para volver | / para buscar"))
	
	if m.message != "" {
		b.WriteString("\n\n")
		if m.messageType == "error" {
			b.WriteString(errorStyle.Render(m.message))
		} else {
			b.WriteString(infoStyle.Render(m.message))
		}
	}
	
	return docStyle.Render(b.String())
}

// Proposal management views and updates
func (m *appModel) updateProposalManagement(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.mgmtMenuCursor > 0 {
				m.mgmtMenuCursor--
			}
		case "down", "j":
			if m.mgmtMenuCursor < len(m.mgmtMenuItems)-1 {
				m.mgmtMenuCursor++
			}
		case "enter":
			if m.selectedProposal == nil {
				return m, nil
			}
			
			if len(m.mgmtMenuItems) == 0 {
				m.updateManagementMenu()
			}
			
			if m.mgmtMenuCursor >= len(m.mgmtMenuItems) {
				m.mgmtMenuCursor = 0
			}
			
			action := m.mgmtMenuItems[m.mgmtMenuCursor]
			
			switch action {
			case "📝 Ver propuesta (MD)":
				m.state = stateLoading
				m.loadingMsg = "Descargando MD..."
				return m, tea.Batch(m.loadingSpinner.Tick, m.downloadProposalFile("md"))
			case "🛠️ Regenerar (título/subtítulo/prompt)":
				m.state = stateRegenerate
				m.regTitleInput.SetValue(m.selectedProposal.Title)
				m.regSubtitleInput.SetValue(m.selectedProposal.Subtitle)
				m.regPromptInput.SetValue(m.selectedProposal.Prompt)
				m.regFormFocus = 0
				m.message = ""
				return m, nil
			case "✍️ Cambiar solo título/subtítulo":
				m.state = stateUpdateTitleSubtitle
				m.updateTitleInput.SetValue(m.selectedProposal.Title)
				m.updateSubtitleInput.SetValue(m.selectedProposal.Subtitle)
				m.updateFormFocus = 0
				m.message = ""
				return m, nil
			case "🌐 Ver HTML":
				m.state = stateLoading
				m.loadingMsg = "Descargando HTML..."
				return m, tea.Batch(m.loadingSpinner.Tick, m.downloadProposalFile("html"))
			case "🌐 Generar HTML":
				m.state = stateLoading
				m.loadingMsg = "Generando HTML..."
				return m, tea.Batch(m.loadingSpinner.Tick, m.generateHTML())
			case "🔁 Regenerar HTML":
				m.state = stateLoading
				m.loadingMsg = "Regenerando HTML..."
				return m, tea.Batch(m.loadingSpinner.Tick, m.regenerateHTMLAndPDF())
			case "📄 Ver PDF":
				m.state = stateLoading
				m.loadingMsg = "Descargando PDF..."
				return m, tea.Batch(m.loadingSpinner.Tick, m.downloadProposalFile("pdf"))
			case "📄 Generar PDF":
				m.state = statePDFMode
				m.pdfModeCursor = 0
				m.message = ""
				return m, nil
			case "🔁 Regenerar PDF":
				m.state = statePDFMode
				m.pdfModeCursor = 0
				m.message = ""
				return m, nil
			case "✏️ Modificar propuesta":
				m.state = stateModifyProposal
				m.modTitleInput.SetValue(m.selectedProposal.Title)
				m.modSubtitleInput.SetValue(m.selectedProposal.Subtitle)
				m.modPromptInput.SetValue(m.selectedProposal.Title + " modificada")
				m.modFormFocus = 0
				m.message = ""
				return m, nil
			case "📥 Descargar archivos":
				m.state = stateLoading
				m.loadingMsg = "Descargando todos los archivos..."
				return m, tea.Batch(m.loadingSpinner.Tick, m.downloadAllFiles())
			case "🏠 Volver al Menú Principal":
				m.state = stateUnified
				m.showManagement = false
				m.selectedProposal = nil
				m.activeSection = 0
				m.message = ""
				return m, nil
			}
		case "esc":
			m.state = stateUnified
			m.showManagement = false
			m.selectedProposal = nil
			// Si hay propuestas activas, volver a esa sección, sino al menú principal
			if m.showProposals || m.showNewForm {
				m.activeSection = 1
			} else {
				m.activeSection = 0
			}
			m.message = ""
			return m, nil
		}
	}
	return m, nil
}

func (m appModel) viewProposalManagement() string {
	if m.selectedProposal == nil {
		return "Error: No hay propuesta seleccionada"
	}
	
	// CORRECCIÓN: Si el menú está vacío, actualizarlo
	if len(m.mgmtMenuItems) == 0 {
		// No podemos modificar el modelo desde View, pero podemos loguear el error
		// El problema real está en que updateManagementMenu no se está llamando correctamente
		// o el modelo no está siendo actualizado por referencia
	}
	
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("📋 Gestionar: %s", m.selectedProposal.Title)))
	b.WriteString("\n\n")
	b.WriteString(normalStyle.Render(fmt.Sprintf("ID: %s\nSubtítulo: %s\n", m.selectedProposal.ID, m.selectedProposal.Subtitle)))
	b.WriteString("\n")
	b.WriteString("Selecciona una acción:")
	b.WriteString("\n\n")
	
	if len(m.mgmtMenuItems) == 0 {
		b.WriteString(errorStyle.Render("No hay opciones disponibles. Por favor, presiona Esc para volver."))
		b.WriteString("\n\n")
		b.WriteString(normalStyle.Render("Esc para volver"))
		return docStyle.Render(b.String())
	}
	
	for i, item := range m.mgmtMenuItems {
		cursor := " "
		if m.mgmtMenuCursor == i {
			cursor = ">"
			// Agregar espacio después del emoji
			formattedItem := addSpaceAfterEmoji(item)
			b.WriteString(selectedStyle.Render(fmt.Sprintf("%s %s", cursor, formattedItem)))
		} else {
			// Agregar espacio después del emoji
			formattedItem := addSpaceAfterEmoji(item)
			b.WriteString(normalStyle.Render(fmt.Sprintf("%s %s", cursor, formattedItem)))
		}
		b.WriteString("\n")
	}
	
	b.WriteString("\n")
	b.WriteString(normalStyle.Render("↑↓ para navegar | Enter para seleccionar | Esc para volver"))
	
	if m.message != "" {
		b.WriteString("\n\n")
		if m.messageType == "error" {
			b.WriteString(errorStyle.Render(m.message))
		} else if m.messageType == "success" {
			b.WriteString(successStyle.Render(m.message))
		} else {
			b.WriteString(infoStyle.Render(m.message))
		}
	}
	
	return docStyle.Render(b.String())
}

func (m *appModel) updateManagementMenu() {
	if m.selectedProposal == nil {
		return
	}
	
	menuItems := []string{
		"📝 Ver propuesta (MD)",
		"🛠️ Regenerar (título/subtítulo/prompt)",
		"✍️ Cambiar solo título/subtítulo",
	}
	
	// Add HTML options
	if m.selectedProposal.HTMLURL != "" {
		menuItems = append(menuItems, "🌐 Ver HTML")
		menuItems = append(menuItems, "🔁 Regenerar HTML")
	} else {
		menuItems = append(menuItems, "🌐 Generar HTML")
	}
	
	// Add PDF options
	if m.selectedProposal.PDFURL != "" {
		menuItems = append(menuItems, "📄 Ver PDF")
		menuItems = append(menuItems, "🔁 Regenerar PDF")
	} else {
		menuItems = append(menuItems, "📄 Generar PDF")
	}
	
	menuItems = append(menuItems, "✏️ Modificar propuesta")
	menuItems = append(menuItems, "📥 Descargar archivos")
	menuItems = append(menuItems, "🏠 Volver al Menú Principal")
	
	m.mgmtMenuItems = menuItems
	m.mgmtMenuCursor = 0
}

// Regenerate form views and updates
func (m appModel) updateRegenerate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.regFormFocus = (m.regFormFocus + 1) % 3
			switch m.regFormFocus {
			case 0:
				m.regTitleInput.Focus()
				m.regSubtitleInput.Blur()
				m.regPromptInput.Blur()
				return m, nil
			case 1:
				m.regTitleInput.Blur()
				m.regSubtitleInput.Focus()
				m.regPromptInput.Blur()
				return m, nil
			case 2:
				m.regTitleInput.Blur()
				m.regSubtitleInput.Blur()
				m.regPromptInput.Focus()
				return m, nil
			}
		case "enter":
			if m.regFormFocus == 2 {
				// Regenerate
				title := strings.TrimSpace(m.regTitleInput.Value())
				subtitle := strings.TrimSpace(m.regSubtitleInput.Value())
				prompt := strings.TrimSpace(m.regPromptInput.Value())
				
				if prompt == "" {
					m.message = "El prompt no puede estar vacío"
					m.messageType = "error"
					return m, nil
				}
				
				if m.selectedProposal == nil {
					return m, nil
				}
				
				m.state = stateLoading
				m.loadingMsg = "Regenerando propuesta..."
				return m, tea.Batch(m.loadingSpinner.Tick, m.regenerateProposal(title, subtitle, prompt))
			}
		case "esc":
			m.state = stateUnified
			m.showManagement = true
			m.activeSection = 2
			m.message = ""
			return m, nil
		}
	}
	
	// Update focused input
	switch m.regFormFocus {
	case 0:
		m.regTitleInput, cmd = m.regTitleInput.Update(msg)
	case 1:
		m.regSubtitleInput, cmd = m.regSubtitleInput.Update(msg)
	case 2:
		m.regPromptInput, cmd = m.regPromptInput.Update(msg)
	}
	
	return m, cmd
}

func (m appModel) viewRegenerate() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("🛠️ Regenerar Propuesta"))
	b.WriteString("\n\n")
	b.WriteString("Edita los campos para regenerar el contenido (MD se reemplaza)")
	b.WriteString("\n\n")
	
	b.WriteString("Título:")
	b.WriteString("\n")
	if m.regFormFocus == 0 {
		b.WriteString(selectedStyle.Render(m.regTitleInput.View()))
	} else {
		b.WriteString(m.regTitleInput.View())
	}
	b.WriteString("\n\n")
	
	b.WriteString("Subtítulo:")
	b.WriteString("\n")
	if m.regFormFocus == 1 {
		b.WriteString(selectedStyle.Render(m.regSubtitleInput.View()))
	} else {
		b.WriteString(m.regSubtitleInput.View())
	}
	b.WriteString("\n\n")
	
	b.WriteString("Prompt:")
	b.WriteString("\n")
	if m.regFormFocus == 2 {
		b.WriteString(selectedStyle.Render(m.regPromptInput.View()))
	} else {
		b.WriteString(m.regPromptInput.View())
	}
	b.WriteString("\n\n")
	
	b.WriteString(normalStyle.Render("Tab para cambiar campo | Enter en prompt para regenerar | Esc para volver"))
	
	if m.message != "" {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.message))
	}
	
	return docStyle.Render(b.String())
}

// Update title/subtitle form views and updates
func (m appModel) updateUpdateTitleSubtitle(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.updateFormFocus = (m.updateFormFocus + 1) % 2
			switch m.updateFormFocus {
			case 0:
				m.updateTitleInput.Focus()
				m.updateSubtitleInput.Blur()
				return m, nil
			case 1:
				m.updateTitleInput.Blur()
				m.updateSubtitleInput.Focus()
				return m, nil
			}
		case "enter":
			if m.updateFormFocus == 1 {
				// Update
				title := strings.TrimSpace(m.updateTitleInput.Value())
				subtitle := strings.TrimSpace(m.updateSubtitleInput.Value())
				
				if m.selectedProposal == nil {
					return m, nil
				}
				
				m.state = stateLoading
				m.loadingMsg = "Actualizando título y subtítulo..."
				return m, tea.Batch(m.loadingSpinner.Tick, m.updateTitleSubtitle(title, subtitle))
			}
		case "esc":
			m.state = stateUnified
			m.showManagement = true
			m.activeSection = 2
			m.message = ""
			return m, nil
		}
	}
	
	// Update focused input
	switch m.updateFormFocus {
	case 0:
		m.updateTitleInput, cmd = m.updateTitleInput.Update(msg)
	case 1:
		m.updateSubtitleInput, cmd = m.updateSubtitleInput.Update(msg)
	}
	
	return m, cmd
}

func (m appModel) viewUpdateTitleSubtitle() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("✍️ Actualizar Título/Subtítulo"))
	b.WriteString("\n\n")
	
	b.WriteString("Título:")
	b.WriteString("\n")
	if m.updateFormFocus == 0 {
		b.WriteString(selectedStyle.Render(m.updateTitleInput.View()))
	} else {
		b.WriteString(m.updateTitleInput.View())
	}
	b.WriteString("\n\n")
	
	b.WriteString("Subtítulo:")
	b.WriteString("\n")
	if m.updateFormFocus == 1 {
		b.WriteString(selectedStyle.Render(m.updateSubtitleInput.View()))
	} else {
		b.WriteString(m.updateSubtitleInput.View())
	}
	b.WriteString("\n\n")
	
	b.WriteString(normalStyle.Render("Tab para cambiar campo | Enter en subtítulo para actualizar | Esc para volver"))
	
	if m.message != "" {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.message))
	}
	
	return docStyle.Render(b.String())
}

// Modify proposal form views and updates
func (m appModel) updateModifyProposal(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.modFormFocus = (m.modFormFocus + 1) % 3
			switch m.modFormFocus {
			case 0:
				m.modTitleInput.Focus()
				m.modSubtitleInput.Blur()
				m.modPromptInput.Blur()
				return m, nil
			case 1:
				m.modTitleInput.Blur()
				m.modSubtitleInput.Focus()
				m.modPromptInput.Blur()
				return m, nil
			case 2:
				m.modTitleInput.Blur()
				m.modSubtitleInput.Blur()
				m.modPromptInput.Focus()
				return m, nil
			}
		case "enter":
			if m.modFormFocus == 2 {
				// Modify
				title := strings.TrimSpace(m.modTitleInput.Value())
				subtitle := strings.TrimSpace(m.modSubtitleInput.Value())
				prompt := strings.TrimSpace(m.modPromptInput.Value())
				
				if prompt == "" {
					m.message = "El prompt no puede estar vacío"
					m.messageType = "error"
					return m, nil
				}
				
				if m.selectedProposal == nil {
					return m, nil
				}
				
				m.state = stateLoading
				m.loadingMsg = "Modificando propuesta..."
				return m, tea.Batch(m.loadingSpinner.Tick, m.modifyProposal(title, subtitle, prompt))
			}
		case "esc":
			m.state = stateUnified
			m.showManagement = true
			m.activeSection = 2
			m.message = ""
			return m, nil
		}
	}
	
	// Update focused input
	switch m.modFormFocus {
	case 0:
		m.modTitleInput, cmd = m.modTitleInput.Update(msg)
	case 1:
		m.modSubtitleInput, cmd = m.modSubtitleInput.Update(msg)
	case 2:
		m.modPromptInput, cmd = m.modPromptInput.Update(msg)
	}
	
	return m, cmd
}

func (m appModel) viewModifyProposal() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("✏️ Modificar Propuesta"))
	b.WriteString("\n\n")
	b.WriteString("Ingresa el nuevo prompt:")
	b.WriteString("\n\n")
	
	b.WriteString("Título:")
	b.WriteString("\n")
	if m.modFormFocus == 0 {
		b.WriteString(selectedStyle.Render(m.modTitleInput.View()))
	} else {
		b.WriteString(m.modTitleInput.View())
	}
	b.WriteString("\n\n")
	
	b.WriteString("Subtítulo:")
	b.WriteString("\n")
	if m.modFormFocus == 1 {
		b.WriteString(selectedStyle.Render(m.modSubtitleInput.View()))
	} else {
		b.WriteString(m.modSubtitleInput.View())
	}
	b.WriteString("\n\n")
	
	b.WriteString("Prompt:")
	b.WriteString("\n")
	if m.modFormFocus == 2 {
		b.WriteString(selectedStyle.Render(m.modPromptInput.View()))
	} else {
		b.WriteString(m.modPromptInput.View())
	}
	b.WriteString("\n\n")
	
	b.WriteString(normalStyle.Render("Tab para cambiar campo | Enter en prompt para modificar | Esc para volver"))
	
	if m.message != "" {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.message))
	}
	
	return docStyle.Render(b.String())
}

// PDF mode selection views and updates
func (m appModel) updatePDFMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.pdfModeCursor > 0 {
				m.pdfModeCursor--
			}
		case "down", "j":
			if m.pdfModeCursor < len(m.pdfModeItems)-1 {
				m.pdfModeCursor++
			}
		case "enter":
			if m.selectedProposal == nil {
				return m, nil
			}
			
			mode := m.pdfModeItems[m.pdfModeCursor]
			m.state = stateLoading
			m.loadingMsg = fmt.Sprintf("Generando PDF en modo %s...", mode)
			return m, tea.Batch(m.loadingSpinner.Tick, m.generatePDF(mode))
		case "esc":
			m.state = stateUnified
			m.showManagement = true
			m.activeSection = 2
			m.message = ""
			return m, nil
		}
	}
	return m, nil
}

func (m appModel) viewPDFMode() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Generar PDF"))
	b.WriteString("\n\n")
	b.WriteString("Selecciona el modo de impresión del PDF:")
	b.WriteString("\n\n")
	
	for i, item := range m.pdfModeItems {
		cursor := " "
		if m.pdfModeCursor == i {
			cursor = ">"
			b.WriteString(selectedStyle.Render(fmt.Sprintf("%s %s", cursor, item)))
		} else {
			b.WriteString(normalStyle.Render(fmt.Sprintf("%s %s", cursor, item)))
		}
		b.WriteString("\n")
	}
	
	b.WriteString("\n")
	b.WriteString(normalStyle.Render("↑↓ para navegar | Enter para seleccionar | Esc para volver"))
	
	if m.message != "" {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.message))
	}
	
	return docStyle.Render(b.String())
}

// Loading view
func (m appModel) viewLoading() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("⏳ Procesando"))
	b.WriteString("\n\n")
	b.WriteString(m.loadingSpinner.View())
	b.WriteString(" ")
	b.WriteString(m.loadingMsg)
	b.WriteString("\n\n")
	b.WriteString(normalStyle.Render("Por favor espera..."))
	return docStyle.Render(b.String())
}

// List item for proposals
type proposalItem struct {
	id       string
	title    string
	subtitle string
	created  string
}

func (i proposalItem) FilterValue() string {
	return i.title + " " + i.subtitle
}

func (i proposalItem) Title() string {
	return i.title
}

func (i proposalItem) Description() string {
	return fmt.Sprintf("ID: %s | %s | Creada: %s", i.id, i.subtitle, i.created)
}

func (m *appModel) updateProposalsListItems() {
	items := []list.Item{}
	for _, prop := range m.proposals {
		items = append(items, proposalItem{
			id:       prop.ID,
			title:    prop.Title,
			subtitle: prop.Subtitle,
			created:  prop.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	m.proposalsList.SetItems(items)
}

// Commands for async operations
func (m appModel) loadProposals() tea.Cmd {
	return func() tea.Msg {
		proposals, err := m.client.GetProposals()
		if err != nil {
			return errMsg(err)
		}
		return proposalsLoadedMsg(proposals)
	}
}

func (m appModel) createProposalAndDownloadMD() tea.Cmd {
	return func() tea.Msg {
		response, err := m.client.CreateProposal(m.genRequest)
		if err != nil {
			return errMsg(err)
		}
		
		homeDir, _ := os.UserHomeDir()
		downloadDir := filepath.Join(homeDir, "Downloads")
		os.MkdirAll(downloadDir, 0755)
		mdPath := filepath.Join(downloadDir, response.ID+".md")
		
		if err := m.client.DownloadProposalFile(response.ID, "md", mdPath); err != nil {
			return errMsg(err)
		}
		
		OpenDirectory(downloadDir)
		
		return operationCompletedMsg{
			message: "Propuesta creada y MD descargado en " + downloadDir,
			nextState: stateMainMenu,
		}
	}
}

func (m appModel) createProposalAndGenerateHTML() tea.Cmd {
	return func() tea.Msg {
		response, err := m.client.CreateProposal(m.genRequest)
		if err != nil {
			return errMsg(err)
		}
		
		_, err = m.client.GenerateHTML(response.ID)
		if err != nil {
			return errMsg(err)
		}
		
		_, err = m.client.GeneratePDF(response.ID, "normal")
		if err != nil {
			return errMsg(err)
		}
		
		return operationCompletedMsg{
			message: "Propuesta creada, HTML y PDF generados exitosamente",
			nextState: stateMainMenu,
		}
	}
}

func (m appModel) createProposalAndGeneratePDF() tea.Cmd {
	return func() tea.Msg {
		response, err := m.client.CreateProposal(m.genRequest)
		if err != nil {
			return errMsg(err)
		}
		
		_, err = m.client.GenerateHTML(response.ID)
		if err != nil {
			return errMsg(err)
		}
		
		pdfResponse, err := m.client.GeneratePDF(response.ID, "normal")
		if err != nil {
			return errMsg(err)
		}
		
		return operationCompletedMsg{
			message: "Propuesta creada, HTML y PDF generados. PDF: " + pdfResponse.PDFURL,
			nextState: stateMainMenu,
		}
	}
}

func (m appModel) downloadProposalFile(fileType string) tea.Cmd {
	return func() tea.Msg {
		if m.selectedProposal == nil {
			return errMsg(fmt.Errorf("no hay propuesta seleccionada"))
		}
		
		homeDir, _ := os.UserHomeDir()
		downloadDir := filepath.Join(homeDir, "Downloads")
		os.MkdirAll(downloadDir, 0755)
		filePath := filepath.Join(downloadDir, m.selectedProposal.ID+"."+fileType)
		
		if err := m.client.DownloadProposalFile(m.selectedProposal.ID, fileType, filePath); err != nil {
			return errMsg(err)
		}
		
		OpenDirectory(downloadDir)
		
		return operationCompletedMsg{
			message: fmt.Sprintf("%s descargado en %s", strings.ToUpper(fileType), downloadDir),
			nextState: stateProposalManagement,
		}
	}
}

func (m appModel) generateHTML() tea.Cmd {
	return func() tea.Msg {
		if m.selectedProposal == nil {
			return errMsg(fmt.Errorf("no hay propuesta seleccionada"))
		}
		
		response, err := m.client.GenerateHTML(m.selectedProposal.ID)
		if err != nil {
			return errMsg(err)
		}
		
		return htmlGeneratedMsg(response)
	}
}

type htmlPDFRegeneratedMsg struct {
	htmlURL string
	pdfURL  string
}

func (m appModel) regenerateHTMLAndPDF() tea.Cmd {
	return func() tea.Msg {
		if m.selectedProposal == nil {
			return errMsg(fmt.Errorf("no hay propuesta seleccionada"))
		}
		
		htmlResponse, err := m.client.GenerateHTML(m.selectedProposal.ID)
		if err != nil {
			return errMsg(err)
		}
		
		pdfResponse, err := m.client.GeneratePDF(m.selectedProposal.ID, "normal")
		if err != nil {
			return errMsg(err)
		}
		
		return htmlPDFRegeneratedMsg{
			htmlURL: htmlResponse.HTMLURL,
			pdfURL:  pdfResponse.PDFURL,
		}
	}
}

func (m appModel) generatePDF(mode string) tea.Cmd {
	return func() tea.Msg {
		if m.selectedProposal == nil {
			return errMsg(fmt.Errorf("no hay propuesta seleccionada"))
		}
		
		response, err := m.client.GeneratePDF(m.selectedProposal.ID, mode)
		if err != nil {
			return errMsg(err)
		}
		
		return pdfGeneratedMsg(response)
	}
}

type proposalRegeneratedMsg struct {
	title    string
	subtitle string
	prompt   string
	htmlURL  string
	pdfURL   string
}

func (m appModel) regenerateProposal(title, subtitle, prompt string) tea.Cmd {
	return func() tea.Msg {
		if m.selectedProposal == nil {
			return errMsg(fmt.Errorf("no hay propuesta seleccionada"))
		}
		
		err := m.client.RegenerateProposal(m.selectedProposal.ID, title, subtitle, prompt)
		if err != nil {
			return errMsg(err)
		}
		
		htmlResponse, err := m.client.GenerateHTML(m.selectedProposal.ID)
		if err != nil {
			return errMsg(err)
		}
		
		pdfResponse, err := m.client.GeneratePDF(m.selectedProposal.ID, "normal")
		if err != nil {
			return errMsg(err)
		}
		
		return proposalRegeneratedMsg{
			title:    title,
			subtitle: subtitle,
			prompt:   prompt,
			htmlURL:  htmlResponse.HTMLURL,
			pdfURL:   pdfResponse.PDFURL,
		}
	}
}

type titleSubtitleUpdatedMsg struct {
	title    string
	subtitle string
}

func (m appModel) updateTitleSubtitle(title, subtitle string) tea.Cmd {
	return func() tea.Msg {
		if m.selectedProposal == nil {
			return errMsg(fmt.Errorf("no hay propuesta seleccionada"))
		}
		
		err := m.client.UpdateTitleSubtitle(m.selectedProposal.ID, title, subtitle)
		if err != nil {
			return errMsg(err)
		}
		
		return titleSubtitleUpdatedMsg{
			title:    title,
			subtitle: subtitle,
		}
	}
}

type proposalModifiedMsg struct {
	htmlURL string
	pdfURL  string
}

func (m appModel) modifyProposal(title, subtitle, prompt string) tea.Cmd {
	return func() tea.Msg {
		if m.selectedProposal == nil {
			return errMsg(fmt.Errorf("no hay propuesta seleccionada"))
		}
		
		_, err := m.client.ModifyProposal(m.selectedProposal.ID, title, subtitle, prompt)
		if err != nil {
			return errMsg(err)
		}
		
		htmlResponse, err := m.client.GenerateHTML(m.selectedProposal.ID)
		if err != nil {
			return errMsg(err)
		}
		
		pdfResponse, err := m.client.GeneratePDF(m.selectedProposal.ID, "normal")
		if err != nil {
			return errMsg(err)
		}
		
		return proposalModifiedMsg{
			htmlURL: htmlResponse.HTMLURL,
			pdfURL:  pdfResponse.PDFURL,
		}
	}
}

func (m appModel) downloadAllFiles() tea.Cmd {
	return func() tea.Msg {
		if m.selectedProposal == nil {
			return errMsg(fmt.Errorf("no hay propuesta seleccionada"))
		}
		
		homeDir, _ := os.UserHomeDir()
		downloadDir := filepath.Join(homeDir, "Downloads")
		os.MkdirAll(downloadDir, 0755)
		
		// Download MD
		mdPath := filepath.Join(downloadDir, m.selectedProposal.ID+".md")
		m.client.DownloadProposalFile(m.selectedProposal.ID, "md", mdPath)
		
		// Download HTML
		htmlPath := filepath.Join(downloadDir, m.selectedProposal.ID+".html")
		m.client.DownloadProposalFile(m.selectedProposal.ID, "html", htmlPath)
		
		// Download PDF
		pdfPath := filepath.Join(downloadDir, m.selectedProposal.ID+".pdf")
		m.client.DownloadProposalFile(m.selectedProposal.ID, "pdf", pdfPath)
		
		OpenDirectory(downloadDir)
		
		return operationCompletedMsg{
			message: fmt.Sprintf("Archivos descargados en %s", downloadDir),
			nextState: stateProposalManagement,
		}
	}
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

// OpenDirectory opens a directory in the file manager
func OpenDirectory(path string) {
	var cmd *exec.Cmd

	switch {
	case isCommandAvailable("xdg-open"):
		cmd = exec.Command("xdg-open", path)
	case isCommandAvailable("open"):
		cmd = exec.Command("open", path)
	case isCommandAvailable("explorer"):
		cmd = exec.Command("explorer", path)
	default:
		return
	}

	// Start the command in background without waiting
	_ = cmd.Start()
}

// isCommandAvailable checks if a command is available in PATH
func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}
