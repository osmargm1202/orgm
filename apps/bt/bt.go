// Package bt provides the bt subcommand.
package bt

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Cmd returns the bt command.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bt",
		Short: "Bluetooth management commands",
		Long:  `Manage Bluetooth devices and connections.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("bt: Bluetooth management (not yet implemented)")
		},
	}

	return cmd
}
