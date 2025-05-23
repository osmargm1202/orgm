package adm

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Implementación personalizada para list.Item
type customItem struct {
	id          string
	title       string
	description string
}

func (i customItem) FilterValue() string { return i.title }
func (i customItem) Title() string       { return i.title }
func (i customItem) Description() string { return i.description }

type Cotizacion struct {
	ID            int     `json:"id,omitempty"`
	IDCliente     int     `json:"id_cliente"`
	IDProyecto    int     `json:"id_proyecto"`
	IDServicio    int     `json:"id_servicio"`
	Moneda        string  `json:"moneda"`
	Fecha         string  `json:"fecha"`
	TasaMoneda    float64 `json:"tasa_moneda"`
	TiempoEntrega string  `json:"tiempo_entrega"`
	Avance        string  `json:"avance"`
	Validez       int     `json:"validez"`
	Estado        string  `json:"estado"`
	Idioma        string  `json:"idioma"`
	Descripcion   string  `json:"descripcion"`
	Retencion     string  `json:"retencion"`
	Subtotal      float64 `json:"subtotal"`
	Indirectos    float64 `json:"indirectos"`
	DescuentoP    float64 `json:"descuentop"`
	DescuentoM    float64 `json:"descuentom"`
	RetencionP    float64 `json:"retencionp"`
	RetencionM    float64 `json:"retencionm"`
	ItbisP        float64 `json:"itbisp"`
	ItbisM        float64 `json:"itbism"`
	Total         float64 `json:"total"`
}

type Presupuesto struct {
	ID           int                    `json:"id,omitempty"`
	IDCotizacion int                    `json:"id_cotizacion"`
	Presupuesto  map[string]interface{} `json:"presupuesto"`
}

type item struct {
	ID          string  `json:"id"`
	Item        string  `json:"item"`
	Total       float64 `json:"total"`
	Moneda      string  `json:"moneda"`
	Precio      float64 `json:"precio"`
	Unidad      string  `json:"unidad"`
	Cantidad    float64 `json:"cantidad"`
	Descripcion string  `json:"descripcion"`
}

type parentItem struct {
	ID          string  `json:"id"`
	Item        string  `json:"item"`
	Total       float64 `json:"total"`
	Moneda      string  `json:"moneda"`
	Precio      string  `json:"precio"`
	Unidad      string  `json:"unidad"`
	Cantidad    float64 `json:"cantidad"`
	Children    []item  `json:"children"`
	Categoria   string  `json:"categoria"`
	Descripcion string  `json:"descripcion"`
}

type presupuestoStruct struct {
	Indirectos  []interface{} `json:"indirectos"`
	Presupuesto []parentItem  `json:"presupuesto"`
}

var cotCmd = &cobra.Command{
	Use:   "quotation",
	Short: "Crear una nueva cotización",
	Long:  `Crear una nueva cotización`,
	Run: func(cmd *cobra.Command, args []string) {
		crearCotizacion()
	},
}


func init() {
	NewCmd.AddCommand(cotCmd)
}

