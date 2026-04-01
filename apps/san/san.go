// Package san provides the san subcommand.
package san

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Cmd returns the san command.
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "san",
		Short: "Sanity check commands",
		Long:  `Run sanity checks and validations.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("san: Sanity checks (not yet implemented)")
		},
	}

	return cmd
}
