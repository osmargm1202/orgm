// Package proy provides the proy subcommand.
package proy

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Cmd returns the proy command.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proy",
		Short: "Project management commands",
		Long:  `Manage projects within the orgm ecosystem.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("proy: Project management (not yet implemented)")
		},
	}

	return cmd
}