func crearCotizacion() {
	// Buscar cliente
	cliente := BuscarCliente()
	if cliente.ID == 0 {
		fmt.Println("No se seleccionó ningún cliente")
		return
	}

	// Buscar proyecto
	proyecto := BuscarProyecto()
	if proyecto.ID == 0 {
		fmt.Println("No se seleccionó ningún proyecto")
		return
	}

	// Buscar servicio
	servicio := BuscarServicio()
	if servicio.ID == 0 {
		fmt.Println("No se seleccionó ningún servicio")
		return
	}

	// Crear la aplicación Fyne
	a := app.New()
	w := a.NewWindow("Nueva Cotización")
	w.Resize(fyne.NewSize(800, 600))

	// Crear bindings para los campos de entrada
	monedaBinding := binding.NewString()
	tiempoEntregaBinding := binding.NewString()
	avanceBinding := binding.NewString()
	validezBinding := binding.NewString()
	estadoBinding := binding.NewString()
	idiomaBinding := binding.NewString()
	descripcionBinding := binding.NewString()
	retencionBinding := binding.NewString()
	descripcionPresupuestoBinding := binding.NewString()
	precioBinding := binding.NewString()
	cantidadBinding := binding.NewString()

	// Valores por defecto
	monedaBinding.Set("RD$")
	tiempoEntregaBinding.Set("3")
	avanceBinding.Set("60")
	validezBinding.Set("30")
	estadoBinding.Set("GENERADA")
	idiomaBinding.Set("ES")
	descripcionBinding.Set("")
	retencionBinding.Set("NINGUNA")

	// Bindings para cálculos
	subtotalBinding := binding.NewFloat()
	itbisBinding := binding.NewFloat()
	retencionMBinding := binding.NewFloat()
	totalBinding := binding.NewFloat()

	// Función para calcular totales
	calcularTotales := func() {
		precioStr, _ := precioBinding.Get()
		cantidadStr, _ := cantidadBinding.Get()
		retencionStr, _ := retencionBinding.Get()

		precio, err := strconv.ParseFloat(precioStr, 64)
		if err != nil {
			precio = 0
		}

		cantidad, err := strconv.ParseFloat(cantidadStr, 64)
		if err != nil {
			cantidad = 0
		}

		subtotal := math.Round((precio*cantidad)*100) / 100
		subtotalBinding.Set(subtotal)

		itbisP := 18.0 // ITBIS fijo al 18%
		itbisM := math.Round((subtotal*(itbisP/100))*100) / 100
		itbisBinding.Set(itbisM)

		var retencionP float64
		if retencionStr != "NINGUNA" {
			retencionP = 30.0
		}

		retencionM := math.Round((retencionP*itbisM/100)*100) / 100
		retencionMBinding.Set(retencionM)

		total := math.Round((subtotal+itbisM-retencionM)*100) / 100
		totalBinding.Set(total)
	}

	// Sección de información del cliente
	clienteLabel := widget.NewLabel(fmt.Sprintf("Cliente: %s", cliente.Nombre))
	clienteLabel.TextStyle = fyne.TextStyle{Bold: true}

	proyectoLabel := widget.NewLabel(fmt.Sprintf("Proyecto: %s", proyecto.Nombre))
	servicioLabel := widget.NewLabel(fmt.Sprintf("Servicio: %s", servicio.Nombre))

	// Cargar logo del cliente si existe
	var logoContainer *fyne.Container
	imgURL := viper.GetString("uri.img")
	if imgURL != "" {
		logoURL := fmt.Sprintf("%s/images/logos/%d", imgURL, cliente.ID)

		// Intenta cargar el logo
		res, err := http.Get(logoURL)
		if err == nil && res.StatusCode == 200 {
			img := canvas.NewImageFromReader(res.Body, "logo-cliente")
			img.FillMode = canvas.ImageFillContain
			img.SetMinSize(fyne.NewSize(200, 100))

			logoContainer = container.NewCenter(img)
		} else {
			// Si no hay logo, mostrar mensaje
			logoContainer = container.NewCenter(widget.NewLabel("Logo no disponible"))
		}
	} else {
		logoContainer = container.NewCenter(widget.NewLabel("URL de imágenes no configurada"))
	}

	// Crear widgets para la entrada de datos
	monedaSelect := widget.NewSelect([]string{"RD$", "USD"}, func(selected string) {
		monedaBinding.Set(selected)
		calcularTotales()
	})
	monedaSelect.SetSelected("RD$")

	tiempoEntregaEntry := widget.NewEntryWithData(tiempoEntregaBinding)
	avanceEntry := widget.NewEntryWithData(avanceBinding)
	validezEntry := widget.NewEntryWithData(validezBinding)

	estadoSelect := widget.NewSelect([]string{"GENERADA", "TERMINADA", "APROBADA"}, func(selected string) {
		estadoBinding.Set(selected)
	})
	estadoSelect.SetSelected("GENERADA")

	idiomaEntry := widget.NewEntryWithData(idiomaBinding)
	descripcionEntry := widget.NewMultiLineEntry()
	descripcionEntry.Bind(descripcionBinding)

	retencionSelect := widget.NewSelect([]string{"NINGUNA", "HONORARIOS PROFESIONALES"}, func(selected string) {
		retencionBinding.Set(selected)
		calcularTotales()
	})
	retencionSelect.SetSelected("NINGUNA")

	// Widgets para el presupuesto
	descripcionPresupuestoEntry := widget.NewMultiLineEntry()
	descripcionPresupuestoEntry.Bind(descripcionPresupuestoBinding)

	precioEntry := widget.NewEntryWithData(precioBinding)
	precioEntry.OnChanged = func(s string) {
		calcularTotales()
	}

	cantidadEntry := widget.NewEntryWithData(cantidadBinding)
	cantidadEntry.SetText("1")
	cantidadEntry.OnChanged = func(s string) {
		calcularTotales()
	}

	// Widgets para mostrar totales
	subtotalLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(subtotalBinding, "Subtotal: %.2f"))
	itbisLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(itbisBinding, "ITBIS (18%%): %.2f"))
	retencionLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(retencionMBinding, "Retención: %.2f"))
	totalLabel := widget.NewLabelWithData(binding.FloatToStringWithFormat(totalBinding, "Total: %.2f"))
	totalLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Botón de guardar
	guardarBtn := widget.NewButton("Guardar Cotización", func() {
		// Obtener valores de los bindings
		moneda, _ := monedaBinding.Get()
		tiempoEntrega, _ := tiempoEntregaBinding.Get()
		avance, _ := avanceBinding.Get()
		validezStr, _ := validezBinding.Get()
		estado, _ := estadoBinding.Get()
		idioma, _ := idiomaBinding.Get()
		descripcion, _ := descripcionBinding.Get()
		retencion, _ := retencionBinding.Get()
		descripcionPresupuesto, _ := descripcionPresupuestoBinding.Get()

		validez, _ := strconv.Atoi(validezStr)

		subtotal, _ := subtotalBinding.Get()
		itbisM, _ := itbisBinding.Get()
		retencionM, _ := retencionMBinding.Get()
		total, _ := totalBinding.Get()

		precioStr, _ := precioBinding.Get()
		cantidadStr, _ := cantidadBinding.Get()
		precio, _ := strconv.ParseFloat(precioStr, 64)
		cantidad, _ := strconv.ParseFloat(cantidadStr, 64)

		// Crear cotización
		cotizacion := Cotizacion{
			IDCliente:     cliente.ID,
			IDProyecto:    proyecto.ID,
			IDServicio:    servicio.ID,
			Fecha:         time.Now().Format("01/02/2006"),
			TiempoEntrega: tiempoEntrega,
			Avance:        avance,
			Validez:       validez,
			Estado:        estado,
			Idioma:        idioma,
			Descripcion:   descripcion,
			Retencion:     retencion,
			Subtotal:      subtotal,
			Indirectos:    0,
			DescuentoP:    0,
			DescuentoM:    0,
			RetencionP:    retencionM * 100 / itbisM, // Calculamos el porcentaje
			RetencionM:    retencionM,
			ItbisP:        18.0,
			ItbisM:        itbisM,
			Total:         total,
		}

		// Guardar cotización
		idCotizacion := guardarCotizacion(cotizacion)
		if idCotizacion == 0 {
			dialog := widget.NewLabel("Error al guardar la cotización")
			dialog.Alignment = fyne.TextAlignCenter
			container := container.NewVBox(dialog, widget.NewButton("Cerrar", func() {
				w.Close()
			}))
			w.SetContent(container)
			return
		}

		// Crear estructura del presupuesto
		childItem := item{
			ID:          uuid.New().String()[:6],
			Item:        "P-1",
			Total:       subtotal,
			Moneda:      moneda,
			Precio:      precio,
			Unidad:      "Ud.",
			Cantidad:    cantidad,
			Descripcion: descripcionPresupuesto,
		}

		parentItemObj := parentItem{
			ID:          uuid.New().String()[:6],
			Item:        "I-1",
			Total:       subtotal,
			Moneda:      "",
			Precio:      "",
			Unidad:      "Ud.",
			Cantidad:    1.0,
			Children:    []item{childItem},
			Categoria:   "cat1",
			Descripcion: "SERVICIO",
		}

		presupuestoObj := presupuestoStruct{
			Indirectos:  []interface{}{},
			Presupuesto: []parentItem{parentItemObj},
		}

		// Guardar presupuesto
		presupuestoData := Presupuesto{
			IDCotizacion: idCotizacion,
			Presupuesto:  map[string]interface{}{"indirectos": presupuestoObj.Indirectos, "presupuesto": presupuestoObj.Presupuesto},
		}

		success := GuardarPresupuestoConID(presupuestoData, idCotizacion)

		if success {
			dialog := widget.NewLabel(fmt.Sprintf("Cotización creada con éxito. ID: %d", idCotizacion))
			dialog.Alignment = fyne.TextAlignCenter

			printBtn := widget.NewButton("Imprimir Cotización", func() {
				go func() {
					filePath, err := GetCotizacion(idCotizacion, false)
					if err != nil {
						fmt.Println("Error al imprimir cotización:", err)
					} else {
						fmt.Printf("Cotización %d guardada en %s\n", idCotizacion, filePath)
					}
				}()
			})

			closeBtn := widget.NewButton("Cerrar", func() {
				w.Close()
			})

			container := container.NewVBox(dialog, printBtn, closeBtn)
			w.SetContent(container)
		} else {
			dialog := widget.NewLabel("Error al guardar el presupuesto")
			dialog.Alignment = fyne.TextAlignCenter
			container := container.NewVBox(dialog, widget.NewButton("Cerrar", func() {
				w.Close()
			}))
			w.SetContent(container)
		}
	})

	// Organizar la interfaz en contenedores
	headerContainer := container.NewVBox(
		logoContainer,
		clienteLabel,
		proyectoLabel,
		servicioLabel,
		widget.NewSeparator(),
	)

	formContainer := container.NewGridWithColumns(2,
		widget.NewLabel("Moneda:"), monedaSelect,
		widget.NewLabel("Tiempo de entrega:"), tiempoEntregaEntry,
		widget.NewLabel("Avance (%):"), avanceEntry,
		widget.NewLabel("Validez (días):"), validezEntry,
		widget.NewLabel("Estado:"), estadoSelect,
		widget.NewLabel("Idioma:"), idiomaEntry,
		widget.NewLabel("Retención:"), retencionSelect,
	)

	descripcionContainer := container.NewBorder(
		widget.NewLabel("Descripción:"), nil, nil, nil,
		descripcionEntry,
	)

	presupuestoContainer := container.NewVBox(
		widget.NewLabel("Detalles del presupuesto:"),
		container.NewGridWithColumns(2,
			widget.NewLabel("Descripción:"), descripcionPresupuestoEntry,
			widget.NewLabel("Precio:"), precioEntry,
			widget.NewLabel("Cantidad:"), cantidadEntry,
		),
	)

	totalesContainer := container.NewVBox(
		widget.NewSeparator(),
		subtotalLabel,
		itbisLabel,
		retencionLabel,
		totalLabel,
	)

	// Contenedor principal
	content := container.NewVBox(
		headerContainer,
		container.NewPadded(formContainer),
		container.NewPadded(descripcionContainer),
		container.NewPadded(presupuestoContainer),
		container.NewPadded(totalesContainer),
		container.NewCenter(guardarBtn),
	)

	// Poner el contenido en un scroll en caso de que la ventana sea pequeña
	scrollContainer := container.NewScroll(content)
	w.SetContent(scrollContainer)

	// Inicializar cálculos
	calcularTotales()

	// Mostrar la ventana
	w.ShowAndRun()
}

