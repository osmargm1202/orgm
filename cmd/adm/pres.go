package adm

import (
	"fmt"

	"github.com/spf13/cobra"
)

// presupuestoCmd represents the pres command
var presupuestoCmd = &cobra.Command{
	Use:   "pres",
	Short: "Gestionar Presupuestos",
	Long:  `Crear, editar y gestionar presupuestos de construcción`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Presupuesto CLI - Funcionalidad en desarrollo")
		fmt.Println("Esta herramienta permitirá gestionar presupuestos desde la línea de comandos.")
		fmt.Println("Vuelva a intentarlo más adelante.")
	},
}

func init() {
	AdmCmd.AddCommand(presupuestoCmd)
}
