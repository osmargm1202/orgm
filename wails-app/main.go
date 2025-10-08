package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
	
)

// App struct
type App struct {
	ctx context.Context
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
		fmt.Printf("Error en respuesta: %d\n", resp.StatusCode)
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
	
	jsonData, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("Error serializando request: %v\n", err)
		return nil
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", baseURL+"/generate", strings.NewReader(string(jsonData)))
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
		fmt.Printf("Error en respuesta: %d\n", resp.StatusCode)
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
	
	url := fmt.Sprintf("%s/proposals/%s/%s", baseURL, proposalID, format)
	
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
		return map[string]interface{}{"success": false, "error": fmt.Sprintf("HTTP %d", resp.StatusCode)}
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

// RegenerateProposal regenerates a proposal
func (a *App) RegenerateProposal(proposalID string) map[string]interface{} {
	baseURL, err := getBaseURL()
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	url := fmt.Sprintf("%s/proposals/%s/regenerate", baseURL, proposalID)
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", url, nil)
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
		return map[string]interface{}{"success": false, "error": fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}
	
	return map[string]interface{}{"success": true}
}

// UpdateProposal updates a proposal
func (a *App) UpdateProposal(updates map[string]interface{}) map[string]interface{} {
	baseURL, err := getBaseURL()
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	proposalID, ok := updates["id"].(string)
	if !ok {
		return map[string]interface{}{"success": false, "error": "ID de propuesta requerido"}
	}
	
	url := fmt.Sprintf("%s/proposals/%s", baseURL, proposalID)
	
	jsonData, err := json.Marshal(updates)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("PUT", url, strings.NewReader(string(jsonData)))
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
		return map[string]interface{}{"success": false, "error": fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}
	
	return map[string]interface{}{"success": true}
}

// DeleteProposal deletes a proposal
func (a *App) DeleteProposal(proposalID string) map[string]interface{} {
	baseURL, err := getBaseURL()
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	url := fmt.Sprintf("%s/proposals/%s", baseURL, proposalID)
	
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
		return map[string]interface{}{"success": false, "error": fmt.Sprintf("HTTP %d", resp.StatusCode)}
	}
	
	return map[string]interface{}{"success": true}
}

// Helper functions
func getBaseURL() (string, error) {
	// Try to load config from parent directory
	configPath := filepath.Join("..", "links.toml")
	if _, err := os.Stat(configPath); err == nil {
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err == nil {
			baseURL := viper.GetString("url.propuestas_api")
			if baseURL != "" {
				return strings.TrimSuffix(baseURL, "/"), nil
			}
		}
	}
	
	// Fallback to environment variable or default
	baseURL := os.Getenv("PROPUESTAS_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return strings.TrimSuffix(baseURL, "/"), nil
}

func attachAuth(req *http.Request) {
	// Try to get token from environment
	token := os.Getenv("GCLOUD_ID_TOKEN")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}
