package adm

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"

	admClient "github.com/osmargm1202/orgm/cmd/adm"
)

// RNCResult representa un resultado de búsqueda en la base de datos RNC
type RNCResult struct {
	RNC                  string
	RazonSocial          string
	ActividadEconomica   string
	FechaInicioActividad string
	Estado               string
	RegimenPago          string
}

// EditarClienteWidget encapsula la funcionalidad de edición/creación de clientes
type EditarClienteWidget struct {
	container *fyne.Container
	window    fyne.Window

	// Bindings para los campos
	idBinding                     binding.String
	nombreBinding                 binding.String
	nombreComercialBinding        binding.String
	numeroBinding                 binding.String
	correoBinding                 binding.String
	direccionBinding              binding.String
	ciudadBinding                 binding.String
	provinciaBinding              binding.String
	telefonoBinding               binding.String
	representanteBinding          binding.String
	telefonoRepresentanteBinding  binding.String
	extensionRepresentanteBinding binding.String
	celularRepresentanteBinding   binding.String
	correoRepresentanteBinding    binding.String
	tipoFacturaBinding            binding.String

	// Cliente actual
	clienteActual *admClient.Cliente
	modoEdicion   bool

	// Callbacks
	onClienteGuardado func(cliente admClient.Cliente)
	onCancelado       func()

	// Widgets
	logoContainer     *fyne.Container
	logoButton        *widget.Button
	logoPath          string
	tipoFacturaSelect *widget.Select
	logoImage         *canvas.Image
	logoLabel         *widget.Label

	// Búsqueda RNC
	busquedaRNCEntry  *widget.Entry
	busquedaRNCButton *widget.Button
	rncResultsTable   *widget.Table
	rncResults        []RNCResult
	rncContainer      *fyne.Container
}

// NewEditarClienteWidget crea un nuevo widget para editar/crear clientes
func NewEditarClienteWidget(
	window fyne.Window,
	cliente *admClient.Cliente,
	onClienteGuardado func(cliente admClient.Cliente),
	onCancelado func()) *EditarClienteWidget {

	e := &EditarClienteWidget{
		window:            window,
		clienteActual:     cliente,
		modoEdicion:       cliente != nil,
		onClienteGuardado: onClienteGuardado,
		onCancelado:       onCancelado,
	}

	e.inicializarBindings()
	e.createUI()

	if cliente != nil {
		e.cargarDatosCliente(*cliente)
		// Cargar logo después de que todo esté inicializado
		e.cargarLogoExistente()
	}

	return e
}

func (e *EditarClienteWidget) inicializarBindings() {
	e.idBinding = binding.NewString()
	e.nombreBinding = binding.NewString()
	e.nombreComercialBinding = binding.NewString()
	e.numeroBinding = binding.NewString()
	e.correoBinding = binding.NewString()
	e.direccionBinding = binding.NewString()
	e.ciudadBinding = binding.NewString()
	e.provinciaBinding = binding.NewString()
	e.telefonoBinding = binding.NewString()
	e.representanteBinding = binding.NewString()
	e.telefonoRepresentanteBinding = binding.NewString()
	e.extensionRepresentanteBinding = binding.NewString()
	e.celularRepresentanteBinding = binding.NewString()
	e.correoRepresentanteBinding = binding.NewString()
	e.tipoFacturaBinding = binding.NewString()
}

