package propapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/osmargm1202/orgm/pkg/cliconfig"
)

// Config represents the configuration structure
type Config struct {
	URL struct {
		PropuestasAPI string `toml:"propuestas_api"`
	} `toml:"url"`
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

// Client represents the API client for proposals
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	AuthFunc   func(*http.Request) // Inject auth from cmd.EnsureGCloudIDToken
}

// NewClient creates a new API client
func NewClient(baseURL string, authFunc func(*http.Request)) *Client {
	return &Client{
		BaseURL:    strings.TrimSuffix(baseURL, "/"),
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		AuthFunc:   authFunc,
	}
}

// GetProposals returns all proposals
func (c *Client) GetProposals() ([]Proposal, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/proposals", nil)
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud: %v", err)
	}
	c.AuthFunc(req)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error de conexión a la API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error del servidor (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var proposals []Proposal
	if err := json.NewDecoder(resp.Body).Decode(&proposals); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return proposals, nil
}

// CreateProposal creates a new proposal
func (c *Client) CreateProposal(request TextGenerationRequest) (*TextGenerationResponse, error) {
	// Set default model if not provided
	if request.Model == "" {
		request.Model = "gpt-5-chat-latest"
	}
	
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}
	
	req, err := http.NewRequest("POST", c.BaseURL+"/generate-text", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	c.AuthFunc(req)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}
	
	var response TextGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	
	return &response, nil
}

// DownloadProposalFile downloads a proposal file
func (c *Client) DownloadProposalFile(proposalID, fileType, filepath string) error {
	url := fmt.Sprintf("%s/proposal/%s/%s", c.BaseURL, proposalID, fileType)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	c.AuthFunc(req)
	
	resp, err := c.HTTPClient.Do(req)
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

// GenerateHTML generates HTML for a proposal
func (c *Client) GenerateHTML(proposalID string) (*HTMLGenerationResponse, error) {
	htmlRequest := HTMLGenerationRequest{ProposalID: proposalID}
	jsonData, err := json.Marshal(htmlRequest)
	if err != nil {
		return nil, fmt.Errorf("error al serializar solicitud HTML: %v", err)
	}
	
	req, err := http.NewRequest("POST", c.BaseURL+"/generate-html", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error al crear solicitud HTML: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	c.AuthFunc(req)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al generar HTML: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error del servidor (%d): %s", resp.StatusCode, string(body))
	}
	
	var response HTMLGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta HTML: %v", err)
	}
	
	return &response, nil
}

// GeneratePDF generates PDF for a proposal
func (c *Client) GeneratePDF(proposalID, modo string) (*PDFGenerationResponse, error) {
	if modo == "" {
		modo = "normal"
	}
	
	pdfRequest := PDFGenerationRequest{ProposalID: proposalID, Modo: modo}
	jsonData, err := json.Marshal(pdfRequest)
	if err != nil {
		return nil, fmt.Errorf("error al serializar solicitud PDF: %v", err)
	}
	
	req, err := http.NewRequest("POST", c.BaseURL+"/generate-pdf", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error al crear solicitud PDF: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	c.AuthFunc(req)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al generar PDF: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error del servidor (%d): %s", resp.StatusCode, string(body))
	}
	
	var response PDFGenerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta PDF: %v", err)
	}
	
	return &response, nil
}

// RegenerateProposal regenerates a proposal with new title/subtitle/prompt
func (c *Client) RegenerateProposal(proposalID, title, subtitle, prompt string) error {
	body := map[string]string{
		"title":    title,
		"subtitle": subtitle,
		"prompt":   prompt,
		"model":    "gpt-5-chat-latest",
	}
	
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error al serializar solicitud: %v", err)
	}
	
	url := fmt.Sprintf("%s/proposal/%s/regenerate", c.BaseURL, proposalID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error al crear solicitud: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	c.AuthFunc(req)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error de red: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("fallo al regenerar: %s", string(body))
	}
	
	return nil
}

// UpdateTitleSubtitle updates only title and subtitle of a proposal
func (c *Client) UpdateTitleSubtitle(proposalID, title, subtitle string) error {
	body := map[string]string{
		"title":    title,
		"subtitle": subtitle,
	}
	
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error al serializar solicitud: %v", err)
	}
	
	url := fmt.Sprintf("%s/proposal/%s/title-subtitle", c.BaseURL, proposalID)
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error al crear solicitud: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	c.AuthFunc(req)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error de red: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("fallo al actualizar: %s", string(body))
	}
	
	return nil
}

// ModifyProposal modifies a proposal with new prompt
func (c *Client) ModifyProposal(proposalID, title, subtitle, prompt string) (*ModifyProposalResponse, error) {
	request := TextGenerationRequest{
		Title:    title,
		Subtitle: subtitle,
		Prompt:   prompt,
		Model:    "gpt-5-chat-latest",
	}
	
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error al serializar la solicitud: %v", err)
	}
	
	url := fmt.Sprintf("%s/proposal/%s", c.BaseURL, proposalID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error al crear la solicitud: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	c.AuthFunc(req)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error al enviar la solicitud: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error del servidor (%d): %s", resp.StatusCode, string(body))
	}
	
	var response ModifyProposalResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error al decodificar la respuesta: %v", err)
	}
	
	return &response, nil
}

// DeleteProposal deletes a proposal
func (c *Client) DeleteProposal(proposalID string) error {
	url := fmt.Sprintf("%s/proposal/%s", c.BaseURL, proposalID)
	
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error al crear solicitud: %v", err)
	}
	
	c.AuthFunc(req)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error de red: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error del servidor (%d): %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// GetBaseURL gets the base URL from config, works for both CLI and Wails contexts
func GetBaseURL() (string, error) {
	// timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	// log.Printf("[DEBUG %s] Obteniendo URL base de propuestas API", timestamp)

	// Test function to validate URL (simple HTTP GET request with timeout)
	testURL := func(url string) error {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			return err
		}
		resp.Body.Close()
		// Accept any 2xx or 3xx status code as valid
		if resp.StatusCode >= 400 {
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		return nil
	}

	// Try cached config first
	baseURL, err := cliconfig.GetCachedConfig("propuestas_api", testURL)
	if err == nil && baseURL != "" {
		// log.Printf("[DEBUG %s] URL obtenida desde caché: %s", timestamp, baseURL)
		return strings.TrimSuffix(baseURL, "/"), nil
	}

	// log.Printf("[DEBUG %s] No se pudo obtener URL desde caché: %v, usando fallback", timestamp, err)

	// Fallback to environment variable or default
	baseURL = os.Getenv("PROPUESTAS_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
		// log.Printf("[DEBUG %s] Usando URL por defecto: %s", timestamp, baseURL)
	} else {
		// log.Printf("[DEBUG %s] Usando URL de variable de entorno: %s", timestamp, baseURL)
	}
	return strings.TrimSuffix(baseURL, "/"), nil
}
