
package adm

import (
	"fmt"
	"github.com/spf13/cobra"
)

// adm/client.goCmd represents the adm/client.go command
var ClientCmd = &cobra.Command{
	Use:   "client",
	Short: "Management of Clients, Projects, Quotes, Invoices, etc.",
	Long:  `Management of Clients, Projects, Quotes, Invoices, Payments, Receipts, Purchases, Logos, Account Statements`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("client called")
	},
}

func init() {
	AdmCmd.AddCommand(ClientCmd)

}
