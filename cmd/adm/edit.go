package adm

import (
	"github.com/spf13/cobra"
)

// EditCmd represents the edit command
var EditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit existing records (quotations, invoices, etc.)",
	Long:  `Edit existing records such as quotations, invoices, clients, projects, and more.`,
}

func init() {
	AdmCmd.AddCommand(EditCmd)
}