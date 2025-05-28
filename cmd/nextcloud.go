/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

// folderCmd represents the folder command
var nextcloudCmd = &cobra.Command{
	Use:   "nextcloud",
	Short: "Creation of folders",
	Long:  `Creation of folders for the project and other tools.`,
	// Run: func(cmd *cobra.Command, args []string) {},
}

var createNextcloudCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new folder",
	Long:  `Create a new folder in Nextcloud`,
}

var createProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Create a new project",
	Long:  `Create a new project in Nextcloud`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatal("Project name is required")
			return
		}
		name := strings.Join(args, " ")
		err := CreateNewProject(name)
		if err != nil {
			log.Fatal("Error creating project:", err)
		}
	},
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

func init() {
	RootCmd.AddCommand(nextcloudCmd)
	nextcloudCmd.AddCommand(listNextcloudCmd)
	nextcloudCmd.AddCommand(createNextcloudCmd)
	createNextcloudCmd.AddCommand(createProjectCmd)
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

func CreateNewFolder(path string, name string) error {
	client := InitializeNextcloud()
	if err := client.Connect(); err != nil {
		log.Fatal("Error connecting to Nextcloud:", err)
		return err
	}

	// Combine path and name, and use proper file permissions
	fullPath := path + "/" + name
	err := client.Mkdir(fullPath, 0755)
	if err != nil {
		log.Fatal("Error creating folder:", err)
		return err
	}
	fmt.Println("Folder created successfully")
	return nil
}

func FindProject(ID string) string {
	// de la lista de carpetas de Proyectos extraer el primer nombre dividido por "-" y verificar si es igual al ID
	// devuelva el path commpleto de la carpeta que coincida.

	client := InitializeNextcloud()
	if err := client.Connect(); err != nil {
		log.Fatal("Error connecting to Nextcloud:", err)
		return ""
	}

	files, _ := client.ReadDir("/Proyectos")
	for _, file := range files {
		//notice that [file] has os.FileInfo type
		fmt.Println(file.Name())
		if strings.Contains(file.Name(), "-") {
			split := strings.Split(file.Name(), "-")
			if split[0] == ID {
				return file.Name()
			}
		}
	}
	return ""
}

func MoveFolder(oldPath string, newPath string) error {
	client := InitializeNextcloud()
	if err := client.Connect(); err != nil {
		log.Fatal("Error connecting to Nextcloud:", err)
		return err
	}

	err := client.Rename(oldPath, newPath, true)
	if err != nil {
		log.Fatal("Error moving folder:", err)
		return err
	}
	fmt.Println("Folder moved successfully")
	return nil
}

func CreateNewProject(name string) error {
	err := CreateNewFolder("/Proyectos", name)
	if err != nil {
		return err
	}
	return nil
}
