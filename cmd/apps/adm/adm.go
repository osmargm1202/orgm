package adm

import (
	"github.com/spf13/cobra"
)

// AdmCmd represents the adm command for the apps module
var AdmCmd = &cobra.Command{
	Use:   "adm",
	Short: "Administration application",
	Long:  `Administration application with modules for clients, projects, quotes, budgets, payments, invoices, receipts, purchases, DGII, TSS`,
	Run: func(cmd *cobra.Command, args []string) {
		// This will launch the Gio GUI application
		runAdmApp()
	},
}

func init() {
	// Add subcommands for different modules if needed
	// For now, we'll handle everything in the main Run function
}

func runAdmApp() {
	// Launch the Gio GUI application
	app := NewAdmApp()
	app.Run()
}
