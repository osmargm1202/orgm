package adm

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/spf13/cobra"

)


var RunAdmCmd = &cobra.Command{
	Use:   "adm",
	Short: "Descargar Base de Datos",
	Long:  `Descargar Base de Datos de RNC de la DGII`,
	Run: func(cmd *cobra.Command, args []string) {
		
	},
}


func init() {
	// AppsCmd.AddCommand(RunAdmCmd)
}

// Run starts the administrative application
func Run() {
	myApp := app.New()
	window := myApp.NewWindow("Sistema Administrativo")

	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Cliente", createClientTab()),
		container.NewTabItem("Proyecto", createProjectTab()),
		container.NewTabItem("Cotización", createQuoteTab()),
		container.NewTabItem("Presupuesto", createBudgetTab()),
		container.NewTabItem("Pagos", createPaymentsTab()),
		container.NewTabItem("Estados", createStatusTab()),
		container.NewTabItem("Facturas", createInvoicesTab()),
		container.NewTabItem("Comprobantes", createVouchersTab()),
		container.NewTabItem("Compras", createPurchasesTab()),
	)

	tabs.SetTabLocation(container.TabLocationTop)
	window.SetContent(tabs)
	window.Resize(fyne.NewSize(1200, 800))
	window.ShowAndRun()
}

// Tab creation functions
func createClientTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión de Clientes"),
	)
}

func createProjectTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión de Proyectos"),
	)
}

func createQuoteTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión de Cotizaciones"),
	)
}

func createBudgetTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión de Presupuestos"),
	)
}

func createPaymentsTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión de Pagos"),
	)
}

func createStatusTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión de Estados"),
	)
}

func createInvoicesTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión de Facturas"),
	)
}

func createVouchersTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión de Comprobantes"),
	)
}

func createPurchasesTab() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("Gestión de Compras"),
	)
}


