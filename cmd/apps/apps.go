package apps

import (
	"github.com/osmargm1202/orgm/cmd/apps/adm"
	"github.com/spf13/cobra"
)

// AppsCmd represents the apps command
var AppsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Execute different applications",
	Long:  `Execute different applications like administration, accounting, etc.`,
}

func init() {
	// Add subcommands for different applications
	AppsCmd.AddCommand(adm.AdmCmd)
}
