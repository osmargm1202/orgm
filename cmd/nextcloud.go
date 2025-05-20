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
var nextcloudCmd = &cobra.Command{
	Use:   "nextcloud",
	Short: "Creation of folders",
	Long:  `Creation of folders for the project and other tools.`,
	// Run: func(cmd *cobra.Command, args []string) {},
}

var listNextcloudCmd = &cobra.Command{
	Use:   "list",
	Short: "List folders",
	Long:  `List all folders in Nextcloud`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			listNextcloud("/")
		} else {
			listNextcloud("/" + args[0])
		}
	},
}

func listNextcloud(path string) {
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
	RootCmd.AddCommand(nextcloudCmd)
	nextcloudCmd.AddCommand(listNextcloudCmd)
}
