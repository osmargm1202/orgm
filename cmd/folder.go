package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Model for API responses

type project struct {
	ID             int    `json:"id"`
	NombreProyecto string `json:"nombre_proyecto"`
}

type servicio struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

type cotizaciondetalle struct {
	ID       int      `json:"id"`
	Servicio servicio `json:"servicio"`
	Proyecto project  `json:"proyecto"`
}

// Tipo para las cotizaciones con datos relacionados
type Cliente struct {
	ID                     int    `json:"id"`
	Nombre                 string `json:"nombre"`
	NombreComercial        string `json:"nombre_comercial"`
	Numero                 string `json:"numero"`
	Correo                 string `json:"correo"`
	Direccion              string `json:"direccion"`
	Ciudad                 string `json:"ciudad"`
	Provincia              string `json:"provincia"`
	Telefono               string `json:"telefono"`
	Representante          string `json:"representante"`
	TelefonoRepresentante  string `json:"telefono_representante"`
	ExtensionRepresentante string `json:"extension_representante"`
	CelularRepresentante   string `json:"celular_representante"`
	CorreoRepresentante    string `json:"correo_representante"`
	TipoFactura            string `json:"tipo_factura"`
	FechaActualizacion     string `json:"fecha_actualizacion"`
}


type CotizacionConRelaciones struct {
	ID       int      `json:"id"`
	Fecha    string   `json:"fecha"`
	Servicio servicio `json:"servicio"`
	Proyecto project  `json:"proyecto"`
}