// Cliente search model
type clienteModel struct {
	textInput       textinput.Model
	list            list.Model
	clientes        []Cliente
	err             error
	searchMode      bool
	selectedCliente Cliente
}

func initialClienteModel() clienteModel {
	ti := textinput.New()
	ti.Placeholder = "Ingrese término de búsqueda (nombre, número, nombre comercial)..."
	ti.Focus()
	ti.Width = 80

	return clienteModel{
		textInput:  ti,
		searchMode: true,
	}
}

func (m clienteModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m clienteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			if m.searchMode {
				// In search mode, perform the search
				query := m.textInput.Value()
				if query != "" {
					m.searchMode = false
					return m, SearchClientes(query)
				}
			} else {
				// In list mode, select an item
				if m.list.Items() != nil && len(m.list.Items()) > 0 {
					selectedItem, ok := m.list.SelectedItem().(customItem)
					if ok {
						id, _ := strconv.Atoi(selectedItem.id)
						for _, c := range m.clientes {
							if c.ID == id {
								m.selectedCliente = c
								return m, tea.Quit
							}
						}
					}
				}
			}
		}

	case clientesSearchResult:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}

		m.clientes = msg.clientes
		items := make([]list.Item, len(m.clientes))
		for i, c := range m.clientes {
			description := c.Representante + " | " + c.TipoFactura
			if c.NombreComercial != "" {
				description = c.NombreComercial + " | " + description
			}
			if c.Numero != "" {
				description = "Núm: " + c.Numero + " | " + description
			}

			items[i] = customItem{
				id:          strconv.Itoa(c.ID),
				title:       c.Nombre,
				description: description,
			}
		}

		// Mejorar la visualización con un delegate personalizado
		delegate := list.NewDefaultDelegate()
		delegate.ShowDescription = true
		delegate.SetHeight(6) // Dos líneas por elemento para mejorar visibilidad

		// Crear una lista con tamaño apropiado
		width, height := 80, 20 // Establecer un tamaño apropiado
		l := list.New(items, delegate, width, height)
		l.Title = "Resultados de la búsqueda"
		l.SetShowStatusBar(true)
		l.SetFilteringEnabled(true)
		l.SetShowHelp(true)
		l.Styles.Title = l.Styles.Title.Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#0000FF"))

		m.list = l

		return m, nil
	}

	if m.searchMode {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	} else if m.list.Items() != nil {
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m clienteModel) View() string {
	if m.err != nil {
		errorStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#FF0000")).
			PaddingLeft(2).
			PaddingRight(2).
			PaddingTop(1).
			PaddingBottom(1).
			Width(60).
			Align(lipgloss.Center)

		return "\n" + errorStyle.Render("ERROR") + "\n\n" +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render(m.err.Error()) + "\n\n" +
			"Presione ESC para volver"
	}

	if m.searchMode {
		style := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#0000AA")).
			PaddingLeft(2).
			PaddingRight(2).
			PaddingTop(1).
			PaddingBottom(1).
			Width(60).
			Align(lipgloss.Center)

		return "\n" + style.Render("BÚSQUEDA DE CLIENTES") + "\n\n" +
			"Ingrese el nombre, número o nombre comercial del cliente y presione ENTER\n\n" +
			m.textInput.View() + "\n\n" +
			"Presione ESC para cancelar"
	}

	if m.list.Items() != nil {
		if m.list.Items() == nil || len(m.list.Items()) == 0 {
			return "No se encontraron resultados. Presione ESC para volver."
		}
		return m.list.View()
	}

	return "Cargando..."
}

