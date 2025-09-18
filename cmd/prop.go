package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
  gen     Crear nueva propuesta y descargar archivos
  html    Generar HTML para propuesta existente y descargar
  pdf     Generar PDF para propuesta existente y descargar
  get     Descargar archivos de propuesta existente
  mod     Modificar propuesta existente
  find    Buscar propuestas`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
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

// modCmd represents the mod command
var modCmd = &cobra.Command{
	Use:   "mod <id>",
	Short: "Modificar propuesta existente",
	Long:  `Modifica una propuesta existente`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		baseURL, err := getBaseURL()
		if err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error: "+err.Error()))
			return
		}
		modifyProposalByID(baseURL, args[0])
	},
}

// findCmd represents the find command
var findCmd = &cobra.Command{
	Use:   "find <query>",
	Short: "Buscar propuestas",
	Long:  `Busca propuestas por título o subtítulo`,
	Run: func(cmd *cobra.Command, args []string) {
		baseURL, err := getBaseURL()
		if err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error: "+err.Error()))
			return
		}
		query := ""
		if len(args) > 0 {
			query = args[0]
		}
		searchProposals(baseURL, query)
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
	PropCmd.AddCommand(genCmd)
	PropCmd.AddCommand(htmlCmd)
	PropCmd.AddCommand(pdfCmd)
	PropCmd.AddCommand(getCmd)
	PropCmd.AddCommand(modCmd)
	PropCmd.AddCommand(findCmd)
}

func getBaseURL() (string, error) {
	baseURL := viper.GetString("propuestas_api.url")
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


func searchProposals(baseURL, query string) {
	fmt.Printf("%s\n", inputs.TitleStyle.Render("Buscar Propuestas"))
	fmt.Println()

	if query == "" {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("Buscando todas las propuestas..."))
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

		if len(proposals) == 0 {
			fmt.Printf("%s\n", inputs.InfoStyle.Render("No hay propuestas disponibles"))
			return
		}

		fmt.Printf("%s\n", inputs.SuccessStyle.Render(fmt.Sprintf("Se encontraron %d propuestas:", len(proposals))))
		fmt.Println()

		for i, prop := range proposals {
			fmt.Printf("%d. %s\n", i+1, prop.Title)
			fmt.Printf("   ID: %s\n", prop.ID)
			fmt.Printf("   Subtítulo: %s\n", prop.Subtitle)
			fmt.Printf("   Creada: %s\n", prop.CreatedAt.Format("2006-01-02 15:04:05"))
			if prop.HTMLURL != "" {
				fmt.Printf("   HTML: %s (%d bytes)\n", prop.HTMLURL, prop.SizeHTML)
			}
			if prop.PDFURL != "" {
				fmt.Printf("   PDF: %s (%d bytes)\n", prop.PDFURL, prop.SizePDF)
			}
			fmt.Println()
		}
	} else {
		fmt.Printf("%s\n", inputs.InfoStyle.Render("Buscando propuestas con: " + query))

		// URL encode the query parameter
		encodedQuery := url.QueryEscape(query)
		resp, err := http.Get(baseURL + "/proposals/search?q=" + encodedQuery)
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

		if len(proposals) == 0 {
			fmt.Printf("%s\n", inputs.InfoStyle.Render("No se encontraron propuestas con el término: " + query))
			return
		}

		fmt.Printf("%s\n", inputs.SuccessStyle.Render(fmt.Sprintf("Se encontraron %d propuestas:", len(proposals))))
		fmt.Println()

		for i, prop := range proposals {
			fmt.Printf("%d. %s\n", i+1, prop.Title)
			fmt.Printf("   ID: %s\n", prop.ID)
			fmt.Printf("   Subtítulo: %s\n", prop.Subtitle)
			fmt.Printf("   Creada: %s\n", prop.CreatedAt.Format("2006-01-02 15:04:05"))
			if prop.HTMLURL != "" {
				fmt.Printf("   HTML: %s (%d bytes)\n", prop.HTMLURL, prop.SizeHTML)
			}
			if prop.PDFURL != "" {
				fmt.Printf("   PDF: %s (%d bytes)\n", prop.PDFURL, prop.SizePDF)
			}
			fmt.Println()
		}
	}
}

