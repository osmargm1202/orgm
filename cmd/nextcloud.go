/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/studio-b12/gowebdav"
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

var downloadAppsCmd = &cobra.Command{
	Use:   "download-apps",
	Short: "Download /Apps folder from Nextcloud",
	Long:  `Download the /Apps folder from Nextcloud to local directory based on home + carpetas.apps configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		err := DownloadAppsFolder()
		if err != nil {
			log.Fatal("Error downloading Apps folder:", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(nextcloudCmd)
	nextcloudCmd.AddCommand(listNextcloudCmd)
	nextcloudCmd.AddCommand(createNextcloudCmd)
	nextcloudCmd.AddCommand(downloadAppsCmd)
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

// DownloadAppsFolder downloads the /Apps folder from Nextcloud to local directory
func DownloadAppsFolder() error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %w", err)
	}

	// Get apps path from viper configuration
	appsPath := viper.GetString("carpetas.apps")
	if appsPath == "" {
		return fmt.Errorf("carpeta de apps no configurada (carpetas.apps)")
	}

	// Create local destination path: home + carpetas.apps
	localAppsPath := filepath.Join(homeDir, appsPath)

	// Check if the directory already exists
	if _, err := os.Stat(localAppsPath); err == nil {
		return fmt.Errorf("la carpeta %s ya existe.", localAppsPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("error verificando si existe la carpeta %s: %w", localAppsPath, err)
	}

	// Create local directory since it doesn't exist
	if err := os.MkdirAll(localAppsPath, 0755); err != nil {
		return fmt.Errorf("error creating local directory %s: %w", localAppsPath, err)
	}

	// Initialize Nextcloud client
	client := InitializeNextcloud()
	if err := client.Connect(); err != nil {
		return fmt.Errorf("error connecting to Nextcloud: %w", err)
	}

	// Download the /Apps folder recursively
	fmt.Printf("Downloading /Apps folder to: %s\n", localAppsPath)
	err = downloadFolderRecursive(client, "/Apps", localAppsPath)
	if err != nil {
		return fmt.Errorf("error downloading Apps folder: %w", err)
	}

	fmt.Println("Apps folder downloaded successfully!")
	return nil
}

// downloadFolderRecursive downloads a folder and all its contents recursively
func downloadFolderRecursive(client *gowebdav.Client, remotePath, localPath string) error {
	// Check if remote path exists and is a directory
	stat, err := client.Stat(remotePath)
	if err != nil {
		return fmt.Errorf("error accessing remote path %s: %w", remotePath, err)
	}

	if !stat.IsDir() {
		// If it's a file, download it directly
		return downloadFile(client, remotePath, localPath)
	}

	// Create local directory if it doesn't exist
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("error creating local directory %s: %w", localPath, err)
	}

	// Read directory contents
	files, err := client.ReadDir(remotePath)
	if err != nil {
		return fmt.Errorf("error reading directory %s: %w", remotePath, err)
	}

	// Process each file/folder
	for _, file := range files {
		remoteFinalPath := remotePath + "/" + file.Name()
		localFinalPath := filepath.Join(localPath, file.Name())

		if file.IsDir() {
			// Recursively download subdirectory
			fmt.Printf("Downloading directory: %s\n", remoteFinalPath)
			err := downloadFolderRecursive(client, remoteFinalPath, localFinalPath)
			if err != nil {
				return err
			}
		} else {
			// Download file
			fmt.Printf("Downloading file: %s\n", remoteFinalPath)
			err := downloadFile(client, remoteFinalPath, localFinalPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// downloadFile downloads a single file from Nextcloud
func downloadFile(client *gowebdav.Client, remotePath, localPath string) error {
	// Create local directories if they don't exist
	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", localDir, err)
	}

	// Read file from Nextcloud
	reader, err := client.ReadStream(remotePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", remotePath, err)
	}
	defer reader.Close()

	// Create local file
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("error creating local file %s: %w", localPath, err)
	}
	defer localFile.Close()

	// Copy content from remote to local
	_, err = io.Copy(localFile, reader)
	if err != nil {
		return fmt.Errorf("error copying file content: %w", err)
	}

	return nil
}