func seleccionarMoneda() string {
	items := []inputs.Item{
		{ID: "RD$", Name: "RD$", Desc: "Peso Dominicano"},
		{ID: "USD", Name: "USD", Desc: "Dólar Estadounidense"},
	}

	selectedItem := inputs.SelectList("Seleccione una moneda", items)
	return selectedItem.ID
}

func seleccionarEstado() string {
	items := []inputs.Item{
		{ID: "GENERADA", Name: "GENERADA", Desc: ""},
		{ID: "TERMINADA", Name: "TERMINADA", Desc: ""},
		{ID: "APROBADA", Name: "APROBADA", Desc: ""},
	}

	selectedItem := inputs.SelectList("Seleccione un estado", items)
	return selectedItem.ID
}

func seleccionarRetencion() string {
	items := []inputs.Item{
		{ID: "NINGUNA", Name: "NINGUNA", Desc: ""},
		{ID: "HP", Name: "HONORARIOS PROFESIONALES", Desc: "Impuesto de Trabajos de Servicios Profesionales"},
	}

	selectedItem := inputs.SelectList("Seleccione retención", items)
	return selectedItem.ID
}

func configurarCampo(etiqueta, valorDefecto string) string {
	return inputs.GetInputWithDefault(etiqueta, valorDefecto)
}

