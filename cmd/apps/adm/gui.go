package adm

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// AdmApp represents the main administration application
type AdmApp struct {
	app    fyne.App
	window fyne.Window
}

// NewAdmApp creates a new administration application
func NewAdmApp() *AdmApp {
	return &AdmApp{
		app: app.New(),
	}
}

// Run starts the GUI application
func (a *AdmApp) Run() {
	a.window = a.app.NewWindow("ORGM - Administración")
	a.window.Resize(fyne.NewSize(1200, 800))

	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Clientes", a.createClientesTab()),
		container.NewTabItem("Proyectos", a.createProyectosTab()),
		container.NewTabItem("Cotizaciones", a.createCotizacionesTab()),
		container.NewTabItem("Presupuestos", a.createPresupuestosTab()),
		container.NewTabItem("Pagos", a.createPagosTab()),
		container.NewTabItem("Facturas", a.createFacturasTab()),
		container.NewTabItem("Comprobantes", a.createComprobantesTab()),
		container.NewTabItem("Compras", a.createComprasTab()),
		container.NewTabItem("DGII", a.createDGIITab()),
		container.NewTabItem("TSS", a.createTSSTab()),
	)

	a.window.SetContent(tabs)
	a.window.ShowAndRun()
}

// Tab creation functions
func (a *AdmApp) createClientesTab() fyne.CanvasObject {
	// Crear el manager de clientes
	clientesManager := NewClientesManager(a.window)

	// Crear botón para agregar cliente
	agregarClienteBtn := widget.NewButton("Agregar Cliente", clientesManager.GetAgregarClienteHandler())
	agregarClienteBtn.Importance = widget.SuccessImportance

	// Container principal con el botón arriba y el manager abajo
	return container.NewBorder(
		container.NewHBox(agregarClienteBtn), // Top
		nil,                                  // Bottom
		nil,                                  // Left
		nil,                                  // Right
		clientesManager.GetContainer(),       // Center
	)
}

func (a *AdmApp) createProyectosTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión de Proyectos"),
		widget.NewLabel("Aquí irá la funcionalidad de proyectos"),
	)
}

func (a *AdmApp) createCotizacionesTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión de Cotizaciones"),
		widget.NewLabel("Aquí irá la funcionalidad de cotizaciones"),
	)
}

func (a *AdmApp) createPresupuestosTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión de Presupuestos"),
		widget.NewLabel("Aquí irá la funcionalidad de presupuestos"),
	)
}

func (a *AdmApp) createPagosTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión de Pagos"),
		widget.NewLabel("Aquí irá la funcionalidad de pagos"),
	)
}

func (a *AdmApp) createFacturasTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión de Facturas"),
		widget.NewLabel("Aquí irá la funcionalidad de facturas"),
	)
}

func (a *AdmApp) createComprobantesTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión de Comprobantes"),
		widget.NewLabel("Aquí irá la funcionalidad de comprobantes"),
	)
}

func (a *AdmApp) createComprasTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión de Compras"),
		widget.NewLabel("Aquí irá la funcionalidad de compras"),
	)
}

func (a *AdmApp) createDGIITab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión DGII"),
		widget.NewLabel("Aquí irá la funcionalidad de DGII"),
	)
}

func (a *AdmApp) createTSSTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión TSS"),
		widget.NewLabel("Aquí irá la funcionalidad de TSS"),
	)
}
