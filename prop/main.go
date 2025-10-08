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

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

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

// Helper functions
func getBaseURL() (string, error) {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error obteniendo directorio home: %v", err)
	}
	
	// Try to load config from ~/.config/orgm/config.json
	configPath := filepath.Join(homeDir, ".config", "orgm", "config.json")
	if _, err := os.Stat(configPath); err == nil {
		configData, err := os.ReadFile(configPath)
		if err == nil {
			var config map[string]interface{}
			if err := json.Unmarshal(configData, &config); err == nil {
				if urls, ok := config["url"].(map[string]interface{}); ok {
					if baseURL, ok := urls["propuestas_api"].(string); ok && baseURL != "" {
						return strings.TrimSuffix(baseURL, "/"), nil
					}
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

// getGCloudIDToken gets the Google Cloud ID token using orgm gauth command
func getGCloudIDToken() (string, error) {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error obteniendo directorio home: %v", err)
	}
	
	// Check if token file exists
	tokenPath := filepath.Join(homeDir, ".config", "orgm", ".gauth_token.json")
	
	// Try to read existing token
	if _, err := os.Stat(tokenPath); err == nil {
		tokenData, err := os.ReadFile(tokenPath)
		if err == nil {
			var tokenInfo TokenInfo
			if err := json.Unmarshal(tokenData, &tokenInfo); err == nil {
				// Check if token is still valid (not expired)
				if time.Now().Before(tokenInfo.ExpiresAt) {
					return tokenInfo.Token, nil
				}
			}
		}
	}
	
	// Token doesn't exist or is expired, generate new one
	return generateNewToken()
}

// generateNewToken generates a new token using orgm gauth command
func generateNewToken() (string, error) {
	// Execute orgm gauth command
	cmd := exec.Command("orgm", "gauth")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error ejecutando orgm gauth: %v", err)
	}
	
	// Parse the output to get the token
	var tokenInfo TokenInfo
	if err := json.Unmarshal(output, &tokenInfo); err != nil {
		return "", fmt.Errorf("error parseando respuesta de orgm gauth: %v", err)
	}
	
	// Save token to file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error obteniendo directorio home: %v", err)
	}
	
	configDir := filepath.Join(homeDir, ".config", "orgm")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("error creando directorio de configuración: %v", err)
	}
	
	tokenPath := filepath.Join(configDir, ".gauth_token.json")
	tokenData, err := json.Marshal(tokenInfo)
	if err != nil {
		return "", fmt.Errorf("error serializando token: %v", err)
	}
	
	if err := os.WriteFile(tokenPath, tokenData, 0600); err != nil {
		return "", fmt.Errorf("error guardando token: %v", err)
	}
	
	return tokenInfo.Token, nil
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