// ObtenerProximoIDCotizacion consulta el ID máximo existente y devuelve el siguiente
func ObtenerProximoIDCotizacion() (int, error) {
	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		return 0, fmt.Errorf("URL de PostgREST no configurada")
	}

	// Consultar la cotización con el ID más alto
	url := postgrestURL + "/cotizacion?select=id&order=id.desc&limit=1"

	resp, err := MakeRequest("GET", url, headers, nil)
	if err != nil {
		return 0, fmt.Errorf("error al consultar ID máximo: %w", err)
	}

	// Imprimir respuesta para depuración
	fmt.Println("Respuesta al buscar ID máximo:", string(resp))

	// Si no hay datos, comenzar desde 1
	if len(resp) == 0 || string(resp) == "[]" || string(resp) == "null" {
		fmt.Println("No se encontraron cotizaciones existentes, iniciando desde ID 1")
		return 1, nil
	}

	// Intentar deserializar como array
	var cotizaciones []map[string]interface{}
	err = json.Unmarshal(resp, &cotizaciones)

	if err != nil {
		// Si falla como array, intentar como objeto único
		var cotizacion map[string]interface{}
		err = json.Unmarshal(resp, &cotizacion)
		if err != nil {
			return 0, fmt.Errorf("error al procesar respuesta JSON: %w", err)
		}

		// Extraer ID del objeto único, manejando diferentes tipos
		return extraerID(cotizacion["id"])
	}

	if len(cotizaciones) == 0 {
		return 1, nil
	}

	// Extraer ID del primer elemento del array
	proximoID, err := extraerID(cotizaciones[0]["id"])
	if err != nil {
		return 0, err
	}

	fmt.Println("Próximo ID de cotización:", proximoID)
	return proximoID, nil
}

