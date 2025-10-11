package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/osmargm1202/orgm/cmd"
	"github.com/osmargm1202/orgm/pkg/admappapi"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	client *admappapi.Client
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Get base URL
	baseURL, err := admappapi.GetBaseURL()
	if err != nil {
		fmt.Printf("⚠️ Error getting base URL: %v\n", err)
		baseURL = "http://localhost:8000" // fallback
	}
	fmt.Printf("🌐 Base URL configurada: %s\n", baseURL)

	// Create auth function that uses cmd.EnsureGCloudIDTokenForAudience
	authFunc := func(req *http.Request) {
		token, err := cmd.EnsureGCloudIDTokenForAudience(baseURL)
		if err != nil || token == "" {
			fmt.Printf("⚠️ Warning: No se pudo obtener token de autenticación: %v\n", err)
			return
		}
		fmt.Printf("🔑 Token obtenido: %s...\n", token[:20])
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Create client
	client := admappapi.NewClient(baseURL, authFunc)
	fmt.Printf("✅ Cliente API inicializado correctamente\n")

	return &App{
		client: client,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetClientes returns all clients
func (a *App) GetClientes(incluirInactivos bool) map[string]interface{} {
	fmt.Printf("🔍 GetClientes llamado con incluirInactivos: %v\n", incluirInactivos)
	clientes, err := a.client.GetClientes(incluirInactivos)
	if err != nil {
		fmt.Printf("❌ Error en GetClientes: %v\n", err)
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	fmt.Printf("✅ GetClientes exitoso, %d clientes encontrados\n", len(clientes))
	return map[string]interface{}{"success": true, "data": clientes}
}

// GetClienteByID returns a specific client by ID
func (a *App) GetClienteByID(id int) map[string]interface{} {
	cliente, err := a.client.GetClienteByID(id)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true, "data": cliente}
}

// CreateCliente creates a new client
func (a *App) CreateCliente(nombre, nombreComercial, numero, correo, direccion, ciudad, provincia, telefono, representante, correoRepresentante, tipoFactura string) map[string]interface{} {
	request := admappapi.CreateClienteRequest{
		Nombre:               nombre,
		NombreComercial:      nombreComercial,
		Numero:               numero,
		Correo:               correo,
		Direccion:            direccion,
		Ciudad:               ciudad,
		Provincia:            provincia,
		Telefono:             telefono,
		Representante:         representante,
		CorreoRepresentante:   correoRepresentante,
		TipoFactura:          tipoFactura,
	}
	
	cliente, err := a.client.CreateCliente(request)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true, "data": cliente}
}

// UpdateCliente updates an existing client
func (a *App) UpdateCliente(id int, nombre, nombreComercial, numero, correo, direccion, ciudad, provincia, telefono, representante, correoRepresentante, tipoFactura string) map[string]interface{} {
	request := admappapi.UpdateClienteRequest{
		Nombre:               nombre,
		NombreComercial:      nombreComercial,
		Numero:               numero,
		Correo:               correo,
		Direccion:            direccion,
		Ciudad:               ciudad,
		Provincia:            provincia,
		Telefono:             telefono,
		Representante:         representante,
		CorreoRepresentante:   correoRepresentante,
		TipoFactura:          tipoFactura,
	}
	
	cliente, err := a.client.UpdateCliente(id, request)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true, "data": cliente}
}

// DeleteCliente soft deletes a client
func (a *App) DeleteCliente(id int) map[string]interface{} {
	err := a.client.DeleteCliente(id)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true}
}

// RestoreCliente restores a soft-deleted client
func (a *App) RestoreCliente(id int) map[string]interface{} {
	err := a.client.RestoreCliente(id)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true}
}

// UploadLogo uploads a logo for a client
func (a *App) UploadLogo(id int, filePath string) map[string]interface{} {
	logoResp, err := a.client.UploadClienteLogo(id, filePath)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true, "data": logoResp}
}

// GetLogoURL gets the logo URL for a client and caches it locally
func (a *App) GetLogoURL(id int) map[string]interface{} {
	logoResp, err := a.client.GetClienteLogoURL(id)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	// Download and cache logo locally
	localPath, err := a.downloadAndCacheLogo(id, logoResp.URL)
	if err != nil {
		fmt.Printf("⚠️ Error caching logo: %v\n", err)
		// Return remote URL if caching fails
		return map[string]interface{}{"success": true, "data": map[string]interface{}{
			"path": logoResp.Path,
			"url":  logoResp.URL,
		}}
	}
	
	// Convert local file to base64 data URL
	base64URL, err := a.fileToBase64(localPath)
	if err != nil {
		fmt.Printf("⚠️ Error converting to base64: %v\n", err)
		// Return remote URL if conversion fails
		return map[string]interface{}{"success": true, "data": map[string]interface{}{
			"path": logoResp.Path,
			"url":  logoResp.URL,
		}}
	}
	
	// Return base64 data URL for preview
	return map[string]interface{}{"success": true, "data": map[string]interface{}{
		"path": logoResp.Path,
		"url":  base64URL,
	}}
}

