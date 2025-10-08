package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
)

//go:embed all:frontend/dist
var assets embed.FS

// Config represents the configuration structure
type Config struct {
	URL struct {
		PropuestasAPI string `toml:"propuestas_api"`
	} `toml:"url"`
}

// App struct
type App struct {
	ctx context.Context
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
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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

// GetProposals returns all proposals
func (a *App) GetProposals() []Proposal {
	baseURL, err := getBaseURL()
	if err != nil {
		fmt.Printf("Error obteniendo URL base: %v\n", err)
		return []Proposal{}
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", baseURL+"/proposals", nil)
	if err != nil {
		fmt.Printf("Error creando request: %v\n", err)
		return []Proposal{}
	}
	
	attachAuth(req)
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error haciendo request: %v\n", err)
		return []Proposal{}
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error en respuesta: %d - %s\n", resp.StatusCode, string(body))
		return []Proposal{}
	}
	
	var proposals []Proposal
	if err := json.NewDecoder(resp.Body).Decode(&proposals); err != nil {
		fmt.Printf("Error decodificando respuesta: %v\n", err)
		return []Proposal{}
	}
	
	return proposals
}

// CreateProposal creates a new proposal
func (a *App) CreateProposal(request TextGenerationRequest) *TextGenerationResponse {
	baseURL, err := getBaseURL()
	if err != nil {
		fmt.Printf("Error obteniendo URL base: %v\n", err)
		return nil
	}
	
	// Set default model if not provided
	if request.Model == "" {
		request.Model = "gpt-5-chat-latest"
	}
	
	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("Error serializando request: %v\n", err)
		return nil
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", baseURL+"/generate-text", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creando request: %v\n", err)
		return nil
	}
	
	req.Header.Set("Content-Type", "application/json")
	attachAuth(req)
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error haciendo request: %v\n", err)
		return nil
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error en respuesta: %d - %s\n", resp.StatusCode, string(body))
		return nil
	}
	
	var response TextGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("Error decodificando respuesta: %v\n", err)
		return nil
	}
	
	return &response
}

// DownloadProposal downloads a proposal file
func (a *App) DownloadProposal(proposalID, format string) map[string]interface{} {
	baseURL, err := getBaseURL()
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	url := fmt.Sprintf("%s/proposal/%s/%s", baseURL, proposalID, format)
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	attachAuth(req)
	
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return map[string]interface{}{"success": false, "error": fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))}
	}
	
	// Create downloads directory
	homeDir, _ := os.UserHomeDir()
	downloadsDir := filepath.Join(homeDir, "Downloads")
	os.MkdirAll(downloadsDir, 0755)
	
	// Create file
	filename := fmt.Sprintf("%s.%s", proposalID, format)
	filepath := filepath.Join(downloadsDir, filename)
	
	file, err := os.Create(filepath)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	defer file.Close()
	
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	return map[string]interface{}{"success": true, "filepath": filepath}
}

// GenerateHTML generates HTML for a proposal
func (a *App) GenerateHTML(proposalID string) map[string]interface{} {
	baseURL, err := getBaseURL()
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	htmlRequest := HTMLGenerationRequest{ProposalID: proposalID}
	jsonData, err := json.Marshal(htmlRequest)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", baseURL+"/generate-html", bytes.NewBuffer(jsonData))
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	req.Header.Set("Content-Type", "application/json")
	attachAuth(req)
	
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return map[string]interface{}{"success": false, "error": fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))}
	}
	
	var response HTMLGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	return map[string]interface{}{"success": true, "html_url": response.HTMLURL}
}

// GeneratePDF generates PDF for a proposal
func (a *App) GeneratePDF(proposalID, modo string) map[string]interface{} {
	baseURL, err := getBaseURL()
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	if modo == "" {
		modo = "normal"
	}
	
	pdfRequest := PDFGenerationRequest{ProposalID: proposalID, Modo: modo}
	jsonData, err := json.Marshal(pdfRequest)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", baseURL+"/generate-pdf", bytes.NewBuffer(jsonData))
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	req.Header.Set("Content-Type", "application/json")
	attachAuth(req)
	
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return map[string]interface{}{"success": false, "error": fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))}
	}
	
	var response PDFGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	return map[string]interface{}{"success": true, "pdf_url": response.PDFURL}
}

// RegenerateProposal regenerates a proposal with new title/subtitle/prompt
func (a *App) RegenerateProposal(proposalID, title, subtitle, prompt string) map[string]interface{} {
	baseURL, err := getBaseURL()
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	body := map[string]string{
		"title":   title,
		"subtitle": subtitle,
		"prompt":  prompt,
		"model":   "gpt-5-chat-latest",
	}
	
	jsonData, err := json.Marshal(body)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	url := fmt.Sprintf("%s/proposal/%s/regenerate", baseURL, proposalID)
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	req.Header.Set("Content-Type", "application/json")
	attachAuth(req)
	
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return map[string]interface{}{"success": false, "error": fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))}
	}
	
	return map[string]interface{}{"success": true}
}

// UpdateTitleSubtitle updates only title and subtitle of a proposal
func (a *App) UpdateTitleSubtitle(proposalID, title, subtitle string) map[string]interface{} {
	baseURL, err := getBaseURL()
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	body := map[string]string{
		"title":    title,
		"subtitle": subtitle,
	}
	
	jsonData, err := json.Marshal(body)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	url := fmt.Sprintf("%s/proposal/%s/title-subtitle", baseURL, proposalID)
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	req.Header.Set("Content-Type", "application/json")
	attachAuth(req)
	
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return map[string]interface{}{"success": false, "error": fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))}
	}
	
	return map[string]interface{}{"success": true}
}

