package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/osmargm1202/orgm/pkg/propapi"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// Helper function to execute CLI commands
func executeCLICommand(command string, args []string) (string, error) {
	// Find the orgm binary - first try PATH, then look for it in common locations
	orgmPath := "orgm"
	
	// Try to find orgm in PATH
	path, err := exec.LookPath(orgmPath)
	if err == nil {
		orgmPath = path
	} else {
		// Try common locations
		homeDir, _ := os.UserHomeDir()
		candidatePaths := []string{
			filepath.Join(homeDir, ".local", "bin", "orgm"),
			filepath.Join(homeDir, "bin", "orgm"),
			"/usr/local/bin/orgm",
			"/usr/bin/orgm",
		}
		for _, candidate := range candidatePaths {
			if _, err := os.Stat(candidate); err == nil {
				orgmPath = candidate
				break
			}
		}
	}
	
	cmd := exec.Command(orgmPath, append([]string{command}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Include the output in the error message for debugging
		return "", fmt.Errorf("error executing orgm %s: %v\nOutput: %s", command, err, string(output))
	}
	
	return strings.TrimSpace(string(output)), nil
}

// Get configuration value from CLI
func getConfigFromCLI(key string) (string, error) {
	return executeCLICommand("viper", []string{"get", key})
}

// Get authentication token from CLI
func getTokenFromCLI() (string, error) {
	return executeCLICommand("gauth", []string{"--print-token"})
}

// Config represents the configuration structure
type Config struct {
	URL struct {
		PropuestasAPI string `toml:"propuestas_api"`
	} `toml:"url"`
}

// App struct
type App struct {
	ctx    context.Context
	client *propapi.Client
}

// TokenInfo represents the token information from gauth
type TokenInfo struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// UnmarshalJSON custom unmarshaler for TokenInfo to handle date parsing
func (t *TokenInfo) UnmarshalJSON(data []byte) error {
	type Alias TokenInfo
	aux := &struct {
		ExpiresAt string `json:"expires_at"`
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
		if parsedTime, err := time.Parse(format, aux.ExpiresAt); err == nil {
			t.ExpiresAt = parsedTime
			return nil
		}
	}
	
	// If all formats fail, set to zero time
	t.ExpiresAt = time.Time{}
	return nil
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Get base URL from CLI
	baseURL, err := getConfigFromCLI("propuestas_api")
	if err != nil {
		fmt.Printf("Error getting base URL: %v\n", err)
		baseURL = "http://localhost:8000" // fallback
	}

	// Create auth function that uses CLI to get token
	authFunc := func(req *http.Request) {
		token, err := getTokenFromCLI()
		if err != nil || token == "" {
			fmt.Printf("Error: No se pudo obtener token de autenticación: %v\n", err)
			fmt.Printf("Por favor ejecuta 'orgm gauth' para obtener un token válido\n")
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Create client
	client := propapi.NewClient(baseURL, authFunc)

	return &App{
		client: client,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// Verify authentication on startup
	token, err := getTokenFromCLI()
	if err != nil || token == "" {
		fmt.Printf("⚠️  Advertencia: Problema de autenticación detectado\n")
		fmt.Printf("   Error: %v\n", err)
		fmt.Printf("   Solución: Ejecuta 'orgm gauth' en la terminal para obtener un token válido\n")
		fmt.Printf("   La aplicación puede funcionar con funcionalidad limitada\n")
	}
}

// Proposal represents a proposal from the API
type Proposal struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Subtitle  string    `json:"subtitle"`
	Prompt    string    `json:"prompt"`
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

// CheckAuthStatus checks the authentication status
func (a *App) CheckAuthStatus() map[string]interface{} {
	token, err := getTokenFromCLI()
	if err != nil || token == "" {
		errorMsg := "Unknown error"
		if err != nil {
			errorMsg = err.Error()
		}
		return map[string]interface{}{
			"authenticated": false,
			"error": errorMsg,
			"message": "No se pudo obtener token de autenticación. Ejecuta 'orgm gauth' para obtener un token válido.",
		}
	}
	return map[string]interface{}{
		"authenticated": true,
		"message": "Autenticación exitosa",
	}
}

// GetProposals returns all proposals
func (a *App) GetProposals() []propapi.Proposal {
	proposals, err := a.client.GetProposals()
	if err != nil {
		fmt.Printf("Error obteniendo propuestas: %v\n", err)
		return []propapi.Proposal{}
	}
	return proposals
}

// CreateProposal creates a new proposal
func (a *App) CreateProposal(request propapi.TextGenerationRequest) *propapi.TextGenerationResponse {
	response, err := a.client.CreateProposal(request)
	if err != nil {
		fmt.Printf("Error creando propuesta: %v\n", err)
		return nil
	}
	return response
}

// DownloadProposal downloads a proposal file
func (a *App) DownloadProposal(proposalID, format string) map[string]interface{} {
	// Create downloads directory
	homeDir, _ := os.UserHomeDir()
	downloadsDir := filepath.Join(homeDir, "Downloads")
	os.MkdirAll(downloadsDir, 0755)
	
	// Create file path
	filename := fmt.Sprintf("%s.%s", proposalID, format)
	filepath := filepath.Join(downloadsDir, filename)
	
	err := a.client.DownloadProposalFile(proposalID, format, filepath)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	return map[string]interface{}{"success": true, "filepath": filepath}
}

// GenerateHTML generates HTML for a proposal
func (a *App) GenerateHTML(proposalID string) map[string]interface{} {
	response, err := a.client.GenerateHTML(proposalID)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	return map[string]interface{}{"success": true, "html_url": response.HTMLURL}
}

// GeneratePDF generates PDF for a proposal
func (a *App) GeneratePDF(proposalID, modo string) map[string]interface{} {
	if modo == "" {
		modo = "normal"
	}
	
	response, err := a.client.GeneratePDF(proposalID, modo)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	return map[string]interface{}{"success": true, "pdf_url": response.PDFURL}
}

// RegenerateProposal regenerates a proposal with new title/subtitle/prompt
func (a *App) RegenerateProposal(proposalID, title, subtitle, prompt string) map[string]interface{} {
	err := a.client.RegenerateProposal(proposalID, title, subtitle, prompt)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	return map[string]interface{}{"success": true}
}

// UpdateTitleSubtitle updates only title and subtitle of a proposal
func (a *App) UpdateTitleSubtitle(proposalID, title, subtitle string) map[string]interface{} {
	err := a.client.UpdateTitleSubtitle(proposalID, title, subtitle)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	return map[string]interface{}{"success": true}
}

// ModifyProposal modifies a proposal with new prompt
func (a *App) ModifyProposal(proposalID, title, subtitle, prompt string) map[string]interface{} {
	response, err := a.client.ModifyProposal(proposalID, title, subtitle, prompt)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	return map[string]interface{}{"success": true, "id": response.ID}
}

// DeleteProposal deletes a proposal
func (a *App) DeleteProposal(proposalID string) map[string]interface{} {
	err := a.client.DeleteProposal(proposalID)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	return map[string]interface{}{"success": true}
}

// OpenDirectory opens the downloads directory
func (a *App) OpenDirectory() map[string]interface{} {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	downloadsDir := filepath.Join(homeDir, "Downloads")
	
	// Try to open directory
	var cmd *exec.Cmd
	switch {
	case isCommandAvailable("xdg-open"):
		cmd = exec.Command("xdg-open", downloadsDir)
	case isCommandAvailable("open"):
		cmd = exec.Command("open", downloadsDir)
	case isCommandAvailable("explorer"):
		cmd = exec.Command("explorer", downloadsDir)
	default:
		return map[string]interface{}{"success": false, "error": "No se pudo abrir el directorio automáticamente"}
	}
	
	if err := cmd.Start(); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	return map[string]interface{}{"success": true}
}

// OpenFile opens a file with the default application
func (a *App) OpenFile(filepath string) map[string]interface{} {
	var cmd *exec.Cmd
	switch {
	case isCommandAvailable("xdg-open"):
		cmd = exec.Command("xdg-open", filepath)
	case isCommandAvailable("open"):
		cmd = exec.Command("open", filepath)
	case isCommandAvailable("explorer"):
		cmd = exec.Command("explorer", filepath)
	default:
		return map[string]interface{}{"success": false, "error": "No se pudo abrir el archivo automáticamente"}
	}
	
	if err := cmd.Start(); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	return map[string]interface{}{"success": true}
}

// Helper functions
// isCommandAvailable checks if a command is available in the system PATH
func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// main is the entry point for the Wails application
func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Gestor de Propuestas",
		Width:  1400,
		Height: 900,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
