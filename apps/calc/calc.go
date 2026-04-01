// Package calc provides the calc subcommand.
package calc

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Cmd returns the calc command.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calc",
		Short: "Calculator utilities",
		Long:  `Calculation tools and utilities for orgm.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("calc: Calculator utilities (not yet implemented)")
		},
	}

	return cmd
}
