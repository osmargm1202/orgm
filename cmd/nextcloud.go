/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
    "fmt"
	"log"
	"github.com/spf13/cobra"
)

// folderCmd represents the folder command
var folderCmd = &cobra.Command{
	Use:   "nextcloud",
	Short: "Creation of folders",
	Long: `Creation of folders for the project and other tools.`,
	// Run: func(cmd *cobra.Command, args []string) {},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List folders",
	Long:  `List all folders in Nextcloud`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			listFolders("/")
		} else {
			listFolders("/" + args[0])
		}
	},
}

func listFolders(path string) {
	client := InitializeNextcloud()

	if err := client.Connect(); err != nil {
		log.Fatal("Error connecting to Nextcloud:", err)
		return
	}

	files, _ := client.ReadDir(path)
	for _, file := range files {
		//notice that [file] has os.FileInfo type
		fmt.Println(file.Name())
	}
}

func init() {
	rootCmd.AddCommand(folderCmd)
	folderCmd.AddCommand(listCmd)
}