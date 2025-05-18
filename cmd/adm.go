/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// admCmd represents the adm command
var admCmd = &cobra.Command{
	Use:   "adm",
	Short: "Management of Clients, Projects, Quotes, Invoices, etc.",
	Long:  `Management of Clients, Projects, Quotes, Invoices, Payments, Receipts, Purchases, Logos, Account Statements`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("adm called")
	},
}

func init() {
	rootCmd.AddCommand(admCmd)

}
