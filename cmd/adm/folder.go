package adm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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
type service struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

type project struct {
	ID             int    `json:"id"`
	NombreProyecto string `json:"nombre_proyecto"`
}

type servicio struct {
	ID     int    `json:"id"`
	Nombre string `json:"nombre"`
}

type cotizaciondetalle struct {
	Servicio servicio `json:"servicio"`
	Proyecto project  `json:"proyecto"`
}

// Tipo para las cotizaciones con datos relacionados
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

// Get all services from PostgREST API
func obtenerServicios() ([]service, error) {
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

	var services []service
	err = json.NewDecoder(resp.Body).Decode(&services)
	if err != nil {
		return nil, err
	}

	return services, nil
}

// Get quotation data from PostgREST API
func obtenerDatosDeCotizacion(cotizacionID int) (*cotizaciondetalle, error) {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		return nil, fmt.Errorf("error getting API URL")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/cotizacion?select=servicio(id,nombre),proyecto(id,nombre_proyecto)&id=eq.%d",
			postgrestURL, cotizacionID), nil)
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

	var cotizaciones []cotizaciondetalle
	err = json.NewDecoder(resp.Body).Decode(&cotizaciones)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	if len(cotizaciones) == 0 {
		return nil, fmt.Errorf("no se encontró cotización con ID %d. Verifique que el ID sea correcto y que la cotización tenga servicio y proyecto asignados", cotizacionID)
	}

	return &cotizaciones[0], nil
}

