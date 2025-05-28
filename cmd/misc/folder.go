package misc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Model for API responses
type Service struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

type Project struct {
	ID             int    `json:"id"`
	NombreProyecto string `json:"nombre_proyecto"`
}

type Cliente struct {
	ID              int    `json:"id"`
	Nombre          string `json:"nombre"`
	NombreComercial string `json:"nombre_comercial"`
}

type Servicio struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

type CotizacionDetalle struct {
	Servicio Servicio `json:"servicio"`
	Proyecto Project  `json:"proyecto"`
}

type Cotizacion struct {
	ID       int      `json:"id"`
	Fecha    string   `json:"fecha"`
	Servicio Servicio `json:"servicio"`
	Proyecto Project  `json:"proyecto"`
}

// FolderCmd represents the folder command
var FolderCmd = &cobra.Command{
	Use:   "folder",
	Short: "Folder commands",
	Long:  `Folder create, delete, list, etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		menu()
	},
}

// Get all services from PostgREST API
func obtenerServicios() ([]Service, error) {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		return nil, fmt.Errorf("error getting API URL")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/servicio?select=id,nombre", postgrestURL), nil)
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

	var services []Service
	err = json.NewDecoder(resp.Body).Decode(&services)
	if err != nil {
		return nil, err
	}

	return services, nil
}

// Get quotation data from PostgREST API
func obtenerDatosDeCotizacion(cotizacionID int) (*CotizacionDetalle, error) {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		return nil, fmt.Errorf("error getting API URL")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/cotizacion?select=servicio(id,nombre),proyecto(id,nombre_proyecto)&id=eq.%d",
			postgrestURL, cotizacionID), nil)
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
		return nil, fmt.Errorf("error al obtener cotización: %d", resp.StatusCode)
	}

	var cotizaciones []CotizacionDetalle
	err = json.NewDecoder(resp.Body).Decode(&cotizaciones)
	if err != nil {
		return nil, err
	}

	if len(cotizaciones) == 0 {
		return nil, fmt.Errorf("no se encontró cotización con ID %d", cotizacionID)
	}

	return &cotizaciones[0], nil
}

// Search for clients by name
func buscarClientes(nombre string) ([]Cliente, error) {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		return nil, fmt.Errorf("error getting API URL")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/cliente?select=id,nombre,nombre_comercial&or=(nombre.ilike.%s,nombre_comercial.ilike.%s)",
			postgrestURL, nombre, nombre), nil)
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
		return nil, fmt.Errorf("error al obtener clientes: %d", resp.StatusCode)
	}

	var clientes []Cliente
	err = json.NewDecoder(resp.Body).Decode(&clientes)
	if err != nil {
		return nil, err
	}

	return clientes, nil
}

// Search for quotations by client ID
func buscarCotizaciones(clienteID int) ([]Cotizacion, error) {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		return nil, fmt.Errorf("error getting API URL")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/cotizacion?select=id,fecha,servicio(id,nombre),proyecto(id,nombre_proyecto)&id_cliente=eq.%d&estado=neq.GENERADA&order=id.desc",
			postgrestURL, clienteID), nil)
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
		return nil, fmt.Errorf("error al obtener cotizaciones: %d", resp.StatusCode)
	}

	var cotizaciones []Cotizacion
	err = json.NewDecoder(resp.Body).Decode(&cotizaciones)
	if err != nil {
		return nil, err
	}

	return cotizaciones, nil
}

// Schema type for folder structure
type SchemaType struct {
	TiposProyecto map[string]struct {
		Carpetas []string `json:"carpetas"`
	} `json:"tipos_proyecto"`
}

// Load folder schemas from JSON file
func cargarEsquemas(tipoProyecto string) ([]string, error) {
	configPath := viper.GetString("config_path")
	schemaPath := filepath.Join(configPath, "folder", "schema.json")

	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}

	var schema SchemaType
	err = json.Unmarshal(data, &schema)
	if err != nil {
		return nil, err
	}

	if tipoProyectoSchema, ok := schema.TiposProyecto[tipoProyecto]; ok {
		return tipoProyectoSchema.Carpetas, nil
	}

	return nil, fmt.Errorf("tipo de proyecto no encontrado en el esquema")
}

// Create folder structure based on schema
func crearCarpeta(esquema []string, ruta string) error {
	for _, carpeta := range esquema {
		rutaCarpeta := filepath.Join(ruta, carpeta)
		err := os.MkdirAll(rutaCarpeta, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// Generate a folder name for a project
func nombreCarpetaProyecto(cotizacion int) (string, error) {
	datos, err := obtenerDatosDeCotizacion(cotizacion)
	if err != nil {
		return "", err
	}

	nombreProyecto := fmt.Sprintf("%d - %s - %s", cotizacion, datos.Servicio.Nombre, datos.Proyecto.NombreProyecto)
	return nombreProyecto, nil
}

// Create a project folder structure
func crearCarpetaProyecto(cotizacion int) error {
	nombreProyecto, err := nombreCarpetaProyecto(cotizacion)
	if err != nil {
		return err
	}

	carpetaProyectos := viper.GetString("carpetas.proyectos")
	if carpetaProyectos == "" {
		return fmt.Errorf("carpeta de proyectos no configurada")
	}

	rutaProyecto := filepath.Join(carpetaProyectos, nombreProyecto)

	carpetas, err := cargarEsquemas("Proyectos")
	if err != nil {
		return err
	}

	err = crearCarpeta(carpetas, rutaProyecto)
	if err != nil {
		return err
	}

	return nil
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
	cotizaciones []Cotizacion
	selectMenu   inputs.SelectionModelS
	msg          string
}

func initialFolderModel() folderModel {
	mainMenu := inputs.SelectionModel(
		[]string{"Crear carpeta por ID de cotización", "Buscar cotizaciones por cliente", "Salir"},
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
			err = crearCarpetaProyecto(id)
			if err != nil {
				m.msg = fmt.Sprintf("Error: %s", err)
			} else {
				m.msg = fmt.Sprintf("Carpeta creada exitosamente para la cotización %d", id)
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
			err := crearCarpetaProyecto(cot.ID)
			if err != nil {
				m.msg = fmt.Sprintf("Error: %s", err)
			} else {
				m.msg = fmt.Sprintf("Carpeta creada exitosamente para la cotización %d", cot.ID)
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

func cotizacionesSelectMenu(cots []Cotizacion) inputs.SelectionModelS {
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

func init() {
	MiscCmd.AddCommand(FolderCmd)
}
