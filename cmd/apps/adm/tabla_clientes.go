package adm

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	admClient "github.com/osmargm1202/orgm/cmd/adm"
)

// ClienteRow representa una fila en la tabla de clientes
type ClienteRow struct {
	cliente admClient.Cliente
}

func (c ClienteRow) ToSlice() []string {
	return []string{
		strconv.Itoa(c.cliente.ID),
		c.cliente.Nombre,
		c.cliente.NombreComercial,
		c.cliente.Numero,
		c.cliente.Representante,
	}
}

// TablaClientesWidget encapsula toda la funcionalidad de búsqueda y tabla
type TablaClientesWidget struct {
	container *fyne.Container

	// Campos de búsqueda
	nombreEntry *widget.Entry
	rncEntry    *widget.Entry
	idEntry     *widget.Entry

	// Tabla
	table             *widget.Table
	clientes          []admClient.Cliente
	clientesFiltrados []admClient.Cliente

	// Callback para cuando se selecciona un cliente
	onClienteSelected func(cliente admClient.Cliente)
}

// NewTablaClientesWidget crea un nuevo widget de tabla de clientes
func NewTablaClientesWidget(onClienteSelected func(cliente admClient.Cliente)) *TablaClientesWidget {
	t := &TablaClientesWidget{
		onClienteSelected: onClienteSelected,
	}
	t.createUI()
	t.cargarTodosLosClientes()
	return t
}

func (t *TablaClientesWidget) createUI() {
	// Crear campos de búsqueda
	t.nombreEntry = widget.NewEntry()
	t.nombreEntry.SetPlaceHolder("Buscar por nombre o nombre comercial...")
	t.nombreEntry.OnChanged = func(text string) {
		t.filtrarClientes()
	}

	t.rncEntry = widget.NewEntry()
	t.rncEntry.SetPlaceHolder("Buscar por RNC...")
	t.rncEntry.OnChanged = func(text string) {
		t.filtrarClientes()
	}

	t.idEntry = widget.NewEntry()
	t.idEntry.SetPlaceHolder("Buscar por ID...")
	t.idEntry.OnChanged = func(text string) {
		t.filtrarClientes()
	}

	// Crear tabla
	t.table = widget.NewTable(
		func() (int, int) {
			return len(t.clientesFiltrados), 5 // 5 columnas: ID, Nombre, Nombre Comercial, RNC, Representante
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Wrapping = fyne.TextWrapWord // Habilitar wrap de texto
			return label
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id.Row < len(t.clientesFiltrados) {
				row := ClienteRow{cliente: t.clientesFiltrados[id.Row]}
				data := row.ToSlice()
				if id.Col < len(data) {
					label.SetText(data[id.Col])
				}
			}
		},
	)

	// Configurar anchos de columnas
	t.table.SetColumnWidth(0, 50)  // ID
	t.table.SetColumnWidth(1, 180) // Nombre
	t.table.SetColumnWidth(2, 180) // Nombre Comercial
	t.table.SetColumnWidth(3, 120) // RNC
	t.table.SetColumnWidth(4, 150) // Representante

	// Configurar eventos de selección
	t.table.OnSelected = func(id widget.TableCellID) {
		if id.Row < len(t.clientesFiltrados) && t.onClienteSelected != nil {
			t.onClienteSelected(t.clientesFiltrados[id.Row])
		}
	}

	// Crear encabezados de tabla
	headerContainer := container.NewHBox(
		widget.NewLabelWithStyle("ID", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Nombre", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Nombre Comercial", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("RNC", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Representante", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	)

	// Crear formulario de búsqueda
	searchForm := container.NewGridWithColumns(2,
		widget.NewLabel("Nombre:"), t.nombreEntry,
		widget.NewLabel("RNC:"), t.rncEntry,
		widget.NewLabel("ID:"), t.idEntry,
	)

	// Botón para refrescar la lista
	refreshBtn := widget.NewButton("Refrescar Lista", func() {
		t.cargarTodosLosClientes()
	})

	// Container para la tabla con altura mínima para mostrar al menos 10 filas
	tableContainer := container.NewScroll(t.table)
	tableContainer.SetMinSize(fyne.NewSize(700, 400)) // Altura suficiente para ~10 filas

	// Container principal
	t.container = container.NewVBox(
		widget.NewCard("Búsqueda de Clientes", "", searchForm),
		refreshBtn,
		widget.NewSeparator(),
		headerContainer,
		tableContainer,
	)
}

// GetContainer retorna el container principal del widget
func (t *TablaClientesWidget) GetContainer() *fyne.Container {
	return t.container
}

// cargarTodosLosClientes carga todos los clientes desde la base de datos
func (t *TablaClientesWidget) cargarTodosLosClientes() {
	postgrestURL, headers := admClient.InitializePostgrest()
	if postgrestURL == "" {
		fmt.Println("Error: URL de PostgREST no configurada")
		return
	}

	// Obtener todos los clientes
	url := postgrestURL + "/cliente?select=*&order=id"

	resp, err := admClient.MakeRequest("GET", url, headers, nil)
	if err != nil {
		fmt.Printf("Error al cargar clientes: %v\n", err)
		return
	}

	var clientes []admClient.Cliente
	if err := json.Unmarshal(resp, &clientes); err != nil {
		fmt.Printf("Error al procesar clientes: %v\n", err)
		return
	}

	t.clientes = clientes
	t.clientesFiltrados = clientes
	t.table.Refresh()
}

// RefrescarClientes actualiza la lista de clientes
func (t *TablaClientesWidget) RefrescarClientes() {
	t.cargarTodosLosClientes()
}

// filtrarClientes filtra la lista según los criterios de búsqueda
func (t *TablaClientesWidget) filtrarClientes() {
	nombreTexto := strings.ToLower(strings.TrimSpace(t.nombreEntry.Text))
	rncTexto := strings.ToLower(strings.TrimSpace(t.rncEntry.Text))
	idTexto := strings.ToLower(strings.TrimSpace(t.idEntry.Text))

	// Si no hay filtros, mostrar todos
	if nombreTexto == "" && rncTexto == "" && idTexto == "" {
		t.clientesFiltrados = t.clientes
		t.table.Refresh()
		return
	}

	var filtrados []admClient.Cliente
	for _, cliente := range t.clientes {
		match := true

		// Filtro por nombre (busca en nombre y nombre comercial)
		if nombreTexto != "" {
			nombreCliente := strings.ToLower(cliente.Nombre)
			nombreComercial := strings.ToLower(cliente.NombreComercial)
			if !strings.Contains(nombreCliente, nombreTexto) && !strings.Contains(nombreComercial, nombreTexto) {
				match = false
			}
		}

		// Filtro por RNC
		if rncTexto != "" && match {
			if !strings.Contains(strings.ToLower(cliente.Numero), rncTexto) {
				match = false
			}
		}

		// Filtro por ID
		if idTexto != "" && match {
			clienteID := strings.ToLower(strconv.Itoa(cliente.ID))
			if !strings.Contains(clienteID, idTexto) {
				match = false
			}
		}

		if match {
			filtrados = append(filtrados, cliente)
		}
	}

	t.clientesFiltrados = filtrados
	t.table.Refresh()
}