// downloadAndCacheLogo downloads logo from URL and saves it locally
func (a *App) downloadAndCacheLogo(clienteId int, url string) (string, error) {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %v", err)
	}
	
	// Create cache directory: ~/.config/orgm/tenant/1/clientes/
	cacheDir := filepath.Join(homeDir, ".config", "orgm", "tenant", "1", "clientes")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("error creating cache directory: %v", err)
	}
	
	// Local file path
	localPath := filepath.Join(cacheDir, fmt.Sprintf("%d.png", clienteId))
	
	// Check if file already exists
	if _, err := os.Stat(localPath); err == nil {
		fmt.Printf("✅ Logo already cached: %s\n", localPath)
		return localPath, nil
	}
	
	// Download logo
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error downloading logo: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("error downloading logo: HTTP %d", resp.StatusCode)
	}
	
	// Create local file
	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("error creating local file: %v", err)
	}
	defer file.Close()
	
	// Copy content
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("error saving logo: %v", err)
	}
	
	fmt.Printf("✅ Logo cached successfully: %s\n", localPath)
	return localPath, nil
}

// fileToBase64 converts a local file to a base64 data URL
func (a *App) fileToBase64(filePath string) (string, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}
	
	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(data)
	
	// Create data URL (assuming PNG, you can detect mime type if needed)
	dataURL := "data:image/png;base64," + encoded
	
	fmt.Printf("✅ Converted to base64 (%d bytes → %d chars)\n", len(data), len(encoded))
	return dataURL, nil
}

// OpenFile opens a file dialog for logo selection
func (a *App) OpenFile() map[string]interface{} {
	file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Seleccionar Logo",
		Filters: []runtime.FileFilter{
			{DisplayName: "Images", Pattern: "*.png;*.jpg;*.jpeg;*.gif"},
		},
	})
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	if file == "" {
		return map[string]interface{}{"success": false, "error": "No file selected"}
	}
	return map[string]interface{}{"success": true, "data": file}
}

// GetProyectos returns projects for a specific client
func (a *App) GetProyectos(idCliente int, incluirInactivos bool) map[string]interface{} {
	fmt.Printf("🔍 GetProyectos llamado con idCliente: %d, incluirInactivos: %v\n", idCliente, incluirInactivos)
	proyectos, err := a.client.GetProyectos(idCliente, incluirInactivos)
	if err != nil {
		fmt.Printf("❌ Error en GetProyectos: %v\n", err)
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	fmt.Printf("✅ GetProyectos exitoso, %d proyectos encontrados\n", len(proyectos))
	return map[string]interface{}{"success": true, "data": proyectos}
}

// GetProyectoByID returns a specific project by ID
func (a *App) GetProyectoByID(id int) map[string]interface{} {
	proyecto, err := a.client.GetProyectoByID(id)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true, "data": proyecto}
}

// CreateProyecto creates a new project
func (a *App) CreateProyecto(idCliente int, nombreProyecto, ubicacion, descripcion string) map[string]interface{} {
	request := admappapi.CreateProyectoRequest{
		IDCliente:      idCliente,
		NombreProyecto: nombreProyecto,
		Ubicacion:      ubicacion,
		Descripcion:    descripcion,
	}
	
	proyecto, err := a.client.CreateProyecto(request)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true, "data": proyecto}
}

// UpdateProyecto updates an existing project
func (a *App) UpdateProyecto(id int, nombreProyecto, ubicacion, descripcion string) map[string]interface{} {
	request := admappapi.UpdateProyectoRequest{
		NombreProyecto: nombreProyecto,
		Ubicacion:      ubicacion,
		Descripcion:    descripcion,
	}
	
	proyecto, err := a.client.UpdateProyecto(id, request)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true, "data": proyecto}
}

// DeleteProyecto soft deletes a project
func (a *App) DeleteProyecto(id int) map[string]interface{} {
	err := a.client.DeleteProyecto(id)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true}
}

// RestoreProyecto restores a soft-deleted project
func (a *App) RestoreProyecto(id int) map[string]interface{} {
	err := a.client.RestoreProyecto(id)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true}
}

// CreateCotizacionFromProyecto creates a cotización from a project with default values
func (a *App) CreateCotizacionFromProyecto(proyectoId, idServicio int) map[string]interface{} {
	// Get project details first
	proyecto, err := a.client.GetProyectoByID(proyectoId)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	
	// Create cotización with default values
	request := admappapi.CreateCotizacionRequest{
		IDCliente:      proyecto.IDCliente,
		IDProyecto:     proyectoId,
		IDServicio:     idServicio,
		Moneda:         "RD$",
		Fecha:          time.Now().Format("2006-01-02"),
		TasaMoneda:     1.0,
		TiempoEntrega:  "30",
		Avance:         "60",
		Validez:        30,
		Estado:         "GENERADA",
		Idioma:         "ES",
		Descripcion:    "",
		Retencion:      "NINGUNA",
		Descuentop:     0.0,
		Retencionp:     0.0,
		Itbisp:         0.0,
	}
	
	cotizacion, err := a.client.CreateCotizacionFromProyecto(proyectoId, request)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true, "data": cotizacion}
}