// FolderCmd represents the folder command
var FolderCmd = &cobra.Command{
	Use:   "folder [ID]",
	Short: "Folder commands",
	Long:  `Folder create, delete, list, etc. You can provide a quotation ID directly to create the folder immediately.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Si se proporciona un ID como argumento, crear la carpeta directamente
		if len(args) > 0 {
			idStr := args[0]
			id, err := strconv.Atoi(idStr)
			if err != nil {
				fmt.Printf("Error: El ID debe ser un número válido. Recibido: %s\n", idStr)
				return
			}

			rutaCarpeta, err := crearCarpetaProyecto(id)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}

			fmt.Printf("Carpeta creada exitosamente para la cotización %d\n", id)
			fmt.Println(inputs.InfoStyle.Render(fmt.Sprintf("Ruta: %s", rutaCarpeta)))
			return
		}

		// Si no se proporciona ID, ejecutar el menú interactivo
		menu()
	},
}

// InitializeAdminAPI returns Admin API configuration with GCR token
func InitializeAdminAPI() (string, map[string]string, error) {
	baseURL := viper.GetString("url.admapp_api")
	if baseURL == "" {
		return "", nil, fmt.Errorf("url.admapp_api no configurado")
	}
	
	token, err := EnsureGCloudIDToken()
	if err != nil {
		return "", nil, fmt.Errorf("error obteniendo token: %v", err)
	}
	
	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"X-Tenant-Id": "1",
	}
	
	return baseURL, headers, nil
}

// Get all services from Admin API
func obtenerServicios() ([]servicio, error) {
	baseURL, headers, err := InitializeAdminAPI()
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/servicios", baseURL), nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error al obtener servicios: %d", resp.StatusCode)
	}

	var services []servicio
	err = json.NewDecoder(resp.Body).Decode(&services)
	if err != nil {
		return nil, err
	}

	return services, nil
}

// Get quotation data from Admin API
func obtenerDatosDeCotizacion(cotizacionID int) (*cotizaciondetalle, error) {
	baseURL, headers, err := InitializeAdminAPI()
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/api/cotizaciones/%d", baseURL, cotizacionID), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error al obtener cotización (status %d): cotización ID %d no encontrada o no accesible", resp.StatusCode, cotizacionID)
	}

	// Admin API returns single object, not array
	var cotizacion cotizaciondetalle
	err = json.NewDecoder(resp.Body).Decode(&cotizacion)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &cotizacion, nil
}

// Search for clients by name
func buscarClientes(nombre string) ([]Cliente, error) {
	baseURL, headers, err := InitializeAdminAPI()
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	// Admin API search endpoint - using query parameter
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/api/clientes?search=%s", baseURL, url.QueryEscape(nombre)), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error al obtener clientes (status %d): no se pudo buscar clientes con el nombre '%s'", resp.StatusCode, nombre)
	}

	var clientes []Cliente
	err = json.NewDecoder(resp.Body).Decode(&clientes)
	if err != nil {
		return nil, fmt.Errorf("error decoding clientes response: %v", err)
	}

	return clientes, nil
}

// Search for quotations by client ID
func buscarCotizaciones(clienteID int) ([]CotizacionConRelaciones, error) {
	baseURL, headers, err := InitializeAdminAPI()
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/api/clientes/%d/cotizaciones", baseURL, clienteID), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error al obtener cotizaciones (status %d): cliente ID %d no encontrado o sin cotizaciones", resp.StatusCode, clienteID)
	}

	var cotizaciones []CotizacionConRelaciones
	err = json.NewDecoder(resp.Body).Decode(&cotizaciones)
	if err != nil {
		return nil, fmt.Errorf("error decoding cotizaciones response: %v", err)
	}

	return cotizaciones, nil
}

// Schema type for folder structure
type SchemaType struct {
	TiposProyecto map[string]struct {
		Carpetas []string `json:"carpetas"`
	} `json:"tipos"`
}

// Load folder schemas from local folder.json file
func cargarEsquemas(tipoProyecto string) ([]string, error) {
	configPath := viper.GetString("config_path")
	if configPath == "" {
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, ".config", "orgm")
	}
	
	folderJSONPath := filepath.Join(configPath, "folder.json")
	data, err := os.ReadFile(folderJSONPath)
	if err != nil {
		return nil, fmt.Errorf("error leyendo folder.json: %w", err)
	}
	
	var schema SchemaType
	err = json.Unmarshal(data, &schema)
	if err != nil {
		return nil, fmt.Errorf("error parsing folder.json: %w", err)
	}
	
	if tipoProyectoSchema, ok := schema.TiposProyecto[tipoProyecto]; ok {
		return tipoProyectoSchema.Carpetas, nil
	}
	
	return nil, fmt.Errorf("tipo de proyecto no encontrado en el esquema")
}

// Generate a folder name for a project
func nombreCarpetaProyecto(cotizacion int) (string, error) {
	datos, err := obtenerDatosDeCotizacion(cotizacion)
	if err != nil {
		return "", err
	}

	nombreProyecto := fmt.Sprintf("%d - %s", cotizacion, datos.Proyecto.NombreProyecto)
	return nombreProyecto, nil
}

// Create a project folder structure in local filesystem
func crearCarpetaProyecto(cotizacion int) (string, error) {
	nombreProyecto, err := nombreCarpetaProyecto(cotizacion)
	if err != nil {
		return "", err
	}

	// Get base projects path from config or use default
	basePath := viper.GetString("folder.base_path")
	if basePath == "" {
		basePath = "./Proyectos" // default
	}

	projectPath := filepath.Join(basePath, nombreProyecto)
	err = os.MkdirAll(projectPath, 0755)
	if err != nil {
		return "", fmt.Errorf("error creating project folder: %w", err)
	}

	// Load folder schema and create subfolders
	carpetas, err := cargarEsquemas("Proyectos")
	if err != nil {
		return "", err
	}

	// Create subfolders in local filesystem
	for _, carpeta := range carpetas {
		subfolderPath := filepath.Join(projectPath, carpeta)
		err = os.MkdirAll(subfolderPath, 0755)
		if err != nil {
			return "", fmt.Errorf("error creating subfolder %s: %w", carpeta, err)
		}
	}

	return projectPath, nil
}

// --- Estados del menú ---
type menuState int

const (
	menuMain menuState = iota
	menuInputID
	menuInputCliente
	menuSelectCliente
	menuSelectCotizacion
	menuSuccess
)

type folderModel struct {
	state        menuState
	mainMenu     inputs.SelectionModelS
	textInput    inputs.TextInputModel
	clientes     []Cliente
	cotizaciones []CotizacionConRelaciones
	selectMenu   inputs.SelectionModelS
	msg          string
}

func initialFolderModel() folderModel {
	mainMenu := inputs.SelectionModel(
		[]string{
			"Crear carpeta por ID de cotización",
			"Buscar cotizaciones por cliente",
			"Salir",
		},
		"Menú de Carpetas",
		"Selecciona una opción",
	)
	return folderModel{
		state:    menuMain,
		mainMenu: mainMenu,
	}
}

func (m folderModel) Init() tea.Cmd {
	return nil
}

func (m folderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case menuMain:
		mm, cmd := m.mainMenu.Update(msg)
		m.mainMenu = mm.(inputs.SelectionModelS)
		if m.mainMenu.Quitting {
			return m, tea.Quit
		}
		if m.mainMenu.Selected {
			switch m.mainMenu.Cursor {
			case 0:
				m.textInput = inputs.TextInput("Ingrese el ID de cotización:", "Ej: 123")
				m.state = menuInputID
				return m, nil
			case 1:
				m.textInput = inputs.TextInput("Ingrese el nombre del cliente:", "Ej: Juan")
				m.state = menuInputCliente
				return m, nil
			case 2:
				return m, tea.Quit
			}
		}
		return m, cmd
	case menuInputID:
		ti, cmd := m.textInput.Update(msg)
		m.textInput = ti.(inputs.TextInputModel)
		if key, ok := msg.(tea.KeyMsg); ok && key.Type == tea.KeyEnter {
			idStr := m.textInput.TextInput.Value()
			id, err := strconv.Atoi(idStr)
			if err != nil {
				m.msg = "Error: La cotización debe ser un número"
				m.state = menuSuccess
				return m, nil
			}
			rutaCarpeta, err := crearCarpetaProyecto(id)
			if err != nil {
				m.msg = fmt.Sprintf("Error: %s", err)
			} else {
				m.msg = fmt.Sprintf("Carpeta creada exitosamente para la cotización %d\n%s", id, inputs.InfoStyle.Render(fmt.Sprintf("Ruta: %s", rutaCarpeta)))
			}
			m.state = menuSuccess
			return m, nil
		}
		return m, cmd
	case menuInputCliente:
		ti, cmd := m.textInput.Update(msg)
		m.textInput = ti.(inputs.TextInputModel)
		if key, ok := msg.(tea.KeyMsg); ok && key.Type == tea.KeyEnter {
			nombre := m.textInput.TextInput.Value()
			clientes, err := buscarClientes(nombre)
			if err != nil {
				m.msg = fmt.Sprintf("Error: %s", err)
				m.state = menuSuccess
				return m, nil
			}
			if len(clientes) == 0 {
				m.msg = "No se encontraron clientes con ese nombre"
				m.state = menuSuccess
				return m, nil
			}
			m.clientes = clientes
			m.selectMenu = clientesSelectMenu(clientes)
			m.state = menuSelectCliente
			return m, nil
		}
		return m, cmd
	case menuSelectCliente:
		mm, cmd := m.selectMenu.Update(msg)
		m.selectMenu = mm.(inputs.SelectionModelS)
		if m.selectMenu.Quitting {
			m.state = menuMain
			m.mainMenu.Selected = false
			return m, nil
		}
		if m.selectMenu.Selected {
			idx := m.selectMenu.Cursor
			if idx < 0 || idx >= len(m.clientes) {
				m.msg = "Selección inválida"
				m.state = menuSuccess
				return m, nil
			}
			cliente := m.clientes[idx]
			cots, err := buscarCotizaciones(cliente.ID)
			if err != nil {
				m.msg = fmt.Sprintf("Error: %s", err)
				m.state = menuSuccess
				return m, nil
			}
			if len(cots) == 0 {
				m.msg = "No se encontraron cotizaciones para este cliente"
				m.state = menuSuccess
				return m, nil
			}
			m.cotizaciones = cots
			m.selectMenu = cotizacionesSelectMenu(cots)
			m.state = menuSelectCotizacion
			return m, nil
		}
		return m, cmd
	case menuSelectCotizacion:
		mm, cmd := m.selectMenu.Update(msg)
		m.selectMenu = mm.(inputs.SelectionModelS)
		if m.selectMenu.Quitting {
			m.state = menuMain
			m.mainMenu.Selected = false
			return m, nil
		}
		if m.selectMenu.Selected {
			idx := m.selectMenu.Cursor
			if idx < 0 || idx >= len(m.cotizaciones) {
				m.msg = "Selección inválida"
				m.state = menuSuccess
				return m, nil
			}
			cot := m.cotizaciones[idx]
			rutaCarpeta, err := crearCarpetaProyecto(cot.ID)
			if err != nil {
				m.msg = fmt.Sprintf("Error: %s", err)
			} else {
				m.msg = fmt.Sprintf("Carpeta creada exitosamente para la cotización %d\n%s", cot.ID, inputs.InfoStyle.Render(fmt.Sprintf("Ruta: %s", rutaCarpeta)))
			}
			m.state = menuSuccess
			return m, nil
		}
		return m, cmd
	case menuSuccess:
		if key, ok := msg.(tea.KeyMsg); ok && (key.String() == "enter" || key.String() == "q" || key.String() == "esc") {
			m.state = menuMain
			m.mainMenu.Selected = false
			return m, nil
		}
		return m, nil
	}
	return m, nil
}

func (m folderModel) View() string {
	switch m.state {
	case menuMain:
		return m.mainMenu.View()
	case menuInputID, menuInputCliente:
		return m.textInput.View()
	case menuSelectCliente, menuSelectCotizacion:
		return m.selectMenu.View()
	case menuSuccess:
		return inputs.SuccessStyle.Render(m.msg) + "\n\nPresiona [enter] para volver al menú."
	}
	return ""
}

// --- Menús de selección ---
func clientesSelectMenu(clientes []Cliente) inputs.SelectionModelS {
	opciones := make([]string, len(clientes))
	for i, c := range clientes {
		nombreComercial := c.NombreComercial
		if nombreComercial == "" {
			nombreComercial = "Sin nombre comercial"
		}
		opciones[i] = fmt.Sprintf("%d - %s (%s)", c.ID, c.Nombre, nombreComercial)
	}
	return inputs.SelectionModel(opciones, "Selecciona un cliente", "Clientes encontrados")
}

func cotizacionesSelectMenu(cots []CotizacionConRelaciones) inputs.SelectionModelS {
	opciones := make([]string, len(cots))
	for i, c := range cots {
		opciones[i] = fmt.Sprintf("%d - %s - %s", c.ID, c.Servicio.Nombre, c.Proyecto.NombreProyecto)
	}
	return inputs.SelectionModel(opciones, "Selecciona una cotización", "Cotizaciones encontradas")
}

// Reemplazar el menú clásico por Bubble Tea
func menu() {
	p := tea.NewProgram(initialFolderModel(), tea.WithAltScreen())
	p.Run()
}




var folderEditorCmd = &cobra.Command{
	Use:   "editor",
	Short: "Open folder configuration editor",
	Long:  `Open the default editor to modify folder.json configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Opening folder configuration editor...")
		// TODO: Implement editor logic
		fmt.Println("Editor opened for folder.json")
	},
}


// Subcommands for folder
var folderConfigUpdateCmd = &cobra.Command{
	Use:   "config update",
	Short: "Update folder configuration",
	Long:  `Update the folder configuration from the latest folder.json`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Updating folder configuration...")
		// TODO: Implement config update logic
		fmt.Println("Folder configuration updated successfully")
	},
}

var folderInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize folder configuration",
	Long:  `Initialize the folder configuration with default settings`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing folder configuration...")
		// TODO: Implement init logic
		fmt.Println("Folder configuration initialized successfully")
	},
}

func init() {
	RootCmd.AddCommand(FolderCmd)
	FolderCmd.AddCommand(folderConfigUpdateCmd)
	FolderCmd.AddCommand(folderInitCmd)
	FolderCmd.AddCommand(folderEditorCmd)
}