// ModifyProposal modifies a proposal with new prompt
func (a *App) ModifyProposal(proposalID, title, subtitle, prompt string) map[string]interface{} {
	baseURL, err := getBaseURL()
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	request := TextGenerationRequest{
		Title:    title,
		Subtitle: subtitle,
		Prompt:   prompt,
		Model:    "gpt-5-chat-latest",
	}
	
	jsonData, err := json.Marshal(request)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	url := fmt.Sprintf("%s/proposal/%s", baseURL, proposalID)
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	req.Header.Set("Content-Type", "application/json")
	attachAuth(req)
	
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return map[string]interface{}{"success": false, "error": fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))}
	}
	
	var response ModifyProposalResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	return map[string]interface{}{"success": true, "id": response.ID}
}

// DeleteProposal deletes a proposal
func (a *App) DeleteProposal(proposalID string) map[string]interface{} {
	baseURL, err := getBaseURL()
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	url := fmt.Sprintf("%s/proposal/%s", baseURL, proposalID)
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	attachAuth(req)
	
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return map[string]interface{}{"success": false, "error": fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body))}
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
func getBaseURL() (string, error) {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error obteniendo directorio home: %v", err)
	}
	
	// Try to load config from ~/.config/orgm/config.toml
	configPath := filepath.Join(homeDir, ".config", "orgm", "config.toml")
	if _, err := os.Stat(configPath); err == nil {
		configData, err := os.ReadFile(configPath)
		if err == nil {
			var config Config
			if _, err := toml.Decode(string(configData), &config); err == nil {
				if config.URL.PropuestasAPI != "" {
					return strings.TrimSuffix(config.URL.PropuestasAPI, "/"), nil
				}
			}
		}
	}
	
	// Fallback to environment variable or default
	baseURL := os.Getenv("PROPUESTAS_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	return strings.TrimSuffix(baseURL, "/"), nil
}

// getGCloudIDToken obtains an ID token for Cloud Run using the same logic as gcr.go
func getGCloudIDToken() (string, error) {
	// Try disk cache first
	if cachedTok, cachedExp, ok := loadCachedToken(); ok {
		if time.Unix(cachedExp, 0).After(time.Now().Add(2 * time.Minute)) {
			return cachedTok, nil
		}
	}

	// Get audience from config
	baseURL, err := getBaseURL()
	if err != nil {
		return "", fmt.Errorf("url.propuestas_api no está configurado: %v", err)
	}

	// Get config path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error obteniendo directorio home: %v", err)
	}
	configPath := filepath.Join(homeDir, ".config", "orgm")
	credFile := filepath.Join(configPath, "orgmdev_google.json")

	// Validate credentials file exists
	if _, err := os.Stat(credFile); err != nil {
		return "", fmt.Errorf("no se encontró el archivo de credenciales: %s", credFile)
	}

	ctx := context.Background()

	// Prefer idtoken for Cloud Run (OIDC ID token for the service URL)
	ts, err := idtoken.NewTokenSource(ctx, baseURL, option.WithCredentialsFile(credFile))
	if err != nil {
		// Fallback: try default credentials path if provided via env
		if _, derr := google.FindDefaultCredentials(ctx); derr != nil {
			return "", fmt.Errorf("no se pudo crear TokenSource: %v", err)
		}
		ts, err = idtoken.NewTokenSource(ctx, baseURL)
		if err != nil {
			return "", fmt.Errorf("no se pudo crear TokenSource (fallback): %v", err)
		}
	}

	tok, err := ts.Token()
	if err != nil {
		return "", fmt.Errorf("no se pudo obtener token: %v", err)
	}

	// Determine expiry
	var expiryUnix int64
	if !tok.Expiry.IsZero() {
		expiryUnix = tok.Expiry.Unix()
	} else {
		// If expiry is missing, set a short-lived default (55 minutes typical for ID tokens)
		expiryUnix = time.Now().Add(55 * time.Minute).Unix()
	}

	// Persist to disk for reuse across runs
	_ = saveCachedToken(tok.AccessToken, expiryUnix)

	return tok.AccessToken, nil
}

// tokenCachePath returns the path where the token cache is stored
func tokenCachePath() string {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "orgm")
	return filepath.Join(configPath, ".gauth_token.json")
}

func loadCachedToken() (string, int64, bool) {
	path := tokenCachePath()
	b, err := os.ReadFile(path)
	if err != nil {
		return "", 0, false
	}
	var data struct {
		Token       string `json:"token"`
		ExpiryUnix  int64  `json:"expiry_unix"`
	}
	if err := json.Unmarshal(b, &data); err != nil {
		return "", 0, false
	}
	if data.Token == "" || data.ExpiryUnix == 0 {
		return "", 0, false
	}
	return data.Token, data.ExpiryUnix, true
}

func saveCachedToken(token string, expiryUnix int64) error {
	path := tokenCachePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data := struct {
		Token      string `json:"token"`
		ExpiryUnix int64  `json:"expiry_unix"`
	}{Token: token, ExpiryUnix: expiryUnix}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

// attachAuth adds the Authorization header using the cached Google ID token
func attachAuth(req *http.Request) {
	// Get token (will generate new one if needed)
	token, err := getGCloudIDToken()
	if err != nil || token == "" {
		fmt.Printf("Warning: No se pudo obtener token de autenticación: %v\n", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
}

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
