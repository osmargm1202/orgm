package adm

import "github.com/spf13/cobra"


var NewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new",
	Long:  `Create a new adm`,
}


func init() {
	AdmCmd.AddCommand(NewCmd)
}