// Search for clients by name
func buscarClientes(nombre string) ([]Cliente, error) {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		return nil, fmt.Errorf("error getting API URL")
	}

	// Add wildcards for PostgreSQL ILIKE pattern matching and URL encode
	searchPattern := fmt.Sprintf("*%s*", nombre)
	encodedPattern := url.QueryEscape(searchPattern)

	client := &http.Client{}
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/cliente?select=id,nombre,nombre_comercial&or=(nombre.ilike.%s,nombre_comercial.ilike.%s)",
			postgrestURL, encodedPattern, encodedPattern), nil)
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
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		return nil, fmt.Errorf("error getting API URL")
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/cotizacion?select=id,fecha,servicio(id,nombre),proyecto(id,nombre_proyecto)&id_cliente=eq.%d&estado=neq.GENERADA&order=id.desc",
			postgrestURL, clienteID), nil)
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
func crearCarpetaProyecto(cotizacion int) (string, error) {
	nombreProyecto, err := nombreCarpetaProyecto(cotizacion)
	if err != nil {
		return "", err
	}

	carpetaProyectos := viper.GetString("carpetas.proyectos")
	if carpetaProyectos == "" {
		return "", fmt.Errorf("carpeta de proyectos no configurada")
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	carpetaProyectos = filepath.Join(homedir, carpetaProyectos)

	rutaProyecto := filepath.Join(carpetaProyectos, nombreProyecto)

	carpetas, err := cargarEsquemas("Proyectos")
	if err != nil {
		return "", err
	}

	err = crearCarpeta(carpetas, rutaProyecto)
	if err != nil {
		return "", err
	}

	return rutaProyecto, nil
}

// --- Estados del menú ---
type menuState int

const (
	menuMain menuState = iota
	menuInputID
	menuInputCliente
	menuSelectCliente
	menuSelectCotizacion
	menuInputUploadPath
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
			"Descargar schema.json",
			"Subir schema.json",
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
				// Descargar schema.json
				err := downloadSchema()
				if err != nil {
					m.msg = fmt.Sprintf("Error: %s", err)
				} else {
					m.msg = inputs.SuccessStyle.Render("Schema descargado exitosamente")
				}
				m.state = menuSuccess
				return m, nil
			case 3:
				// Subir schema.json
				m.textInput = inputs.TextInput("Ingrese la ruta del archivo schema.json:", "/ruta/al/schema.json")
				m.state = menuInputUploadPath
				return m, nil
			case 4:
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
	case menuInputUploadPath:
		ti, cmd := m.textInput.Update(msg)
		m.textInput = ti.(inputs.TextInputModel)
		if key, ok := msg.(tea.KeyMsg); ok && key.Type == tea.KeyEnter {
			filePath := m.textInput.TextInput.Value()
			if filePath == "" {
				m.msg = "Operación cancelada"
				m.state = menuSuccess
				return m, nil
			}
			err := uploadSchema(filePath)
			if err != nil {
				m.msg = fmt.Sprintf("Error: %s", err)
			} else {
				m.msg = inputs.SuccessStyle.Render("Schema subido exitosamente")
			}
			m.state = menuSuccess
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
	case menuInputID, menuInputCliente, menuInputUploadPath:
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

// downloadSchema descarga el archivo schema.json desde el servidor
func downloadSchema() error {
	baseURL, _ := InitializeApi()
	if baseURL == "" {
		return fmt.Errorf("URL base no configurada")
	}

	// Construir la URL del endpoint con el sufijo assets
	downloadURL := fmt.Sprintf("%s/assets/configs/adm/schema.json", baseURL)

	// Hacer la petición HTTP
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("error al conectar con el servidor: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error al descargar schema (status %d): archivo no encontrado o no accesible", resp.StatusCode)
	}

	// Crear la carpeta de destino
	configPath := viper.GetString("config_path")
	if configPath == "" {
		return fmt.Errorf("ruta de configuración no encontrada (config_path)")
	}

	folderPath := filepath.Join(configPath, "folder")
	err = os.MkdirAll(folderPath, 0755)
	if err != nil {
		return fmt.Errorf("error al crear carpeta de destino: %v", err)
	}

	// Crear el archivo de destino
	destPath := filepath.Join(folderPath, "schema.json")
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error al crear archivo de destino: %v", err)
	}
	defer destFile.Close()

	// Copiar el contenido
	_, err = io.Copy(destFile, resp.Body)
	if err != nil {
		return fmt.Errorf("error al guardar archivo: %v", err)
	}

	fmt.Println(inputs.InfoStyle.Render(fmt.Sprintf("Schema guardado en: %s", destPath)))
	return nil
}

// uploadSchema sube un archivo schema.json al servidor
func uploadSchema(filePath string) error {
	baseURL, _ := InitializeApi()
	if baseURL == "" {
		return fmt.Errorf("URL base no configurada")
	}

	// Verificar que el archivo existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("el archivo no existe: %s", filePath)
	}

	// Abrir el archivo
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error al abrir archivo: %v", err)
	}
	defer file.Close()

	// Crear el buffer y multipart writer
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Agregar el archivo al formulario
	part, err := writer.CreateFormFile("file", "schema.json")
	if err != nil {
		return fmt.Errorf("error al crear formulario: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("error al copiar archivo: %v", err)
	}

	// Cerrar el writer
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("error al cerrar formulario: %v", err)
	}

	// Construir la URL del endpoint con el sufijo assets
	uploadURL := fmt.Sprintf("%s/assets/configs/adm", baseURL)

	// Crear la petición HTTP
	req, err := http.NewRequest("POST", uploadURL, &requestBody)
	if err != nil {
		return fmt.Errorf("error al crear petición: %v", err)
	}

	// Establecer headers
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Agregar parámetros de query
	q := req.URL.Query()
	q.Add("file_id", "schema.json")
	q.Add("description", "Schema configuration for folder structure")
	req.URL.RawQuery = q.Encode()

	// Hacer la petición
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error al enviar archivo: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error al subir archivo (status %d): %s", resp.StatusCode, string(body))
	}

	fmt.Println(inputs.InfoStyle.Render(fmt.Sprintf("Archivo subido desde: %s", filePath)))
	return nil
}

// DownloadCmd para descargar el schema.json
var DownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download schema.json from server",
	Long:  `Download the schema.json file from the configuration server and save it locally.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := downloadSchema()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}
		fmt.Println(inputs.SuccessStyle.Render("Schema descargado exitosamente"))
	},
}

// UploadCmd para subir el schema.json
var UploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload schema.json to server",
	Long:  `Upload a schema.json file to the configuration server.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Usar textinput para obtener la ruta del archivo
		p := tea.NewProgram(inputs.TextInput("Ingrese la ruta del archivo schema.json:", "/ruta/al/schema.json"), tea.WithAltScreen())
		m, err := p.Run()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}

		textInputModel, ok := m.(inputs.TextInputModel)
		if !ok {
			fmt.Println("Error: no se pudo obtener la ruta del archivo")
			return
		}

		filePath := textInputModel.TextInput.Value()
		if filePath == "" {
			fmt.Println("Operación cancelada.")
			return
		}

		err = uploadSchema(filePath)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return
		}
		fmt.Println(inputs.SuccessStyle.Render("Schema subido exitosamente"))
	},
}

func init() {
	AdmCmd.AddCommand(FolderCmd)
	FolderCmd.AddCommand(DownloadCmd)
	FolderCmd.AddCommand(UploadCmd)
}