// extraerID intenta convertir varios tipos de datos a un entero y devuelve ese valor + 1
func extraerID(idValue interface{}) (int, error) {
	var idInt int

	switch v := idValue.(type) {
	case float64:
		idInt = int(v)
	case float32:
		idInt = int(v)
	case int:
		idInt = v
	case int64:
		idInt = int(v)
	case int32:
		idInt = int(v)
	case string:
		// Intentar convertir string a int
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("no se pudo convertir ID string '%s' a entero: %w", v, err)
		}
		idInt = parsed
	case nil:
		return 1, nil // Si es nil, empezar desde 1
	default:
		return 0, fmt.Errorf("tipo de ID no soportado: %T", idValue)
	}

	return idInt + 1, nil
}

func guardarCotizacion(cotizacion Cotizacion) int {
	// Obtener el próximo ID disponible
	proximoID, err := ObtenerProximoIDCotizacion()
	if err != nil {
		fmt.Println("Error al obtener próximo ID:", err)
		return 0
	}

	// Asignar el ID a la cotización
	cotizacion.ID = proximoID

	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		fmt.Println("Error: URL de PostgREST no configurada")
		return 0
	}

	url := postgrestURL + "/cotizacion"

	cotizacionBytes, err := json.Marshal(cotizacion)
	if err != nil {
		fmt.Println("Error al serializar cotización:", err)
		return 0
	}

	resp, err := MakeRequest("POST", url, headers, cotizacionBytes)
	if err != nil {
		fmt.Println("Error al guardar cotización:", err)
		return 0
	}

	var cotizaciones []Cotizacion
	if err := json.Unmarshal(resp, &cotizaciones); err != nil {
		// Intentar deserializar como objeto único
		var cotizacionResp Cotizacion
		if err := json.Unmarshal(resp, &cotizacionResp); err != nil {
			fmt.Println("Error al procesar respuesta:", err)
			return 0
		}
		return cotizacionResp.ID
	}

	if len(cotizaciones) > 0 {
		return cotizaciones[0].ID
	}

	return proximoID // Retornar el ID asignado incluso si no se pudo verificar
}

func MakeRequest(method, url string, headers map[string]string, body []byte) ([]byte, error) {
	client := &http.Client{}

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, strings.NewReader(string(body)))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("error creando request: %w", err)
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando request: %w", err)
	}
	defer resp.Body.Close()

	// Verificar código de respuesta
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	return data, nil
}

// Usar el mismo ID para el presupuesto que para la cotización
func GuardarPresupuestoConID(presupuesto Presupuesto, idCotizacion int) bool {
	// Asignar el ID de la cotización al presupuesto
	presupuesto.ID = idCotizacion

	postgrestURL, headers := InitializePostgrest()
	if postgrestURL == "" {
		fmt.Println("Error: URL de PostgREST no configurada")
		return false
	}

	url := postgrestURL + "/presupuesto"

	presupuestoBytes, err := json.Marshal(presupuesto)
	if err != nil {
		fmt.Println("Error al serializar presupuesto:", err)
		return false
	}

	resp, err := MakeRequest("POST", url, headers, presupuestoBytes)
	if err != nil {
		fmt.Println("Error al guardar presupuesto:", err)
		return false
	}

	fmt.Println("Presupuesto guardado con ID:", idCotizacion)
	fmt.Println("Respuesta:", string(resp))

	return true
}
