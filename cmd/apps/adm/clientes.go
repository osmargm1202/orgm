package adm

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	admClient "github.com/osmargm1202/orgm/cmd/adm"
)

// ClientesManager maneja la navegación entre vistas de clientes
type ClientesManager struct {
	window        fyne.Window
	container     *fyne.Container
	tablaWidget   *TablaClientesWidget
	editarWidget  *EditarClienteWidget
	vistaActual   string // "tabla" o "editar"
	clienteActual *admClient.Cliente
}

// NewClientesManager crea un nuevo manager de clientes
func NewClientesManager(window fyne.Window) *ClientesManager {
	cm := &ClientesManager{
		window:      window,
		vistaActual: "tabla",
	}

	cm.inicializar()
	return cm
}

func (cm *ClientesManager) inicializar() {
	// Crear la tabla de clientes
	cm.tablaWidget = NewTablaClientesWidget(func(cliente admClient.Cliente) {
		cm.editarCliente(&cliente)
	})

	// Container principal
	cm.container = container.NewBorder(nil, nil, nil, nil, cm.tablaWidget.GetContainer())
}

// GetContainer retorna el container principal del manager
func (cm *ClientesManager) GetContainer() *fyne.Container {
	return cm.container
}

// agregarCliente muestra la vista de creación de cliente
func (cm *ClientesManager) agregarCliente() {
	cm.editarWidget = NewEditarClienteWidget(
		cm.window,
		nil, // nil significa crear nuevo cliente
		func(cliente admClient.Cliente) {
			// Cliente guardado exitosamente
			cm.tablaWidget.RefrescarClientes()
			cm.volverATabla()
		},
		func() {
			// Cancelado
			cm.volverATabla()
		},
	)

	cm.vistaActual = "editar"
	cm.container.Objects = []fyne.CanvasObject{cm.editarWidget.GetContainer()}
	cm.container.Refresh()
}

// editarCliente muestra la vista de edición de cliente
func (cm *ClientesManager) editarCliente(cliente *admClient.Cliente) {
	cm.clienteActual = cliente
	cm.editarWidget = NewEditarClienteWidget(
		cm.window,
		cliente,
		func(clienteGuardado admClient.Cliente) {
			// Cliente guardado exitosamente
			cm.tablaWidget.RefrescarClientes()
			cm.volverATabla()
		},
		func() {
			// Cancelado
			cm.volverATabla()
		},
	)

	cm.vistaActual = "editar"
	cm.container.Objects = []fyne.CanvasObject{cm.editarWidget.GetContainer()}
	cm.container.Refresh()
}

// volverATabla vuelve a mostrar la tabla de clientes
func (cm *ClientesManager) volverATabla() {
	cm.vistaActual = "tabla"
	cm.clienteActual = nil
	cm.editarWidget = nil
	cm.container.Objects = []fyne.CanvasObject{cm.tablaWidget.GetContainer()}
	cm.container.Refresh()
}

// refrescarDatos actualiza los datos de la tabla
func (cm *ClientesManager) refrescarDatos() {
	if cm.tablaWidget != nil {
		cm.tablaWidget.RefrescarClientes()
	}
}

// GetAgregarClienteHandler retorna la función para agregar cliente
func (cm *ClientesManager) GetAgregarClienteHandler() func() {
	return cm.agregarCliente
}
