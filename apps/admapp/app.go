package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/osmargm1202/orgm/cmd"
	"github.com/osmargm1202/orgm/pkg/admappapi"
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
		fmt.Printf("Error getting base URL: %v\n", err)
		baseURL = "http://localhost:8000" // fallback
	}

	// Create auth function that uses cmd.EnsureGCloudIDToken
	authFunc := func(req *http.Request) {
		token, err := cmd.EnsureGCloudIDToken()
		if err != nil || token == "" {
			fmt.Printf("Warning: No se pudo obtener token de autenticación: %v\n", err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Create client
	client := admappapi.NewClient(baseURL, authFunc)

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
	clientes, err := a.client.GetClientes(incluirInactivos)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
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

// GetLogoURL gets the logo URL for a client
func (a *App) GetLogoURL(id int) map[string]interface{} {
	logoResp, err := a.client.GetClienteLogoURL(id)
	if err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}
	}
	return map[string]interface{}{"success": true, "data": logoResp}
}

// OpenFile opens a file with the default application
func (a *App) OpenFile(filePath string) map[string]interface{} {
	// For now, just return success - file opening can be implemented later
	return map[string]interface{}{"success": true, "filepath": filePath}
}
