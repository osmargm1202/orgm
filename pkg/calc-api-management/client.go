package calcapi

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

// Client represents the API client for calc management
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

// GetBaseURL gets the base URL from API worker
func GetBaseURL() (string, error) {
	// timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	// log.Printf("[DEBUG %s] Obteniendo URL base de calc API management", timestamp)

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
	baseURL, err := cliconfig.GetCachedConfig("api_calc_management", testURL)
	if err == nil && baseURL != "" {
		// log.Printf("[DEBUG %s] URL obtenida desde caché: %s", timestamp, baseURL)
		return strings.TrimSuffix(baseURL, "/"), nil
	}

	// log.Printf("[DEBUG %s] No se pudo obtener URL desde caché: %v, usando fallback", timestamp, err)

	// Fallback to environment variable or default
	baseURL = os.Getenv("CALC_API_MANAGEMENT_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
		// log.Printf("[DEBUG %s] Usando URL por defecto: %s", timestamp, baseURL)
	} else {
		// log.Printf("[DEBUG %s] Usando URL de variable de entorno: %s", timestamp, baseURL)
	}
	return strings.TrimSuffix(baseURL, "/"), nil
}

// Empresa represents an empresa
type Empresa struct {
	ID        int       `json:"id"`
	Nombre    string    `json:"nombre"`
	URLLogo   *string   `json:"url_logo,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UnmarshalJSON custom unmarshaler for Empresa to handle date parsing
func (e *Empresa) UnmarshalJSON(data []byte) error {
	type Alias Empresa
	aux := &struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	dateFormats := []string{
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05.999",
		"2006-01-02T15:04:05.99",
		"2006-01-02T15:04:05.9",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.000000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range dateFormats {
		if t, err := time.Parse(format, aux.CreatedAt); err == nil {
			e.CreatedAt = t
			break
		}
	}

	for _, format := range dateFormats {
		if t, err := time.Parse(format, aux.UpdatedAt); err == nil {
			e.UpdatedAt = t
			break
		}
	}

	return nil
}

// CreateEmpresaRequest represents the request to create an empresa
type CreateEmpresaRequest struct {
	Nombre  string  `json:"nombre"`
	URLLogo *string `json:"url_logo,omitempty"`
}

// Ingeniero represents an ingeniero
type Ingeniero struct {
	ID        int       `json:"id"`
	Nombre    string    `json:"nombre"`
	Profesion string    `json:"profesion"`
	CODIA     string    `json:"codia"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UnmarshalJSON custom unmarshaler for Ingeniero
func (i *Ingeniero) UnmarshalJSON(data []byte) error {
	type Alias Ingeniero
	aux := &struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		*Alias
	}{
		Alias: (*Alias)(i),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	dateFormats := []string{
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05.999",
		"2006-01-02T15:04:05.99",
		"2006-01-02T15:04:05.9",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.000000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range dateFormats {
		if t, err := time.Parse(format, aux.CreatedAt); err == nil {
			i.CreatedAt = t
			break
		}
	}

	for _, format := range dateFormats {
		if t, err := time.Parse(format, aux.UpdatedAt); err == nil {
			i.UpdatedAt = t
			break
		}
	}

	return nil
}

// CreateIngenieroRequest represents the request to create an ingeniero
type CreateIngenieroRequest struct {
	Nombre    string `json:"nombre"`
	Profesion string `json:"profesion"`
	CODIA     string `json:"codia"`
}

// Proyecto represents a proyecto
type Proyecto struct {
	ID        int       `json:"id"`
	Nombre    string    `json:"nombre"`
	Cliente   string    `json:"cliente"`
	Ubicacion string    `json:"ubicacion"`
	URLLogo   *string   `json:"url_logo,omitempty"`
	EmpresaID int       `json:"empresa_id"`
	Empresa   *Empresa  `json:"empresa,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UnmarshalJSON custom unmarshaler for Proyecto
func (p *Proyecto) UnmarshalJSON(data []byte) error {
	type Alias Proyecto
	aux := &struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	dateFormats := []string{
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05.999",
		"2006-01-02T15:04:05.99",
		"2006-01-02T15:04:05.9",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.000000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range dateFormats {
		if t, err := time.Parse(format, aux.CreatedAt); err == nil {
			p.CreatedAt = t
			break
		}
	}

	for _, format := range dateFormats {
		if t, err := time.Parse(format, aux.UpdatedAt); err == nil {
			p.UpdatedAt = t
			break
		}
	}

	return nil
}

// CreateProyectoRequest represents the request to create a proyecto
type CreateProyectoRequest struct {
	Nombre    string  `json:"nombre"`
	Cliente   string  `json:"cliente"`
	Ubicacion string  `json:"ubicacion"`
	URLLogo   *string `json:"url_logo,omitempty"`
	EmpresaID int     `json:"empresa_id"`
}

// Calculo represents a calculo
type Calculo struct {
	ID              int         `json:"id"`
	ProyectoID      int         `json:"proyecto_id"`
	TipoCalculoID   int         `json:"tipo_calculo_id"`
	IngenieroID     int         `json:"ingeniero_id"`
	HTML            bool        `json:"html"`
	BN              bool        `json:"bn"`
	Eficiencia      bool        `json:"eficiencia"`
	FP              bool        `json:"fp"`
	AlimentadoresSize string    `json:"alimentadores_size"`
	CuadrosSize     string      `json:"cuadros_size"`
	ModulosTRFSize  string      `json:"modulos_trf_size"`
	Color           string      `json:"color"`
	Proyecto        *Proyecto   `json:"proyecto,omitempty"`
	Ingeniero       *Ingeniero  `json:"ingeniero,omitempty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// UnmarshalJSON custom unmarshaler for Calculo
func (c *Calculo) UnmarshalJSON(data []byte) error {
	type Alias Calculo
	aux := &struct {
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	dateFormats := []string{
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05.999",
		"2006-01-02T15:04:05.99",
		"2006-01-02T15:04:05.9",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.000000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range dateFormats {
		if t, err := time.Parse(format, aux.CreatedAt); err == nil {
			c.CreatedAt = t
			break
		}
	}

	for _, format := range dateFormats {
		if t, err := time.Parse(format, aux.UpdatedAt); err == nil {
			c.UpdatedAt = t
			break
		}
	}

	return nil
}

// CreateCalculoRequest represents the request to create a calculo
type CreateCalculoRequest struct {
	TipoCalculoID   int    `json:"tipo_calculo_id"`
	IngenieroID     int    `json:"ingeniero_id"`
	HTML            bool   `json:"html"`
	BN              bool   `json:"bn"`
	Eficiencia      bool   `json:"eficiencia"`
	FP              bool   `json:"fp"`
	AlimentadoresSize string `json:"alimentadores_size"`
	CuadrosSize     string `json:"cuadros_size"`
	ModulosTRFSize  string `json:"modulos_trf_size"`
	Color           string `json:"color"`
}

// UpdateCalculoRequest represents the request to update a calculo
type UpdateCalculoRequest struct {
	HTML            *bool   `json:"html,omitempty"`
	BN              *bool   `json:"bn,omitempty"`
	Eficiencia      *bool   `json:"eficiencia,omitempty"`
	FP              *bool   `json:"fp,omitempty"`
	AlimentadoresSize *string `json:"alimentadores_size,omitempty"`
	CuadrosSize     *string `json:"cuadros_size,omitempty"`
	ModulosTRFSize  *string `json:"modulos_trf_size,omitempty"`
	Color           *string `json:"color,omitempty"`
	TipoCalculoID   *int    `json:"tipo_calculo_id,omitempty"`
	IngenieroID     *int    `json:"ingeniero_id,omitempty"`
}

// SearchProyectosResponse represents the search response
type SearchProyectosResponse struct {
	Proyectos []Proyecto `json:"proyectos"`
	Total     int        `json:"total"`
}

// SearchEmpresasResponse represents the search empresas response
type SearchEmpresasResponse struct {
	Empresas []Empresa `json:"empresas"`
	Total    int       `json:"total"`
}

// SearchIngenierosResponse represents the search ingenieros response
type SearchIngenierosResponse struct {
	Ingenieros []Ingeniero `json:"ingenieros"`
	Total      int         `json:"total"`
}

// SearchCalculosResponse represents the search calculos response
type SearchCalculosResponse struct {
	Calculos []Calculo `json:"calculos"`
	Total    int       `json:"total"`
}

// ExportCalculoResponse represents the export response
type ExportCalculoResponse struct {
	Calculo     Calculo     `json:"calculo"`
	Proyecto    Proyecto    `json:"proyecto"`
	Empresa     Empresa     `json:"empresa"`
	Ingeniero   Ingeniero   `json:"ingeniero"`
}

// Empresas methods

// GetEmpresas returns all empresas
func (c *Client) GetEmpresas() ([]Empresa, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/v1/empresas", nil)
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

	var empresas []Empresa
	if err := json.NewDecoder(resp.Body).Decode(&empresas); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return empresas, nil
}

// GetEmpresa returns an empresa by ID
func (c *Client) GetEmpresa(id int) (*Empresa, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/empresas/%d", c.BaseURL, id), nil)
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

	var empresa Empresa
	if err := json.NewDecoder(resp.Body).Decode(&empresa); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &empresa, nil
}

// CreateEmpresa creates a new empresa
func (c *Client) CreateEmpresa(req CreateEmpresaRequest) (*Empresa, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/api/v1/empresas", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	c.AuthFunc(httpReq)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error de conexión a la API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error del servidor (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var empresa Empresa
	if err := json.NewDecoder(resp.Body).Decode(&empresa); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &empresa, nil
}

// UpdateEmpresa updates an empresa
func (c *Client) UpdateEmpresa(id int, req CreateEmpresaRequest) (*Empresa, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	httpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/empresas/%d", c.BaseURL, id), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	c.AuthFunc(httpReq)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error de conexión a la API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error del servidor (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var empresa Empresa
	if err := json.NewDecoder(resp.Body).Decode(&empresa); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &empresa, nil
}

// DeleteEmpresa deletes an empresa
func (c *Client) DeleteEmpresa(id int) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/empresas/%d", c.BaseURL, id), nil)
	if err != nil {
		return fmt.Errorf("error creando solicitud: %v", err)
	}
	c.AuthFunc(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error de conexión a la API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error del servidor (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// Ingenieros methods

// GetIngenieros returns all ingenieros
func (c *Client) GetIngenieros() ([]Ingeniero, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/v1/ingenieros", nil)
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

	var ingenieros []Ingeniero
	if err := json.NewDecoder(resp.Body).Decode(&ingenieros); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return ingenieros, nil
}

// GetIngeniero returns an ingeniero by ID
func (c *Client) GetIngeniero(id int) (*Ingeniero, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/ingenieros/%d", c.BaseURL, id), nil)
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

	var ingeniero Ingeniero
	if err := json.NewDecoder(resp.Body).Decode(&ingeniero); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &ingeniero, nil
}

// CreateIngeniero creates a new ingeniero
func (c *Client) CreateIngeniero(req CreateIngenieroRequest) (*Ingeniero, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/api/v1/ingenieros", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	c.AuthFunc(httpReq)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error de conexión a la API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error del servidor (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var ingeniero Ingeniero
	if err := json.NewDecoder(resp.Body).Decode(&ingeniero); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &ingeniero, nil
}

// UpdateIngeniero updates an ingeniero
func (c *Client) UpdateIngeniero(id int, req CreateIngenieroRequest) (*Ingeniero, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	httpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/ingenieros/%d", c.BaseURL, id), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	c.AuthFunc(httpReq)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error de conexión a la API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error del servidor (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var ingeniero Ingeniero
	if err := json.NewDecoder(resp.Body).Decode(&ingeniero); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &ingeniero, nil
}

// DeleteIngeniero deletes an ingeniero
func (c *Client) DeleteIngeniero(id int) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/ingenieros/%d", c.BaseURL, id), nil)
	if err != nil {
		return fmt.Errorf("error creando solicitud: %v", err)
	}
	c.AuthFunc(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error de conexión a la API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error del servidor (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// Proyectos methods

// GetProyectos returns all proyectos
func (c *Client) GetProyectos() ([]Proyecto, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/v1/proyectos", nil)
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

	var proyectos []Proyecto
	if err := json.NewDecoder(resp.Body).Decode(&proyectos); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return proyectos, nil
}

// GetProyecto returns a proyecto by ID
func (c *Client) GetProyecto(id int) (*Proyecto, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/proyectos/%d", c.BaseURL, id), nil)
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

	var proyecto Proyecto
	if err := json.NewDecoder(resp.Body).Decode(&proyecto); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &proyecto, nil
}

// CreateProyecto creates a new proyecto
func (c *Client) CreateProyecto(req CreateProyectoRequest) (*Proyecto, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/api/v1/proyectos", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	c.AuthFunc(httpReq)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error de conexión a la API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error del servidor (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var proyecto Proyecto
	if err := json.NewDecoder(resp.Body).Decode(&proyecto); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &proyecto, nil
}

// UpdateProyecto updates a proyecto
func (c *Client) UpdateProyecto(id int, req CreateProyectoRequest) (*Proyecto, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	httpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/proyectos/%d", c.BaseURL, id), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	c.AuthFunc(httpReq)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error de conexión a la API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error del servidor (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var proyecto Proyecto
	if err := json.NewDecoder(resp.Body).Decode(&proyecto); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &proyecto, nil
}

// DeleteProyecto deletes a proyecto
func (c *Client) DeleteProyecto(id int) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/proyectos/%d", c.BaseURL, id), nil)
	if err != nil {
		return fmt.Errorf("error creando solicitud: %v", err)
	}
	c.AuthFunc(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error de conexión a la API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error del servidor (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// SearchProyectos searches proyectos by query
func (c *Client) SearchProyectos(query string, limit int) (*SearchProyectosResponse, error) {
	url := fmt.Sprintf("%s/api/v1/proyectos/search", c.BaseURL)
	params := []string{}
	if query != "" {
		params = append(params, fmt.Sprintf("q=%s", query))
	}
	if limit > 0 {
		params = append(params, fmt.Sprintf("limit=%d", limit))
	}
	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	req, err := http.NewRequest("GET", url, nil)
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

	var searchResp SearchProyectosResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &searchResp, nil
}

// SearchEmpresas searches empresas by query
func (c *Client) SearchEmpresas(query string, limit int) (*SearchEmpresasResponse, error) {
	url := fmt.Sprintf("%s/api/v1/empresas/search", c.BaseURL)
	params := []string{}
	if query != "" {
		params = append(params, fmt.Sprintf("q=%s", query))
	}
	if limit > 0 {
		params = append(params, fmt.Sprintf("limit=%d", limit))
	}
	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	req, err := http.NewRequest("GET", url, nil)
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

	var searchResp SearchEmpresasResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &searchResp, nil
}

// SearchIngenieros searches ingenieros by query
func (c *Client) SearchIngenieros(query string, limit int) (*SearchIngenierosResponse, error) {
	url := fmt.Sprintf("%s/api/v1/ingenieros/search", c.BaseURL)
	params := []string{}
	if query != "" {
		params = append(params, fmt.Sprintf("q=%s", query))
	}
	if limit > 0 {
		params = append(params, fmt.Sprintf("limit=%d", limit))
	}
	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	req, err := http.NewRequest("GET", url, nil)
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

	var searchResp SearchIngenierosResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &searchResp, nil
}

// SearchCalculos searches calculos by query and filters
func (c *Client) SearchCalculos(query string, proyectoID, ingenieroID, tipoCalculoID *int, limit int) (*SearchCalculosResponse, error) {
	url := fmt.Sprintf("%s/api/v1/calculos/search", c.BaseURL)
	params := []string{}
	if query != "" {
		params = append(params, fmt.Sprintf("q=%s", query))
	}
	if proyectoID != nil {
		params = append(params, fmt.Sprintf("proyecto_id=%d", *proyectoID))
	}
	if ingenieroID != nil {
		params = append(params, fmt.Sprintf("ingeniero_id=%d", *ingenieroID))
	}
	if tipoCalculoID != nil {
		params = append(params, fmt.Sprintf("tipo_calculo_id=%d", *tipoCalculoID))
	}
	if limit > 0 {
		params = append(params, fmt.Sprintf("limit=%d", limit))
	}
	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	req, err := http.NewRequest("GET", url, nil)
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

	var searchResp SearchCalculosResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &searchResp, nil
}

// Calculos methods

// GetCalculos returns all calculos for a proyecto
func (c *Client) GetCalculos(proyectoID int) ([]Calculo, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/proyectos/%d/calculos", c.BaseURL, proyectoID), nil)
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

	var calculos []Calculo
	if err := json.NewDecoder(resp.Body).Decode(&calculos); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return calculos, nil
}

// GetCalculo returns a calculo by ID
func (c *Client) GetCalculo(id int) (*Calculo, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/calculos/%d", c.BaseURL, id), nil)
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

	var calculo Calculo
	if err := json.NewDecoder(resp.Body).Decode(&calculo); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &calculo, nil
}

// CreateCalculo creates a new calculo
func (c *Client) CreateCalculo(proyectoID int, req CreateCalculoRequest) (*Calculo, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/proyectos/%d/calculos", c.BaseURL, proyectoID), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	c.AuthFunc(httpReq)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error de conexión a la API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error del servidor (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var calculo Calculo
	if err := json.NewDecoder(resp.Body).Decode(&calculo); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &calculo, nil
}

// UpdateCalculo updates a calculo
func (c *Client) UpdateCalculo(id int, req UpdateCalculoRequest) (*Calculo, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	httpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/calculos/%d", c.BaseURL, id), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	c.AuthFunc(httpReq)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error de conexión a la API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error del servidor (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var calculo Calculo
	if err := json.NewDecoder(resp.Body).Decode(&calculo); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &calculo, nil
}

// DeleteCalculo deletes a calculo
func (c *Client) DeleteCalculo(id int) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/calculos/%d", c.BaseURL, id), nil)
	if err != nil {
		return fmt.Errorf("error creando solicitud: %v", err)
	}
	c.AuthFunc(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error de conexión a la API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error del servidor (HTTP %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// ExportCalculo exports a calculo with all related data
func (c *Client) ExportCalculo(id int) (*ExportCalculoResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/calculos/%d/export", c.BaseURL, id), nil)
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

	var exportResp ExportCalculoResponse
	if err := json.NewDecoder(resp.Body).Decode(&exportResp); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &exportResp, nil
}

