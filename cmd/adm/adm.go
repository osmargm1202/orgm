package adm


import (
	"fmt"
	"github.com/spf13/cobra"
)

// adm/client.goCmd represents the adm/client.go command
var AdmCmd = &cobra.Command{
	Use:   "adm",
	Short: "Management of Clients, Projects, Quotes, Invoices, etc.",
	Long:  `Management of Clients, Projects, Quotes, Invoices, Payments, Receipts, Purchases, Logos, Account Statements`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("client called")
	},
}

func init() {

}
