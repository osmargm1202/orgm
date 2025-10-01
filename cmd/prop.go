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
	"time"

	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Proposal represents a proposal from the API
type Proposal struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Subtitle  string    `json:"subtitle"`
	CreatedAt time.Time `json:"created_at"`
	MDURL     string    `json:"md_url,omitempty"`
	HTMLURL   string    `json:"html_url,omitempty"`
	PDFURL    string    `json:"pdf_url,omitempty"`
	SizeHTML  int       `json:"size_html"`
	SizePDF   int       `json:"size_pdf"`
}

// UnmarshalJSON custom unmarshaler for Proposal to handle date parsing
func (p *Proposal) UnmarshalJSON(data []byte) error {
	type Alias Proposal
	aux := &struct {
		CreatedAt string `json:"created_at"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	// Try different date formats
	dateFormats := []string{
		"2006-01-02T15:04:05.999999", // API format without timezone (6 digits)
		"2006-01-02T15:04:05.999",    // API format without timezone (3 digits)
		"2006-01-02T15:04:05.99",     // API format without timezone (2 digits)
		"2006-01-02T15:04:05.9",      // API format without timezone (1 digit)
		"2006-01-02T15:04:05",        // API format without timezone (no decimal)
		"2006-01-02T15:04:05.000000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}
	
	for _, format := range dateFormats {
		if t, err := time.Parse(format, aux.CreatedAt); err == nil {
			p.CreatedAt = t
			return nil
		}
	}
	
	// If all formats fail, set to zero time
	p.CreatedAt = time.Time{}
	return nil
}

// TextGenerationRequest represents the request for text generation
type TextGenerationRequest struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Prompt   string `json:"prompt"`
	Model    string `json:"model,omitempty"`
}

// TextGenerationResponse represents the response from text generation
type TextGenerationResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	MDURL     string    `json:"md_url,omitempty"`
}

// UnmarshalJSON custom unmarshaler for TextGenerationResponse to handle date parsing
func (t *TextGenerationResponse) UnmarshalJSON(data []byte) error {
	type Alias TextGenerationResponse
	aux := &struct {
		CreatedAt string `json:"created_at"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	// Try different date formats
	dateFormats := []string{
		"2006-01-02T15:04:05.999999", // API format without timezone (6 digits)
		"2006-01-02T15:04:05.999",    // API format without timezone (3 digits)
		"2006-01-02T15:04:05.99",     // API format without timezone (2 digits)
		"2006-01-02T15:04:05.9",      // API format without timezone (1 digit)
		"2006-01-02T15:04:05",        // API format without timezone (no decimal)
		"2006-01-02T15:04:05.000000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}
	
	for _, format := range dateFormats {
		if parsedTime, err := time.Parse(format, aux.CreatedAt); err == nil {
			t.CreatedAt = parsedTime
			return nil
		}
	}
	
	// If all formats fail, set to zero time
	t.CreatedAt = time.Time{}
	return nil
}

// HTMLGenerationRequest represents the request for HTML generation
type HTMLGenerationRequest struct {
	ProposalID string `json:"proposal_id"`
	Model      string `json:"model,omitempty"`
}

// HTMLGenerationResponse represents the response from HTML generation
type HTMLGenerationResponse struct {
	ID      string `json:"id"`
	HTMLURL string `json:"html_url,omitempty"`
	MDURL   string `json:"md_url,omitempty"`
}

// PDFGenerationRequest represents the request for PDF generation
type PDFGenerationRequest struct {
	ProposalID string `json:"proposal_id"`
	Modo       string `json:"modo,omitempty"`
}

// PDFGenerationResponse represents the response from PDF generation
type PDFGenerationResponse struct {
	ID     string `json:"id"`
	PDFURL string `json:"pdf_url,omitempty"`
}

// ModifyProposalRequest represents the request for modifying a proposal
type ModifyProposalRequest struct {
	Prompt string `json:"prompt"`
}

// ModifyProposalResponse represents the response from modifying a proposal
type ModifyProposalResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	MDURL     string    `json:"md_url,omitempty"`
}

// UnmarshalJSON custom unmarshaler for ModifyProposalResponse to handle date parsing
func (m *ModifyProposalResponse) UnmarshalJSON(data []byte) error {
	type Alias ModifyProposalResponse
	aux := &struct {
		CreatedAt string `json:"created_at"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Try different date formats
	dateFormats := []string{
		"2006-01-02T15:04:05.999999", // API format without timezone (6 digits)
		"2006-01-02T15:04:05.999",    // API format without timezone (3 digits)
		"2006-01-02T15:04:05.99",     // API format without timezone (2 digits)
		"2006-01-02T15:04:05.9",      // API format without timezone (1 digit)
		"2006-01-02T15:04:05",        // API format without timezone (no decimal)
		"2006-01-02T15:04:05.000000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range dateFormats {
		if t, err := time.Parse(format, aux.CreatedAt); err == nil {
			m.CreatedAt = t
			return nil
		}
	}

	// If all formats fail, set to zero time
	m.CreatedAt = time.Time{}
	return nil
}


// PropCmd represents the prop command
var PropCmd = &cobra.Command{
	Use:   "prop",
	Short: "Gestión de propuestas con API",
	Long: `Comando para crear, modificar y gestionar propuestas usando la API de propuestas.

Subcomandos disponibles:
  new     Crear nueva propuesta con interfaz gráfica
  mod     Modificar propuesta existente con interfaz gráfica
  view    Ver y descargar propuestas con interfaz gráfica`,
	Run: func(cmd *cobra.Command, args []string) {
		baseURL, err := getBaseURL()
		if err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error: "+err.Error()))
			return
		}
		showMainProposalMenu(baseURL)
	},
}

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Crear nueva propuesta con interfaz gráfica",
	Long:  `Crea una nueva propuesta usando interfaz gráfica con yad`,
	Run: func(cmd *cobra.Command, args []string) {
		baseURL, err := getBaseURL()
		if err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error: "+err.Error()))
			return
		}
		createNewProposalGUI(baseURL)
	},
}

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Crear nueva propuesta",
	Long:  `Crea una nueva propuesta usando la API`,
	Run: func(cmd *cobra.Command, args []string) {
		baseURL, err := getBaseURL()
		if err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error: "+err.Error()))
			return
		}
		createNewProposal(baseURL)
	},
}

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Descargar archivos de propuesta",
	Long:  `Descarga los archivos MD, HTML y PDF de una propuesta`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		baseURL, err := getBaseURL()
		if err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error: "+err.Error()))
			return
		}
		downloadProposalByID(baseURL, args[0])
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



// htmlCmd represents the html command
var htmlCmd = &cobra.Command{
	Use:   "html <id>",
	Short: "Generar HTML para propuesta",
	Long:  `Genera HTML para una propuesta existente y descarga los archivos`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		baseURL, err := getBaseURL()
		if err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error: "+err.Error()))
			return
		}
		generateHTMLAndDownload(baseURL, args[0])
	},
}

// pdfCmd represents the pdf command
var pdfCmd = &cobra.Command{
	Use:   "pdf <id>",
	Short: "Generar PDF para propuesta",
	Long:  `Genera PDF para una propuesta existente y descarga los archivos`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		baseURL, err := getBaseURL()
		if err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error: "+err.Error()))
			return
		}
		generatePDFAndDownload(baseURL, args[0])
	},
}

func init() {
	// Add subcommands to prop
	PropCmd.AddCommand(newCmd)
	PropCmd.AddCommand(installCmd)
}

func getBaseURL() (string, error) {
	baseURL := viper.GetString("url.propuestas_api")
	if baseURL == "" {
		return "", fmt.Errorf("No se encontró la URL de la API de propuestas en links.toml")
	}
	return strings.TrimSuffix(baseURL, "/"), nil
}

func createNewProposal(baseURL string) {
	fmt.Printf("%s\n", inputs.TitleStyle.Render("Crear Nueva Propuesta"))
	fmt.Println()

	// Get prompt using textarea
	prompt := inputs.GetTextArea("Ingresa el prompt para la propuesta:", "Escribe aquí tu prompt...")
	if prompt == "" {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error: El prompt no puede estar vacío"))
		return
	}

	// Get title
	title := inputs.GetInput("Título de la propuesta:")
	if title == "" {
		title = "Propuesta de Servicios"
	}

	// Get subtitle
	subtitle := inputs.GetInput("Subtítulo de la propuesta:")
	if subtitle == "" {
		subtitle = "Sistema de Gestión"
	}

	// Create request
	request := TextGenerationRequest{
		Title:    title,
		Subtitle: subtitle,
		Prompt:   prompt,
		Model:    "gpt-5-chat-latest", // Default model
	}

	// Send request
	fmt.Printf("%s\n", inputs.InfoStyle.Render("Enviando solicitud a la API..."))

	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al serializar la solicitud: "+err.Error()))
		return
	}

	req, err := http.NewRequest("POST", baseURL+"/generate-text", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al crear la solicitud: "+err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al enviar la solicitud: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", inputs.ErrorStyle.Render(fmt.Sprintf("Error del servidor (%d): %s", resp.StatusCode, string(body))))
		return
	}

	var response TextGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al decodificar la respuesta: "+err.Error()))
		return
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("¡Propuesta creada exitosamente!"))
	fmt.Printf("ID: %s\n", response.ID)
	fmt.Printf("Creada: %s\n", response.CreatedAt.Format("2006-01-02 15:04:05"))
	if response.MDURL != "" {
		fmt.Printf("MD URL: %s\n", response.MDURL)
	}

	// Automatically download the generated files
	fmt.Println()
	downloadProposalByID(baseURL, response.ID)
}


func modifySpecificProposal(baseURL string, proposal Proposal) {
	fmt.Printf("%s\n", inputs.TitleStyle.Render("Modificar Propuesta: " + proposal.Title))
	fmt.Printf("ID: %s\n", proposal.ID)
	fmt.Printf("Creada: %s\n", proposal.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Get new title (use existing as default)
	title := inputs.GetInputWithDefault("Título (presiona Enter para mantener actual):", proposal.Title)
	if title == "" {
		title = proposal.Title
	}

	// Get new subtitle (use existing as default)
	subtitle := inputs.GetInputWithDefault("Subtítulo (presiona Enter para mantener actual):", proposal.Subtitle)
	if subtitle == "" {
		subtitle = proposal.Subtitle
	}

	// Get new prompt
	prompt := inputs.GetTextArea("Ingresa el nuevo prompt:", "Escribe aquí tu prompt...")
	if prompt == "" {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error: El prompt no puede estar vacío"))
		return
	}

	// Create request
	request := TextGenerationRequest{
		Title:    title,
		Subtitle: subtitle,
		Prompt:   prompt,
		Model:    "gpt-5-chat-latest", // Default model
	}

	// Send request
	fmt.Printf("%s\n", inputs.InfoStyle.Render("Enviando solicitud de modificación..."))

	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al serializar la solicitud: "+err.Error()))
		return
	}

	req, err := http.NewRequest("PUT", baseURL+"/proposal/"+proposal.ID, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al crear la solicitud: "+err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al enviar la solicitud: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", inputs.ErrorStyle.Render(fmt.Sprintf("Error del servidor (%d): %s", resp.StatusCode, string(body))))
		return
	}

	var response ModifyProposalResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al decodificar la respuesta: "+err.Error()))
		return
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("¡Propuesta modificada exitosamente!"))
	fmt.Printf("ID: %s\n", response.ID)
	fmt.Printf("Modificada: %s\n", response.CreatedAt.Format("2006-01-02 15:04:05"))
	if response.MDURL != "" {
		fmt.Printf("MD URL: %s\n", response.MDURL)
	}

	// Download the modified MD file only
	fmt.Println()
	downloadMDOnly(baseURL, response.ID)
}

func generateHTMLAndDownload(baseURL, proposalID string) {
	fmt.Printf("%s\n", inputs.TitleStyle.Render("Generar HTML para Propuesta: " + proposalID))
	fmt.Println()

	// Generate HTML
	fmt.Printf("%s\n", inputs.InfoStyle.Render("Generando HTML..."))

	htmlRequest := HTMLGenerationRequest{ProposalID: proposalID}
	jsonData, err := json.Marshal(htmlRequest)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al serializar solicitud HTML: "+err.Error()))
		return
	}

	req, err := http.NewRequest("POST", baseURL+"/generate-html", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al crear solicitud HTML: "+err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al generar HTML: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", inputs.ErrorStyle.Render(fmt.Sprintf("Error del servidor (%d): %s", resp.StatusCode, string(body))))
		return
	}

	var htmlResponse HTMLGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&htmlResponse); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al decodificar respuesta HTML: "+err.Error()))
		return
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ HTML generado: " + htmlResponse.HTMLURL))

	// Download the files
	downloadProposalByID(baseURL, proposalID)
}

func generatePDFAndDownload(baseURL, proposalID string) {
	fmt.Printf("%s\n", inputs.TitleStyle.Render("Generar PDF para Propuesta: " + proposalID))
	fmt.Println()

	// Generate PDF
	fmt.Printf("%s\n", inputs.InfoStyle.Render("Generando PDF..."))

	pdfRequest := PDFGenerationRequest{ProposalID: proposalID}
	jsonData, err := json.Marshal(pdfRequest)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al serializar solicitud PDF: "+err.Error()))
		return
	}

	req, err := http.NewRequest("POST", baseURL+"/generate-pdf", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al crear solicitud PDF: "+err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al generar PDF: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", inputs.ErrorStyle.Render(fmt.Sprintf("Error del servidor (%d): %s", resp.StatusCode, string(body))))
		return
	}

	var pdfResponse PDFGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&pdfResponse); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al decodificar respuesta PDF: "+err.Error()))
		return
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ PDF generado: " + pdfResponse.PDFURL))

	// Download the files
	downloadProposalByID(baseURL, proposalID)
}


func downloadSpecificProposal(baseURL string, proposal Proposal) {
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
	if err := downloadProposalFile(baseURL, proposal.ID, "md", filepath.Join(downloadDir, proposal.ID+".md")); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al descargar MD: "+err.Error()))
	} else {
		fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ MD descargado"))
	}

	// Download HTML file (only if HTML URL exists)
	if proposal.HTMLURL != "" {
		if err := downloadProposalFile(baseURL, proposal.ID, "html", filepath.Join(downloadDir, proposal.ID+".html")); err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al descargar HTML: "+err.Error()))
		} else {
			fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ HTML descargado"))
		}
	}

	// Download PDF file (only if PDF URL exists)
	if proposal.PDFURL != "" {
		if err := downloadProposalFile(baseURL, proposal.ID, "pdf", filepath.Join(downloadDir, proposal.ID+".pdf")); err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al descargar PDF: "+err.Error()))
		} else {
			fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ PDF descargado"))
		}
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("Descarga completada en: " + downloadDir))

	// Open download directory
	openDirectory(downloadDir)
}

func downloadProposalByID(baseURL, proposalID string) {
	// First get the proposal details
	resp, err := http.Get(baseURL + "/proposals")
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error de conexión a la API: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", inputs.ErrorStyle.Render(fmt.Sprintf("Error del servidor (HTTP %d): %s", resp.StatusCode, string(body))))
		return
	}

	var proposals []Proposal
	if err := json.NewDecoder(resp.Body).Decode(&proposals); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al decodificar respuesta: "+err.Error()))
		return
	}

	// Find the proposal by ID
	var foundProposal *Proposal
	for _, prop := range proposals {
		if prop.ID == proposalID {
			foundProposal = &prop
			break
		}
	}

	if foundProposal == nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Propuesta con ID "+proposalID+" no encontrada"))
		return
	}

	downloadSpecificProposal(baseURL, *foundProposal)
}

func downloadMDOnly(baseURL, proposalID string) {
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

	fmt.Printf("%s\n", inputs.InfoStyle.Render("Descargando archivo MD..."))

	// Download only MD file
	filepath := filepath.Join(downloadDir, proposalID+".md")
	if err := downloadProposalFile(baseURL, proposalID, "md", filepath); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al descargar MD: "+err.Error()))
	} else {
		fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ MD descargado"))
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("Descarga completada en: " + downloadDir))

	// Open download directory
	openDirectory(downloadDir)
}

func modifyProposalByID(baseURL, proposalID string) {
	// First get the proposal details
	resp, err := http.Get(baseURL + "/proposals")
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error de conexión a la API: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", inputs.ErrorStyle.Render(fmt.Sprintf("Error del servidor (HTTP %d): %s", resp.StatusCode, string(body))))
		return
	}

	var proposals []Proposal
	if err := json.NewDecoder(resp.Body).Decode(&proposals); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al decodificar respuesta: "+err.Error()))
		return
	}

	// Find the proposal by ID
	var foundProposal *Proposal
	for _, prop := range proposals {
		if prop.ID == proposalID {
			foundProposal = &prop
			break
		}
	}

	if foundProposal == nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Propuesta con ID "+proposalID+" no encontrada"))
		return
	}

	modifySpecificProposal(baseURL, *foundProposal)
}

func downloadProposalFile(baseURL, proposalID, fileType, filepath string) error {
	url := fmt.Sprintf("%s/proposal/%s/%s", baseURL, proposalID, fileType)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
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

func generateHTMLAndPDF(baseURL, proposalID string) {
	// Generate HTML
	fmt.Printf("%s\n", inputs.InfoStyle.Render("Generando HTML..."))
	
	htmlRequest := HTMLGenerationRequest{ProposalID: proposalID}
	jsonData, err := json.Marshal(htmlRequest)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al serializar solicitud HTML: "+err.Error()))
		return
	}

	req, err := http.NewRequest("POST", baseURL+"/generate-html", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al crear solicitud HTML: "+err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al generar HTML: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var htmlResponse HTMLGenerationResponse
		if err := json.NewDecoder(resp.Body).Decode(&htmlResponse); err == nil {
			fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ HTML generado: " + htmlResponse.HTMLURL))
		}
	}

	// Generate PDF
	fmt.Printf("%s\n", inputs.InfoStyle.Render("Generando PDF..."))
	
	pdfRequest := PDFGenerationRequest{ProposalID: proposalID}
	jsonData, err = json.Marshal(pdfRequest)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al serializar solicitud PDF: "+err.Error()))
		return
	}

	req, err = http.NewRequest("POST", baseURL+"/generate-pdf", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al crear solicitud PDF: "+err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al generar PDF: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var pdfResponse PDFGenerationResponse
		if err := json.NewDecoder(resp.Body).Decode(&pdfResponse); err == nil {
			fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ PDF generado: " + pdfResponse.PDFURL))
		}
	}
}

func getProposals(baseURL string) ([]Proposal, error) {
	resp, err := http.Get(baseURL + "/proposals")
	if err != nil {
		return nil, fmt.Errorf("Error de conexión a la API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Error del servidor (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var proposals []Proposal
	if err := json.NewDecoder(resp.Body).Decode(&proposals); err != nil {
		return nil, fmt.Errorf("Error al decodificar respuesta: %v", err)
	}

	return proposals, nil
}


// showMainProposalMenu shows the main unified proposal menu
func showMainProposalMenu(baseURL string) {
	for {
		// Show main menu with only two options
		cmd := exec.Command("yad",
			"--form",
			"--title=📋 Gestor de Propuestas",
			"--text=Selecciona una opción:",
			"--field=Opción:CB", "🆕 Nueva Propuesta!📁 Propuesta Existente",
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

		// Parse action selection
		action := strings.TrimSpace(strings.Split(result, "|")[0])
		if action == "" {
			return
		}

		// Handle selected action
		switch action {
		case "🆕 Nueva Propuesta":
			createNewProposalFlow(baseURL)
		case "📁 Propuesta Existente":
			showExistingProposalFlow(baseURL)
		}
	}
}

// createNewProposalFlow handles the complete flow for creating new proposals
func createNewProposalFlow(baseURL string) {
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
	request := TextGenerationRequest{
		Title:    title,
		Subtitle: subtitle,
		Prompt:   prompt,
		Model:    "gpt-5-chat-latest",
	}

	// Show generation menu
	showGenerationMenu(baseURL, request)
}

// showGenerationMenu shows options for generating documents after creating proposal
func showGenerationMenu(baseURL string, request TextGenerationRequest) {
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
			// Create proposal and show content
			proposal, err := createProposalFromRequest(baseURL, request)
			if err != nil {
				fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creando propuesta: "+err.Error()))
				continue
			}
			showProposalContent(baseURL, proposal)
		case "🌐 Generar HTML":
			// Create proposal and generate HTML
			proposal, err := createProposalFromRequest(baseURL, request)
			if err != nil {
				fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creando propuesta: "+err.Error()))
				continue
			}
			generateHTMLGUI(baseURL, proposal.ID)
		case "📄 Generar PDF":
			// Create proposal and generate PDF
			proposal, err := createProposalFromRequest(baseURL, request)
			if err != nil {
				fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creando propuesta: "+err.Error()))
				continue
			}
			generatePDFGUI(baseURL, proposal.ID)
		case "🏠 Volver al Menú Principal":
			return
		}
	}
}

// createProposalFromRequest creates a new proposal from a TextGenerationRequest
func createProposalFromRequest(baseURL string, request TextGenerationRequest) (Proposal, error) {
	// Convert request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return Proposal{}, fmt.Errorf("error marshaling request: %v", err)
	}

	// Make API call
	resp, err := http.Post(baseURL+"/generate-text", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return Proposal{}, fmt.Errorf("error calling API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return Proposal{}, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var response TextGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return Proposal{}, fmt.Errorf("error decoding response: %v", err)
	}

	// Convert to Proposal struct
	proposal := Proposal{
		ID:        response.ID,
		Title:     request.Title,
		Subtitle:  request.Subtitle,
		CreatedAt: response.CreatedAt,
		MDURL:     response.MDURL,
		HTMLURL:   "", // Not generated yet
		PDFURL:    "", // Not generated yet
		SizeHTML:  0,
		SizePDF:   0,
	}

	return proposal, nil
}

// showExistingProposalFlow handles the complete flow for existing proposals
func showExistingProposalFlow(baseURL string) {
	// Get proposals from API
	proposals, err := getProposals(baseURL)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error obteniendo propuestas: "+err.Error()))
		return
	}

	if len(proposals) == 0 {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("No hay propuestas disponibles"))
		return
	}

	// Create yad list dialog
	cmd := exec.Command("yad", 
		"--list",
		"--title=📁 Propuestas Existentes",
		"--text=Selecciona una propuesta:",
		"--column=ID",
		"--column=Título",
		"--column=Subtítulo", 
		"--column=Creada",
		"--width=800",
		"--height=400",
		"--search-column=2", // Search in title column
		"--print-column=1",   // Return only ID
		"--single-click",
		"--separator=|")

	// Add proposal items as arguments
	for _, prop := range proposals {
		// Add each field as separate argument
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
	// Remove separator if present (yad sometimes includes it)
	selectedID = strings.TrimSuffix(selectedID, "|")
	
	if selectedID == "" {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("No se seleccionó ninguna propuesta"))
		return
	}

	// Find selected proposal
	var selectedProposal *Proposal
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
	showProposalManagementLoop(baseURL, *selectedProposal)
}

// showProposalManagementLoop shows the management menu for a specific proposal in a loop
func showProposalManagementLoop(baseURL string, proposal Proposal) {
	for {
		// Create menu options based on available documents
		menuItems := []string{
			"📝 Ver propuesta (MD)",
		}
		
		// Add conditional buttons based on document availability
		if proposal.HTMLURL != "" {
			menuItems = append(menuItems, "🌐 Ver HTML")
		} else {
			menuItems = append(menuItems, "🌐 Generar HTML")
		}
		
		if proposal.PDFURL != "" {
			menuItems = append(menuItems, "📄 Ver PDF")
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
		for _, item := range menuItems {
			cmd.Args = append(cmd.Args, item)
		}

		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error ejecutando yad: "+err.Error()))
			return
		}

		selectedAction := strings.TrimSpace(string(output))
		if selectedAction == "" {
			return
		}

		// Handle selected action
		switch selectedAction {
		case "📝 Ver propuesta (MD)":
			showProposalContent(baseURL, proposal)
		case "🌐 Ver HTML":
			// Open HTML file in browser
			if proposal.HTMLURL != "" {
				downloadPath := getDownloadPath(proposal.ID + ".html")
				if err := downloadProposalFile(baseURL, proposal.ID, "html", downloadPath); err == nil {
					openFile(downloadPath)
				}
			}
		case "🌐 Generar HTML":
			generateHTMLGUI(baseURL, proposal.ID)
		case "📄 Ver PDF":
			// Open PDF file
			if proposal.PDFURL != "" {
				downloadPath := getDownloadPath(proposal.ID + ".pdf")
				if err := downloadProposalFile(baseURL, proposal.ID, "pdf", downloadPath); err == nil {
					openFile(downloadPath)
				}
			}
		case "📄 Generar PDF":
			generatePDFGUI(baseURL, proposal.ID)
		case "✏️ Modificar propuesta":
			modifyProposalGUI(baseURL, proposal)
		case "📥 Descargar archivos":
			downloadSpecificProposal(baseURL, proposal)
		case "🏠 Volver al Menú Principal":
			return
		}
	}
}

// showHTMLGenerator shows a list of proposals to generate HTML for
func showHTMLGenerator(baseURL string) {
	// Get proposals from API
	proposals, err := getProposals(baseURL)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error obteniendo propuestas: "+err.Error()))
		return
	}

	if len(proposals) == 0 {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("No hay propuestas disponibles"))
		return
	}

	// Create yad list dialog
	cmd := exec.Command("yad", 
		"--list",
		"--title=Generar HTML",
		"--text=Selecciona una propuesta para generar HTML:",
		"--column=ID",
		"--column=Título",
		"--column=Subtítulo", 
		"--column=Creada",
		"--width=800",
		"--height=400",
		"--search-column=2", // Search in title column
		"--print-column=1",   // Return only ID
		"--single-click",
		"--separator=|")

	// Add proposal items as arguments
	for _, prop := range proposals {
		// Add each field as separate argument
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
	// Remove separator if present (yad sometimes includes it)
	selectedID = strings.TrimSuffix(selectedID, "|")
	
	if selectedID == "" {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("No se seleccionó ninguna propuesta"))
		return
	}

	// Generate HTML for selected proposal
	generateHTMLGUI(baseURL, selectedID)
}

// showPDFGenerator shows a list of proposals to generate PDF for
func showPDFGenerator(baseURL string) {
	// Get proposals from API
	proposals, err := getProposals(baseURL)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error obteniendo propuestas: "+err.Error()))
		return
	}

	if len(proposals) == 0 {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("No hay propuestas disponibles"))
		return
	}

	// Create yad list dialog
	cmd := exec.Command("yad", 
		"--list",
		"--title=Generar PDF",
		"--text=Selecciona una propuesta para generar PDF:",
		"--column=ID",
		"--column=Título",
		"--column=Subtítulo", 
		"--column=Creada",
		"--width=800",
		"--height=400",
		"--search-column=2", // Search in title column
		"--print-column=1",   // Return only ID
		"--single-click",
		"--separator=|")

	// Add proposal items as arguments
	for _, prop := range proposals {
		// Add each field as separate argument
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
	// Remove separator if present (yad sometimes includes it)
	selectedID = strings.TrimSuffix(selectedID, "|")
	
	if selectedID == "" {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("No se seleccionó ninguna propuesta"))
		return
	}

	// Generate PDF for selected proposal
	generatePDFGUI(baseURL, selectedID)
}

// showProposalDownloader shows a list of proposals to download
func showProposalDownloader(baseURL string) {
	// Get proposals from API
	proposals, err := getProposals(baseURL)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error obteniendo propuestas: "+err.Error()))
		return
	}

	if len(proposals) == 0 {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("No hay propuestas disponibles"))
		return
	}

	// Create yad list dialog
	cmd := exec.Command("yad", 
		"--list",
		"--title=Descargar Archivos",
		"--text=Selecciona una propuesta para descargar:",
		"--column=ID",
		"--column=Título",
		"--column=Subtítulo", 
		"--column=Creada",
		"--width=800",
		"--height=400",
		"--search-column=2", // Search in title column
		"--print-column=1",   // Return only ID
		"--single-click",
		"--separator=|")

	// Add proposal items as arguments
	for _, prop := range proposals {
		// Add each field as separate argument
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
	// Remove separator if present (yad sometimes includes it)
	selectedID = strings.TrimSuffix(selectedID, "|")
	
	if selectedID == "" {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("No se seleccionó ninguna propuesta"))
		return
	}

	// Find selected proposal
	var selectedProposal *Proposal
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

	// Download all available files
	downloadSpecificProposal(baseURL, *selectedProposal)
}

// openDownloadsFolder opens the downloads folder
func openDownloadsFolder() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al obtener directorio home: "+err.Error()))
		return
	}
	downloadDir := filepath.Join(homeDir, "Downloads")
	openDirectory(downloadDir)
}

// showProposalManager shows the main proposal manager interface with yad
func showProposalManager(baseURL string) {
	// Get proposals from API
	proposals, err := getProposals(baseURL)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error obteniendo propuestas: "+err.Error()))
		return
	}

	if len(proposals) == 0 {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("No hay propuestas disponibles"))
		return
	}

	// Create yad list dialog
	cmd := exec.Command("yad", 
		"--list",
		"--title=Gestor de Propuestas",
		"--text=Selecciona una propuesta:",
		"--column=ID",
		"--column=Título",
		"--column=Subtítulo", 
		"--column=Creada",
		"--width=800",
		"--height=400",
		"--search-column=2", // Search in title column
		"--print-column=1",   // Return only ID
		"--single-click",
		"--separator=|")

	// Add proposal items as arguments
	for _, prop := range proposals {
		// Add each field as separate argument
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
	// Remove separator if present (yad sometimes includes it)
	selectedID = strings.TrimSuffix(selectedID, "|")
	
	if selectedID == "" {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("No se seleccionó ninguna propuesta"))
		return
	}

	// Find selected proposal
	var selectedProposal *Proposal
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

	// Show proposal management menu
	showProposalMenu(baseURL, *selectedProposal)
}

// showProposalMenu shows the menu for managing a specific proposal
func showProposalMenu(baseURL string, proposal Proposal) {
	// Create menu options based on available documents
	menuItems := []string{
		"Modificar propuesta",
		"Ver propuesta (MD)",
	}
	
	// Add conditional buttons based on document availability
	if proposal.HTMLURL != "" {
		menuItems = append(menuItems, "Ver HTML")
	} else {
		menuItems = append(menuItems, "Generar HTML")
	}
	
	if proposal.PDFURL != "" {
		menuItems = append(menuItems, "Ver PDF")
	} else {
		menuItems = append(menuItems, "Generar PDF")
	}
	
	menuItems = append(menuItems, "Descargar archivos")
	menuItems = append(menuItems, "Volver al listado")

	// Show yad menu dialog
	cmd := exec.Command("yad", 
		"--list",
		"--title=Gestionar Propuesta: "+proposal.Title,
		"--text=Selecciona una acción:",
		"--column=Acción",
		"--width=400",
		"--height=300",
		"--print-column=1",
		"--single-click",
		"--separator=|")

	// Add menu items
	for _, item := range menuItems {
		cmd.Args = append(cmd.Args, item)
	}

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error ejecutando yad: "+err.Error()))
		return
	}

	selectedAction := strings.TrimSpace(string(output))
	if selectedAction == "" {
		return
	}

	// Handle selected action
	switch selectedAction {
	case "Modificar propuesta":
		modifyProposalGUI(baseURL, proposal)
	case "Ver propuesta (MD)":
		showProposalContent(baseURL, proposal)
	case "Ver HTML":
		// Open HTML file in browser
		if proposal.HTMLURL != "" {
			downloadPath := getDownloadPath(proposal.ID + ".html")
			if err := downloadProposalFile(baseURL, proposal.ID, "html", downloadPath); err == nil {
				openFile(downloadPath)
			}
		}
	case "Generar HTML":
		generateHTMLGUI(baseURL, proposal.ID)
	case "Ver PDF":
		// Open PDF file
		if proposal.PDFURL != "" {
			downloadPath := getDownloadPath(proposal.ID + ".pdf")
			if err := downloadProposalFile(baseURL, proposal.ID, "pdf", downloadPath); err == nil {
				openFile(downloadPath)
			}
		}
	case "Generar PDF":
		generatePDFGUI(baseURL, proposal.ID)
	case "Descargar archivos":
		downloadSpecificProposal(baseURL, proposal)
	case "Volver al listado":
		showProposalManager(baseURL)
	}
}

// modifyProposalGUI shows a dialog to modify a proposal
func modifyProposalGUI(baseURL string, proposal Proposal) {
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

	// Create request
	request := TextGenerationRequest{
		Title:    title,
		Subtitle: subtitle,
		Prompt:   prompt,
		Model:    "gpt-5-chat-latest",
	}

	// Send modification request
	fmt.Printf("%s\n", inputs.InfoStyle.Render("Modificando propuesta..."))

	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al serializar la solicitud: "+err.Error()))
		return
	}

	req, err := http.NewRequest("PUT", baseURL+"/proposal/"+proposal.ID, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al crear la solicitud: "+err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al enviar la solicitud: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", inputs.ErrorStyle.Render(fmt.Sprintf("Error del servidor (%d): %s", resp.StatusCode, string(body))))
		return
	}

	var response ModifyProposalResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al decodificar la respuesta: "+err.Error()))
		return
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("¡Propuesta modificada exitosamente!"))
	fmt.Printf("ID: %s\n", response.ID)

	// Show success dialog
	exec.Command("yad", "--info", "--title=Éxito", "--text=Propuesta modificada exitosamente").Run()
}

// showProposalContent shows the proposal content in a yad dialog
func showProposalContent(baseURL string, proposal Proposal) {
	// Download MD content
	url := fmt.Sprintf("%s/proposal/%s/md", baseURL, proposal.ID)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error descargando contenido: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", inputs.ErrorStyle.Render(fmt.Sprintf("Error del servidor (%d): %s", resp.StatusCode, string(body))))
		return
	}

	content, err := io.ReadAll(resp.Body)
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
		generateHTMLGUI(baseURL, proposal.ID)
	case 2: // Generate PDF
		generatePDFGUI(baseURL, proposal.ID)
	}
}

// generateHTMLGUI generates HTML and shows success message
func generateHTMLGUI(baseURL string, proposalID string) {
	fmt.Printf("%s\n", inputs.InfoStyle.Render("Generando HTML..."))

	htmlRequest := HTMLGenerationRequest{ProposalID: proposalID}
	jsonData, err := json.Marshal(htmlRequest)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al serializar solicitud HTML: "+err.Error()))
		return
	}

	req, err := http.NewRequest("POST", baseURL+"/generate-html", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al crear solicitud HTML: "+err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al generar HTML: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", inputs.ErrorStyle.Render(fmt.Sprintf("Error del servidor (%d): %s", resp.StatusCode, string(body))))
		return
	}

	var htmlResponse HTMLGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&htmlResponse); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al decodificar respuesta HTML: "+err.Error()))
		return
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ HTML generado exitosamente"))

	// Download HTML file
	downloadProposalFile(baseURL, proposalID, "html", getDownloadPath(proposalID+".html"))

	// Show success dialog
	exec.Command("yad", "--info", "--title=Éxito", "--text=HTML generado y descargado exitosamente").Run()
}

// generatePDFGUI generates PDF and opens it
func generatePDFGUI(baseURL string, proposalID string) {
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

	pdfRequest := PDFGenerationRequest{
		ProposalID: proposalID,
		Modo:       mode,
	}
	jsonData, err := json.Marshal(pdfRequest)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al serializar solicitud PDF: "+err.Error()))
		return
	}

	req, err := http.NewRequest("POST", baseURL+"/generate-pdf", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al crear solicitud PDF: "+err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al generar PDF: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", inputs.ErrorStyle.Render(fmt.Sprintf("Error del servidor (%d): %s", resp.StatusCode, string(body))))
		return
	}

	var pdfResponse PDFGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&pdfResponse); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al decodificar respuesta PDF: "+err.Error()))
		return
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ PDF generado exitosamente"))

	// Download PDF file
	filepath := getDownloadPath(proposalID + ".pdf")
	if err := downloadProposalFile(baseURL, proposalID, "pdf", filepath); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al descargar PDF: "+err.Error()))
		return
	}

	// Open PDF file
	openFile(filepath)
}

// createNewProposalGUI creates a new proposal using GUI
func createNewProposalGUI(baseURL string) {
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
	request := TextGenerationRequest{
		Title:    title,
		Subtitle: subtitle,
		Prompt:   prompt,
		Model:    "gpt-5-chat-latest",
	}

	// Send request
	fmt.Printf("%s\n", inputs.InfoStyle.Render("Creando propuesta..."))

	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al serializar la solicitud: "+err.Error()))
		return
	}

	req, err := http.NewRequest("POST", baseURL+"/generate-text", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al crear la solicitud: "+err.Error()))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al enviar la solicitud: "+err.Error()))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", inputs.ErrorStyle.Render(fmt.Sprintf("Error del servidor (%d): %s", resp.StatusCode, string(body))))
		return
	}

	var response TextGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al decodificar la respuesta: "+err.Error()))
		return
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("¡Propuesta creada exitosamente!"))
	fmt.Printf("ID: %s\n", response.ID)

	// Download MD file
	filepath := getDownloadPath(response.ID + ".md")
	if err := downloadProposalFile(baseURL, response.ID, "md", filepath); err != nil {
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
		tempProposal := Proposal{
			ID:        response.ID,
			Title:     title,
			Subtitle:  subtitle,
			CreatedAt: response.CreatedAt,
		}
		showProposalContent(baseURL, tempProposal)
	case 2: // Generate HTML
		generateHTMLGUI(baseURL, response.ID)
	case 3: // Generate PDF
		generatePDFGUI(baseURL, response.ID)
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

// showProposalViewer shows the proposal viewer interface for downloading and viewing
func showProposalViewer(baseURL string) {
	// Get proposals from API
	proposals, err := getProposals(baseURL)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error obteniendo propuestas: "+err.Error()))
		return
	}

	if len(proposals) == 0 {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("No hay propuestas disponibles"))
		return
	}

	// Create yad list dialog
	cmd := exec.Command("yad", 
		"--list",
		"--title=Ver Propuestas",
		"--text=Selecciona una propuesta para descargar y ver:",
		"--column=ID",
		"--column=Título",
		"--column=Subtítulo", 
		"--column=Creada",
		"--width=800",
		"--height=400",
		"--search-column=2", // Search in title column
		"--print-column=1",   // Return only ID
		"--single-click",
		"--separator=|")

	// Add proposal items as arguments
	for _, prop := range proposals {
		// Add each field as separate argument
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
	// Remove separator if present (yad sometimes includes it)
	selectedID = strings.TrimSuffix(selectedID, "|")
	
	if selectedID == "" {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("No se seleccionó ninguna propuesta"))
		return
	}

	// Find selected proposal
	var selectedProposal *Proposal
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

	// Download all available files for viewing
	fmt.Printf("%s\n", inputs.InfoStyle.Render("Descargando archivos de la propuesta: "+selectedProposal.Title))
	
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

	// Download MD file (always try since MD is always generated)
	mdPath := filepath.Join(downloadDir, selectedProposal.ID+".md")
	if err := downloadProposalFile(baseURL, selectedProposal.ID, "md", mdPath); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al descargar MD: "+err.Error()))
	} else {
		fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ MD descargado"))
	}

	// Download HTML file (only if HTML URL exists)
	if selectedProposal.HTMLURL != "" {
		htmlPath := filepath.Join(downloadDir, selectedProposal.ID+".html")
		if err := downloadProposalFile(baseURL, selectedProposal.ID, "html", htmlPath); err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al descargar HTML: "+err.Error()))
		} else {
			fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ HTML descargado"))
		}
	}

	// Download PDF file (only if PDF URL exists)
	if selectedProposal.PDFURL != "" {
		pdfPath := filepath.Join(downloadDir, selectedProposal.ID+".pdf")
		if err := downloadProposalFile(baseURL, selectedProposal.ID, "pdf", pdfPath); err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error al descargar PDF: "+err.Error()))
		} else {
			fmt.Printf("%s\n", inputs.SuccessStyle.Render("✓ PDF descargado"))
		}
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("Archivos descargados en: " + downloadDir))

	// Show success dialog with options
	cmd = exec.Command("yad",
		"--question",
		"--title=Archivos Descargados",
		"--text=Archivos descargados exitosamente.\n¿Qué deseas hacer?",
		"--button=Abrir Carpeta:1",
		"--button=Abrir MD:2",
		"--button=Abrir HTML:3",
		"--button=Abrir PDF:4",
		"--button=Cerrar:0")

	cmd.Run()
	exitCode := cmd.ProcessState.ExitCode()

	switch exitCode {
	case 1: // Open folder
		openDirectory(downloadDir)
	case 2: // Open MD
		openFile(mdPath)
	case 3: // Open HTML
		if selectedProposal.HTMLURL != "" {
			openFile(filepath.Join(downloadDir, selectedProposal.ID+".html"))
		}
	case 4: // Open PDF
		if selectedProposal.PDFURL != "" {
			openFile(filepath.Join(downloadDir, selectedProposal.ID+".pdf"))
		}
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