func (e *EditarClienteWidget) createUI() {
	// Widgets de entrada
	idEntry := widget.NewEntryWithData(e.idBinding)
	idEntry.Disable() // ID no editable

	nombreEntry := widget.NewEntryWithData(e.nombreBinding)
	nombreComercialEntry := widget.NewEntryWithData(e.nombreComercialBinding)
	numeroEntry := widget.NewEntryWithData(e.numeroBinding)
	correoEntry := widget.NewEntryWithData(e.correoBinding)
	direccionEntry := widget.NewEntryWithData(e.direccionBinding)
	ciudadEntry := widget.NewEntryWithData(e.ciudadBinding)
	provinciaEntry := widget.NewEntryWithData(e.provinciaBinding)
	telefonoEntry := widget.NewEntryWithData(e.telefonoBinding)
	representanteEntry := widget.NewEntryWithData(e.representanteBinding)
	telefonoRepresentanteEntry := widget.NewEntryWithData(e.telefonoRepresentanteBinding)
	extensionRepresentanteEntry := widget.NewEntryWithData(e.extensionRepresentanteBinding)
	celularRepresentanteEntry := widget.NewEntryWithData(e.celularRepresentanteBinding)
	correoRepresentanteEntry := widget.NewEntryWithData(e.correoRepresentanteBinding)

	// Select para tipo de factura
	e.tipoFacturaSelect = widget.NewSelect(
		[]string{"NCFC", "NCF", "NCG", "NCRE"},
		func(selected string) {
			e.tipoFacturaBinding.Set(selected)
		},
	)

	// Sección de logo
	e.crearSeccionLogo()

	// Sección de búsqueda RNC
	e.crearSeccionBusquedaRNC()

	// Botones
	guardarBtn := widget.NewButton("Guardar", func() {
		e.guardarCliente()
	})
	guardarBtn.Importance = widget.SuccessImportance

	cancelarBtn := widget.NewButton("Cancelar", func() {
		if e.onCancelado != nil {
			e.onCancelado()
		}
	})

	botones := container.NewHBox(
		guardarBtn,
		cancelarBtn,
	)

	// Formulario principal
	form := container.NewVBox(
		widget.NewCard("Información Básica", "",
			container.NewGridWithColumns(2,
				widget.NewLabel("ID:"), idEntry,
				widget.NewLabel("Nombre:"), nombreEntry,
				widget.NewLabel("Nombre Comercial:"), nombreComercialEntry,
				widget.NewLabel("RNC/Número:"), numeroEntry,
				widget.NewLabel("Correo:"), correoEntry,
				widget.NewLabel("Teléfono:"), telefonoEntry,
			),
		),

		widget.NewCard("Dirección", "",
			container.NewGridWithColumns(2,
				widget.NewLabel("Dirección:"), direccionEntry,
				widget.NewLabel("Ciudad:"), ciudadEntry,
				widget.NewLabel("Provincia:"), provinciaEntry,
			),
		),

		widget.NewCard("Representante", "",
			container.NewGridWithColumns(2,
				widget.NewLabel("Nombre:"), representanteEntry,
				widget.NewLabel("Teléfono:"), telefonoRepresentanteEntry,
				widget.NewLabel("Extensión:"), extensionRepresentanteEntry,
				widget.NewLabel("Celular:"), celularRepresentanteEntry,
				widget.NewLabel("Correo:"), correoRepresentanteEntry,
			),
		),

		widget.NewCard("Configuración", "",
			container.NewGridWithColumns(2,
				widget.NewLabel("Tipo Factura:"), e.tipoFacturaSelect,
			),
		),

		widget.NewCard("Logo", "", e.logoContainer),

		// Botones después del logo y antes de búsqueda RNC
		botones,

		widget.NewCard("Búsqueda de Empresa (RNC)", "", e.rncContainer),
	)

	// Container principal con scroll
	scrollContent := container.NewVBox(form)
	scroll := container.NewScroll(scrollContent)

	e.container = container.NewBorder(nil, nil, nil, nil, scroll)
}

