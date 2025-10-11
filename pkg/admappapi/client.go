package admappapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// Config represents the configuration structure
type Config struct {
	URL struct {
		AdminAPI string `toml:"admapp_api"`
	} `toml:"url"`
}

// Cliente represents a client from the API
type Cliente struct {
	ID                   int       `json:"id"`
	IDTenant             int       `json:"id_tenant"`
	Nombre               string    `json:"nombre"`
	NombreComercial      string    `json:"nombre_comercial"`
	Numero               string    `json:"numero"`
	Correo               string    `json:"correo"`
	Direccion            string    `json:"direccion"`
	Ciudad               string    `json:"ciudad"`
	Provincia            string    `json:"provincia"`
	Telefono             string    `json:"telefono"`
	Representante        string    `json:"representante"`
	CorreoRepresentante   string    `json:"correo_representante"`
	TipoFactura          string    `json:"tipo_factura"`
	Activo               bool      `json:"activo"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// UnmarshalJSON custom unmarshaler for Cliente to handle date parsing
func (c *Cliente) UnmarshalJSON(data []byte) error {
	type Alias Cliente
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

// CreateClienteRequest represents the request for creating a client
type CreateClienteRequest struct {
	Nombre               string `json:"nombre"`
	NombreComercial      string `json:"nombre_comercial"`
	Numero               string `json:"numero"`
	Correo               string `json:"correo"`
	Direccion            string `json:"direccion"`
	Ciudad               string `json:"ciudad"`
	Provincia            string `json:"provincia"`
	Telefono             string `json:"telefono"`
	Representante        string `json:"representante"`
	CorreoRepresentante   string `json:"correo_representante"`
	TipoFactura          string `json:"tipo_factura"`
}

// UpdateClienteRequest represents the request for updating a client
type UpdateClienteRequest struct {
	Nombre               string `json:"nombre"`
	NombreComercial      string `json:"nombre_comercial"`
	Numero               string `json:"numero"`
	Correo               string `json:"correo"`
	Direccion            string `json:"direccion"`
	Ciudad               string `json:"ciudad"`
	Provincia            string `json:"provincia"`
	Telefono             string `json:"telefono"`
	Representante        string `json:"representante"`
	CorreoRepresentante   string `json:"correo_representante"`
	TipoFactura          string `json:"tipo_factura"`
}

// LogoResponse represents the response from logo operations
type LogoResponse struct {
	Path string `json:"path"`
	URL  string `json:"url"`
}

// Proyecto represents a project
type Proyecto struct {
	ID             int    `json:"id"`
	IDTenant       int    `json:"id_tenant"`
	IDCliente      int    `json:"id_cliente"`
	NombreProyecto string `json:"nombre_proyecto"`
	Ubicacion      string `json:"ubicacion"`
	Descripcion    string `json:"descripcion"`
	Activo         bool   `json:"activo"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// CreateProyectoRequest represents the request for creating a project
type CreateProyectoRequest struct {
	IDCliente      int    `json:"id_cliente"`
	NombreProyecto string `json:"nombre_proyecto"`
	Ubicacion      string `json:"ubicacion"`
	Descripcion    string `json:"descripcion"`
}

// UpdateProyectoRequest represents the request for updating a project
type UpdateProyectoRequest struct {
	NombreProyecto string `json:"nombre_proyecto"`
	Ubicacion      string `json:"ubicacion"`
	Descripcion    string `json:"descripcion"`
}

// Cotizacion represents a basic cotización
type Cotizacion struct {
	ID           int     `json:"id"`
	IDTenant     int     `json:"id_tenant"`
	IDCliente    int     `json:"id_cliente"`
	IDProyecto   int     `json:"id_proyecto"`
	IDServicio   int     `json:"id_servicio"`
	Moneda       string  `json:"moneda"`
	Fecha        string  `json:"fecha"`
	TasaMoneda   float64 `json:"tasa_moneda"`
	TiempoEntrega string `json:"tiempo_entrega"`
	Avance       string  `json:"avance"`
	Validez      int     `json:"validez"`
	Estado       string  `json:"estado"`
	Idioma       string  `json:"idioma"`
	Descripcion  string  `json:"descripcion"`
	Retencion    string  `json:"retencion"`
	Descuentop   float64 `json:"descuentop"`
	Retencionp   float64 `json:"retencionp"`
	Itbisp       float64 `json:"itbisp"`
	Activo       bool    `json:"activo"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// CreateCotizacionRequest represents the request for creating a cotización
type CreateCotizacionRequest struct {
	IDCliente      int     `json:"id_cliente"`
	IDProyecto     int     `json:"id_proyecto"`
	IDServicio     int     `json:"id_servicio"`
	Moneda         string  `json:"moneda"`
	Fecha          string  `json:"fecha"`
	TasaMoneda     float64 `json:"tasa_moneda"`
	TiempoEntrega  string  `json:"tiempo_entrega"`
	Avance         string  `json:"avance"`
	Validez        int     `json:"validez"`
	Estado         string  `json:"estado"`
	Idioma         string  `json:"idioma"`
	Descripcion    string  `json:"descripcion"`
	Retencion      string  `json:"retencion"`
	Descuentop     float64 `json:"descuentop"`
	Retencionp     float64 `json:"retencionp"`
	Itbisp         float64 `json:"itbisp"`
}

// Client represents the API client for admin operations
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

// GetClientes returns all clients
func (c *Client) GetClientes(incluirInactivos bool) ([]Cliente, error) {
	url := c.BaseURL + "/api/clientes"
	if incluirInactivos {
		url += "?incluir_inactivos=true"
	}
	
	fmt.Printf("🌐 Realizando GET a: %s\n", url)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud: %v", err)
	}
	
	c.AuthFunc(req)
	req.Header.Set("X-Tenant-Id", "1")
	
	fmt.Printf("📤 Headers enviados: Authorization=%s, X-Tenant-Id=%s\n", 
		req.Header.Get("Authorization")[:20]+"...", req.Header.Get("X-Tenant-Id"))
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		fmt.Printf("❌ Error de conexión: %v\n", err)
		return nil, fmt.Errorf("error de conexión a la API: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("📥 Respuesta recibida: Status %d\n", resp.StatusCode)

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Error del servidor: %s\n", string(body))
		return nil, fmt.Errorf("error del servidor (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var clientes []Cliente
	if err := json.NewDecoder(resp.Body).Decode(&clientes); err != nil {
		fmt.Printf("❌ Error decodificando JSON: %v\n", err)
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	fmt.Printf("✅ Decodificación exitosa: %d clientes\n", len(clientes))
	return clientes, nil
}

// GetClienteByID returns a specific client by ID
func (c *Client) GetClienteByID(id int) (*Cliente, error) {
	url := fmt.Sprintf("%s/api/clientes/%d", c.BaseURL, id)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud: %v", err)
	}
	
	c.AuthFunc(req)
	req.Header.Set("X-Tenant-Id", "1")
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error de conexión a la API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error del servidor (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var cliente Cliente
	if err := json.NewDecoder(resp.Body).Decode(&cliente); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %v", err)
	}

	return &cliente, nil
}

// CreateCliente creates a new client
func (c *Client) CreateCliente(request CreateClienteRequest) (*Cliente, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}
	
	req, err := http.NewRequest("POST", c.BaseURL+"/api/clientes", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Id", "1")
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
	
	var cliente Cliente
	if err := json.NewDecoder(resp.Body).Decode(&cliente); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	
	return &cliente, nil
}

// UpdateCliente updates an existing client
func (c *Client) UpdateCliente(id int, request UpdateClienteRequest) (*Cliente, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}
	
	url := fmt.Sprintf("%s/api/clientes/%d", c.BaseURL, id)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Id", "1")
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
	
	var cliente Cliente
	if err := json.NewDecoder(resp.Body).Decode(&cliente); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	
	return &cliente, nil
}

// DeleteCliente soft deletes a client
func (c *Client) DeleteCliente(id int) error {
	url := fmt.Sprintf("%s/api/clientes/%d", c.BaseURL, id)
	
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	
	req.Header.Set("X-Tenant-Id", "1")
	c.AuthFunc(req)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error calling API: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// RestoreCliente restores a soft-deleted client
func (c *Client) RestoreCliente(id int) error {
	url := fmt.Sprintf("%s/api/clientes/%d/restore", c.BaseURL, id)
	
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	
	req.Header.Set("X-Tenant-Id", "1")
	c.AuthFunc(req)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error calling API: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// UploadClienteLogo uploads a logo for a client
func (c *Client) UploadClienteLogo(id int, filePath string) (*LogoResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Detect MIME type based on file extension
	filename := filepath.Base(filePath)
	mimeType := mime.TypeByExtension(filepath.Ext(filePath))
	if mimeType == "" {
		mimeType = "application/octet-stream" // fallback
	}

	// Create form file field named "file" with correct Content-Type
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	h.Set("Content-Type", mimeType)
	
	part, err := writer.CreatePart(h)
	if err != nil {
		return nil, fmt.Errorf("error creating form file: %v", err)
	}

	// Copy file content to the form field
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("error copying file content: %v", err)
	}

	// Close the writer to finalize the multipart data
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing multipart writer: %v", err)
	}

	url := fmt.Sprintf("%s/api/clientes/%d/logo", c.BaseURL, id)
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Set the correct Content-Type header with boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.AuthFunc(req)
	req.Header.Set("X-Tenant-Id", "1")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var logoResp LogoResponse
	if err := json.NewDecoder(resp.Body).Decode(&logoResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &logoResp, nil
}

// GetClienteLogoURL gets the logo URL for a client
func (c *Client) GetClienteLogoURL(id int) (*LogoResponse, error) {
	url := fmt.Sprintf("%s/api/clientes/%d/logo", c.BaseURL, id)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	
	c.AuthFunc(req)
	req.Header.Set("X-Tenant-Id", "1")
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Read and log the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	fmt.Printf("🖼️ GetClienteLogoURL response: %s\n", string(body))

	var logoResp LogoResponse
	if err := json.Unmarshal(body, &logoResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	
	fmt.Printf("✅ Logo parsed - Path: %s, URL: %s\n", logoResp.Path, logoResp.URL)

	return &logoResp, nil
}

// GetProyectos returns projects for a specific client
func (c *Client) GetProyectos(idCliente int, incluirInactivos bool) ([]Proyecto, error) {
	var url string
	if idCliente > 0 {
		url = fmt.Sprintf("%s/api/clientes/%d/proyectos", c.BaseURL, idCliente)
	} else {
		url = c.BaseURL + "/api/proyectos"
	}
	
	if incluirInactivos {
		url += "?incluir_inactivos=true"
	}
	
	fmt.Printf("🌐 Realizando GET a: %s\n", url)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud: %v", err)
	}
	
	c.AuthFunc(req)
	req.Header.Set("X-Tenant-Id", "1")
	
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

	fmt.Printf("✅ Decodificación exitosa: %d proyectos\n", len(proyectos))
	return proyectos, nil
}

// GetProyectoByID returns a specific project by ID
func (c *Client) GetProyectoByID(id int) (*Proyecto, error) {
	url := fmt.Sprintf("%s/api/proyectos/%d", c.BaseURL, id)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud: %v", err)
	}
	
	c.AuthFunc(req)
	req.Header.Set("X-Tenant-Id", "1")
	
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

// CreateProyecto creates a new project
func (c *Client) CreateProyecto(request CreateProyectoRequest) (*Proyecto, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}
	
	req, err := http.NewRequest("POST", c.BaseURL+"/api/proyectos", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Id", "1")
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
	
	var proyecto Proyecto
	if err := json.NewDecoder(resp.Body).Decode(&proyecto); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	
	return &proyecto, nil
}

// UpdateProyecto updates an existing project
func (c *Client) UpdateProyecto(id int, request UpdateProyectoRequest) (*Proyecto, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}
	
	url := fmt.Sprintf("%s/api/proyectos/%d", c.BaseURL, id)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Id", "1")
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
	
	var proyecto Proyecto
	if err := json.NewDecoder(resp.Body).Decode(&proyecto); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	
	return &proyecto, nil
}

// DeleteProyecto soft deletes a project
func (c *Client) DeleteProyecto(id int) error {
	url := fmt.Sprintf("%s/api/proyectos/%d", c.BaseURL, id)
	
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	
	req.Header.Set("X-Tenant-Id", "1")
	c.AuthFunc(req)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error calling API: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// RestoreProyecto restores a soft-deleted project
func (c *Client) RestoreProyecto(id int) error {
	url := fmt.Sprintf("%s/api/proyectos/%d/restore", c.BaseURL, id)
	
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	
	req.Header.Set("X-Tenant-Id", "1")
	c.AuthFunc(req)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error calling API: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// CreateCotizacionFromProyecto creates a cotización from a project
func (c *Client) CreateCotizacionFromProyecto(proyectoId int, request CreateCotizacionRequest) (*Cotizacion, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}
	
	url := fmt.Sprintf("%s/api/proyectos/%d/crear-cotizacion", c.BaseURL, proyectoId)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-Id", "1")
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
	
	var cotizacion Cotizacion
	if err := json.NewDecoder(resp.Body).Decode(&cotizacion); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}
	
	return &cotizacion, nil
}

// GetBaseURL gets the base URL from config, works for both CLI and Wails contexts
func GetBaseURL() (string, error) {
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
				if config.URL.AdminAPI != "" {
					return strings.TrimSuffix(config.URL.AdminAPI, "/"), nil
				}
			}
		}
	}
	
	// Fallback to environment variable or default
	baseURL := os.Getenv("ADMAPP_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	return strings.TrimSuffix(baseURL, "/"), nil
}

