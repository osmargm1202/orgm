package misc


import (
	"fmt"
	"github.com/spf13/cobra"
)

// adm/client.goCmd represents the adm/client.go command
var MiscCmd = &cobra.Command{
	Use:   "misc",
	Short: "Miscellaneous commands",
	Long:  `Miscellaneous commands`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("client called")
	},
}

func init() {

}