// crearSeccionBusquedaRNC crea la sección de búsqueda en la base de datos RNC
func (e *EditarClienteWidget) crearSeccionBusquedaRNC() {
	// Campo de búsqueda
	e.busquedaRNCEntry = widget.NewEntry()
	e.busquedaRNCEntry.SetPlaceHolder("Ingrese nombre de empresa o RNC...")

	// Botón de búsqueda
	e.busquedaRNCButton = widget.NewButton("Buscar", func() {
		e.buscarEnRNC()
	})

	// Tabla de resultados
	e.rncResultsTable = widget.NewTable(
		func() (int, int) {
			return len(e.rncResults), 2 // Solo mostrar RNC y Razón Social
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id.Row < len(e.rncResults) {
				switch id.Col {
				case 0:
					label.SetText(e.rncResults[id.Row].RNC)
				case 1:
					label.SetText(e.rncResults[id.Row].RazonSocial)
				}
			}
		},
	)

	// Configurar tabla
	e.rncResultsTable.SetColumnWidth(0, 120) // RNC
	e.rncResultsTable.SetColumnWidth(1, 300) // Razón Social

	// Configurar selección de la tabla
	e.rncResultsTable.OnSelected = func(id widget.TableCellID) {
		if id.Row < len(e.rncResults) {
			e.seleccionarEmpresaRNC(e.rncResults[id.Row])
		}
	}

	// Container con altura limitada para mostrar máximo 5 resultados
	tableContainer := container.NewScroll(e.rncResultsTable)
	tableContainer.SetMinSize(fyne.NewSize(400, 150))

	// Encabezados
	headers := container.NewHBox(
		widget.NewLabelWithStyle("RNC", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Razón Social", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)

	// Hacer el campo de búsqueda más ancho
	e.busquedaRNCEntry.Resize(fyne.NewSize(400, 0))

	searchContainer := container.NewBorder(nil, nil, nil, e.busquedaRNCButton, e.busquedaRNCEntry)

	e.rncContainer = container.NewVBox(
		searchContainer,
		widget.NewSeparator(),
		headers,
		tableContainer,
		widget.NewLabel("Haga clic en una empresa para prellenar los datos"),
	)
}

// buscarEnRNC busca empresas en la base de datos RNC
func (e *EditarClienteWidget) buscarEnRNC() {
	searchQuery := strings.TrimSpace(e.busquedaRNCEntry.Text)
	if searchQuery == "" {
		dialog.ShowInformation("Búsqueda vacía", "Por favor ingrese un término de búsqueda", e.window)
		return
	}

	results, err := e.searchInRNCDatabase(searchQuery)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Error al buscar en base de datos RNC: %v", err), e.window)
		return
	}

	// Limitar a 5 resultados
	if len(results) > 5 {
		results = results[:5]
	}

	e.rncResults = results
	e.rncResultsTable.Refresh()

	if len(results) == 0 {
		dialog.ShowInformation("Sin resultados", "No se encontraron empresas con ese término", e.window)
	}
}

// searchInRNCDatabase busca en la base de datos SQLite RNC
func (e *EditarClienteWidget) searchInRNCDatabase(searchQuery string) ([]RNCResult, error) {
	dbPath := filepath.Join(viper.GetString("carpetas.apps"), "rnc", "dgii.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Split the search query into individual terms
	terms := strings.Fields(searchQuery)
	if len(terms) == 0 {
		return nil, fmt.Errorf("no search terms provided")
	}

	// Build the SQL query to match any of the terms
	query := `SELECT RNC, RazonSocial, ActividadEconomica, FechaInicioActividad, Estado, RegimenPago 
              FROM tabla WHERE `

	var conditions []string
	args := []interface{}{}

	for _, term := range terms {
		termWithWildcards := "%" + term + "%"
		conditions = append(conditions, "(RNC LIKE ? OR RazonSocial LIKE ?)")
		args = append(args, termWithWildcards, termWithWildcards)
	}

	query += strings.Join(conditions, " AND ")
	query += " LIMIT 5" // Limit the results

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	var results []RNCResult
	for rows.Next() {
		var result RNCResult
		if err := rows.Scan(
			&result.RNC,
			&result.RazonSocial,
			&result.ActividadEconomica,
			&result.FechaInicioActividad,
			&result.Estado,
			&result.RegimenPago,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// seleccionarEmpresaRNC prellla los datos del cliente con la empresa seleccionada
func (e *EditarClienteWidget) seleccionarEmpresaRNC(empresa RNCResult) {
	// Preguntar al usuario si quiere prellenar los datos
	dialog.ShowConfirm("Prellenar datos",
		fmt.Sprintf("¿Desea prellenar los datos del cliente con la información de:\n%s\nRNC: %s?",
			empresa.RazonSocial, empresa.RNC),
		func(confirmed bool) {
			if confirmed {
				e.nombreBinding.Set(empresa.RazonSocial)
				e.numeroBinding.Set(empresa.RNC)
				// Opcional: también podemos prellenar el nombre comercial si está vacío
				nombreComercial, _ := e.nombreComercialBinding.Get()
				if nombreComercial == "" {
					e.nombreComercialBinding.Set(empresa.RazonSocial)
				}
			}
		}, e.window)
}

func (e *EditarClienteWidget) crearSeccionLogo() {
	// Botón para seleccionar logo
	e.logoButton = widget.NewButton("Seleccionar Logo", func() {
		e.seleccionarLogo()
	})

	// Label para mostrar el archivo seleccionado
	e.logoLabel = widget.NewLabel("Ningún archivo seleccionado")

	// Imagen del logo (inicialmente oculta)
	e.logoImage = canvas.NewImageFromResource(nil)
	e.logoImage.FillMode = canvas.ImageFillContain
	e.logoImage.SetMinSize(fyne.NewSize(300, 100)) // Máximo 300px ancho, 100px alto
	e.logoImage.Hide()                             // Ocultar inicialmente

	// Logo se cargará después de crear el container completo

	e.logoContainer = container.NewVBox(
		container.NewHBox(e.logoButton, e.logoLabel),
		e.logoImage,
	)
}

func (e *EditarClienteWidget) cargarLogoExistente() {
	// Verificar que tenemos cliente y que el container está inicializado
	if e.clienteActual == nil || e.logoContainer == nil || len(e.logoContainer.Objects) < 2 {
		return
	}

	imgURL := viper.GetString("url.img")
	if imgURL == "" {
		e.logoLabel.SetText("URL de imágenes no configurada")
		return
	}

	logoURL := fmt.Sprintf("%s/images/logos/%d", imgURL, e.clienteActual.ID)

	// Intentar cargar el logo
	resp, err := http.Get(logoURL)
	if err == nil && resp.StatusCode == 200 {
		// Crear imagen desde la respuesta HTTP
		img := canvas.NewImageFromReader(resp.Body, fmt.Sprintf("logo-%d", e.clienteActual.ID))
		img.FillMode = canvas.ImageFillContain
		img.SetMinSize(fyne.NewSize(300, 100))

		// Reemplazar la imagen en el container
		e.logoImage = img
		e.logoImage.Show()
		e.logoLabel.SetText(fmt.Sprintf("Logo actual del cliente %s", e.clienteActual.Nombre))

		// Actualizar el container de forma segura
		if len(e.logoContainer.Objects) >= 2 {
			e.logoContainer.Objects[1] = e.logoImage
			e.logoContainer.Refresh()
		}
	} else {
		e.logoLabel.SetText("No hay logo disponible")
		if e.logoImage != nil {
			e.logoImage.Hide()
		}
	}

	if resp != nil {
		resp.Body.Close()
	}
}

// mostrarVistaPrevia muestra una vista previa del logo seleccionado
func (e *EditarClienteWidget) mostrarVistaPrevia(filePath, fileName string) {
	// Cargar imagen desde el archivo temporal
	img := canvas.NewImageFromFile(filePath)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(300, 100))

	// Reemplazar la imagen en el container
	e.logoImage = img
	e.logoImage.Show()
	e.logoLabel.SetText(fmt.Sprintf("Nuevo logo seleccionado: %s", fileName))

	// Actualizar el container de forma segura
	if e.logoContainer != nil && len(e.logoContainer.Objects) >= 2 {
		e.logoContainer.Objects[1] = e.logoImage
		e.logoContainer.Refresh()
	}
}

func (e *EditarClienteWidget) seleccionarLogo() {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(fmt.Errorf("Error seleccionando archivo: %v", err), e.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		// Guardar temporalmente el archivo
		tempFile, err := os.CreateTemp("", "logo_*.png")
		if err != nil {
			dialog.ShowError(fmt.Errorf("Error creando archivo temporal: %v", err), e.window)
			return
		}
		defer tempFile.Close()

		_, err = io.Copy(tempFile, reader)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Error copiando archivo: %v", err), e.window)
			return
		}

		e.logoPath = tempFile.Name()

		// Mostrar vista previa del nuevo logo
		e.mostrarVistaPrevia(tempFile.Name(), reader.URI().Name())
	}, e.window)

	// Filtros de archivos
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg", ".gif"}))
	fileDialog.Show()
}

func (e *EditarClienteWidget) cargarDatosCliente(cliente admClient.Cliente) {
	e.idBinding.Set(strconv.Itoa(cliente.ID))
	e.nombreBinding.Set(cliente.Nombre)
	e.nombreComercialBinding.Set(cliente.NombreComercial)
	e.numeroBinding.Set(cliente.Numero)
	e.correoBinding.Set(cliente.Correo)
	e.direccionBinding.Set(cliente.Direccion)
	e.ciudadBinding.Set(cliente.Ciudad)
	e.provinciaBinding.Set(cliente.Provincia)
	e.telefonoBinding.Set(cliente.Telefono)
	e.representanteBinding.Set(cliente.Representante)
	e.telefonoRepresentanteBinding.Set(cliente.TelefonoRepresentante)
	e.extensionRepresentanteBinding.Set(cliente.ExtensionRepresentante)
	e.celularRepresentanteBinding.Set(cliente.CelularRepresentante)
	e.correoRepresentanteBinding.Set(cliente.CorreoRepresentante)
	e.tipoFacturaBinding.Set(cliente.TipoFactura)

	// Preseleccionar tipo de factura en el widget después de cargar los datos
	if e.tipoFacturaSelect != nil {
		e.tipoFacturaSelect.SetSelected(cliente.TipoFactura)
	}
}

func (e *EditarClienteWidget) guardarCliente() {
	// Obtener valores de los bindings
	nombre, _ := e.nombreBinding.Get()
	if nombre == "" {
		dialog.ShowError(fmt.Errorf("El nombre es obligatorio"), e.window)
		return
	}

	// Crear cliente con los datos del formulario
	cliente := admClient.Cliente{}

	if e.modoEdicion {
		cliente = *e.clienteActual
	} else {
		// Obtener próximo ID para cliente nuevo
		nextID, err := admClient.GetLastID("cliente")
		if err != nil {
			dialog.ShowError(fmt.Errorf("Error obteniendo ID: %v", err), e.window)
			return
		}
		cliente.ID = nextID
	}

	// Actualizar campos
	cliente.Nombre = nombre
	nombreComercial, _ := e.nombreComercialBinding.Get()
	cliente.NombreComercial = nombreComercial
	numero, _ := e.numeroBinding.Get()
	cliente.Numero = numero
	correo, _ := e.correoBinding.Get()
	cliente.Correo = correo
	direccion, _ := e.direccionBinding.Get()
	cliente.Direccion = direccion
	ciudad, _ := e.ciudadBinding.Get()
	cliente.Ciudad = ciudad
	provincia, _ := e.provinciaBinding.Get()
	cliente.Provincia = provincia
	telefono, _ := e.telefonoBinding.Get()
	cliente.Telefono = telefono
	representante, _ := e.representanteBinding.Get()
	cliente.Representante = representante
	telefonoRepresentante, _ := e.telefonoRepresentanteBinding.Get()
	cliente.TelefonoRepresentante = telefonoRepresentante
	extensionRepresentante, _ := e.extensionRepresentanteBinding.Get()
	cliente.ExtensionRepresentante = extensionRepresentante
	celularRepresentante, _ := e.celularRepresentanteBinding.Get()
	cliente.CelularRepresentante = celularRepresentante
	correoRepresentante, _ := e.correoRepresentanteBinding.Get()
	cliente.CorreoRepresentante = correoRepresentante
	tipoFactura, _ := e.tipoFacturaBinding.Get()
	cliente.TipoFactura = tipoFactura
	cliente.FechaActualizacion = time.Now().Format("02/01/2006")

	// Guardar cliente
	var clienteGuardado admClient.Cliente
	var err error

	if e.modoEdicion {
		clienteGuardado, err = e.actualizarCliente(cliente)
	} else {
		clienteGuardado = admClient.GuardarCliente(cliente)
		if clienteGuardado.ID == 0 {
			err = fmt.Errorf("error guardando cliente")
		}
	}

	if err != nil {
		dialog.ShowError(fmt.Errorf("Error guardando cliente: %v", err), e.window)
		return
	}

	// Subir logo si se seleccionó uno
	if e.logoPath != "" {
		e.subirLogo(clienteGuardado.ID)
	}

	// Notificar que se guardó
	if e.onClienteGuardado != nil {
		e.onClienteGuardado(clienteGuardado)
	}

	dialog.ShowInformation("Éxito",
		fmt.Sprintf("Cliente %s guardado exitosamente", clienteGuardado.Nombre),
		e.window)
}

func (e *EditarClienteWidget) actualizarCliente(cliente admClient.Cliente) (admClient.Cliente, error) {
	postgrestURL, headers := admClient.InitializePostgrest()
	if postgrestURL == "" {
		return admClient.Cliente{}, fmt.Errorf("URL de PostgREST no configurada")
	}

	url := fmt.Sprintf("%s/cliente?id=eq.%d", postgrestURL, cliente.ID)

	clienteBytes, err := json.Marshal(cliente)
	if err != nil {
		return admClient.Cliente{}, fmt.Errorf("error serializando cliente: %v", err)
	}

	// Agregar header para PATCH
	headers["Prefer"] = "return=representation"

	resp, err := admClient.MakeRequest("PATCH", url, headers, clienteBytes)
	if err != nil {
		return admClient.Cliente{}, fmt.Errorf("error actualizando cliente: %v", err)
	}

	var clientesResp []admClient.Cliente
	if err := json.Unmarshal(resp, &clientesResp); err != nil {
		return admClient.Cliente{}, fmt.Errorf("error procesando respuesta: %v", err)
	}

	if len(clientesResp) > 0 {
		return clientesResp[0], nil
	}

	return cliente, nil
}

func (e *EditarClienteWidget) subirLogo(clienteID int) {
	if e.logoPath == "" {
		return
	}

	// Abrir el archivo
	file, err := os.Open(e.logoPath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Error abriendo archivo de logo: %v", err), e.window)
		return
	}
	defer file.Close()

	// Crear request multipart
	body := &io.PipeReader{}
	writer := &io.PipeWriter{}
	body, writer = io.Pipe()

	multipartWriter := multipart.NewWriter(writer)

	go func() {
		defer writer.Close()
		defer multipartWriter.Close()

		// Agregar el archivo
		part, err := multipartWriter.CreateFormFile("image", fmt.Sprintf("logo_%d.png", clienteID))
		if err != nil {
			return
		}

		_, err = io.Copy(part, file)
		if err != nil {
			return
		}
	}()

	// Crear request HTTP usando la configuración de img
	imgURL, imgHeaders := admClient.InitializeImg()
	if imgURL == "" {
		dialog.ShowError(fmt.Errorf("URL de imágenes no configurada"), e.window)
		return
	}

	uploadURL := fmt.Sprintf("%s/imges/logos/%d", imgURL, clienteID)

	req, err := http.NewRequest("PATCH", uploadURL, body)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Error creando request: %v", err), e.window)
		return
	}

	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	// Agregar headers de autenticación de img
	for key, value := range imgHeaders {
		req.Header.Set(key, value)
	}

	// Ejecutar request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Error subiendo logo: %v", err), e.window)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		dialog.ShowError(fmt.Errorf("Error del servidor al subir logo: %d", resp.StatusCode), e.window)
		return
	}

	// Limpiar archivo temporal
	os.Remove(e.logoPath)
	e.logoPath = ""
}

// GetContainer retorna el container principal del widget
func (e *EditarClienteWidget) GetContainer() *fyne.Container {
	return e.container
}
