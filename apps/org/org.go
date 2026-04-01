// Package org provides the org subcommand.
package org

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Cmd returns the org command.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org",
		Short: "Organization management commands",
		Long:  `Manage organizational structure and settings.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("org: Organization management (not yet implemented)")
		},
	}

	return cmd
}
