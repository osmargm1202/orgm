// Package env provides the env subcommand.
package env

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Cmd returns the env command.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Environment management commands",
		Long:  `Manage environment variables and configurations for orgm.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("env: Environment management (not yet implemented)")
		},
	}

	return cmd
}
