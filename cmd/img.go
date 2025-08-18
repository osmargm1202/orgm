package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
)

var imgCmd = &cobra.Command{
	Use:   "img",
	Short: "Gestionar archivos en la API",
	Long:  `Comando para subir, modificar, obtener y buscar archivos (imágenes, PDFs, documentos, configs) en la API de orgmapp.com`,
	Run:   runImg,
}

func init() {
	RootCmd.AddCommand(imgCmd)
}

func runImg(cmd *cobra.Command, args []string) {
	fmt.Println("=== Gestión de Archivos ===")

	// Preguntar la acción
	action := selectAction()

	switch action {
	case "subir":
		uploadImage()
	case "modificar":
		updateImage()
	case "obtener":
		downloadImage()
	case "buscar":
		searchImages()
	default:
		fmt.Println("Acción no válida")
	}
}

func selectAction() string {
	actions := []string{
		"Subir imagen",
		"Modificar imagen",
		"Obtener imagen",
		"Buscar imágenes",
	}

	model := inputs.SelectionModel(actions, "Gestión de Archivos", "Selecciona una acción:")
	p := tea.NewProgram(model)
	result, _ := p.Run()
	finalModel := result.(inputs.SelectionModelS)

	if finalModel.Selected && finalModel.Cursor < len(actions) {
		switch finalModel.Cursor {
		case 0:
			return "subir"
		case 1:
			return "modificar"
		case 2:
			return "obtener"
		case 3:
			return "buscar"
		}
	}

	return "subir"
}

func listFolders(file_type string) []string {
	resp, err := http.Get(fmt.Sprintf("https://img.orgmapp.com/folders/%s", file_type))
	if err != nil {
		fmt.Printf("❌ Error al obtener carpetas: %v\n", err)
		return []string{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Error HTTP %d: %s\n", resp.StatusCode, string(body))
		return []string{}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Error al leer respuesta: %v\n", err)
		return []string{}
	}

	var result struct {
		Folders []string `json:"folders"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Printf("❌ Error al parsear respuesta: %v\n", err)
		return []string{}
	}

	return result.Folders

}

func uploadImage() {
	fmt.Println("\n=== Subir Imagen ===")

	// Seleccionar tipo de archivo
	tipos := []string{"images", "pdfs", "docx", "configs"}
	model := inputs.SelectionModel(tipos, "Tipo de archivo", "Selecciona el tipo de archivo:")
	p := tea.NewProgram(model)
	result, _ := p.Run()
	finalModel := result.(inputs.SelectionModelS)

	if !finalModel.Selected || finalModel.Cursor >= len(tipos) {
		fmt.Println("❌ Error: Debes seleccionar un tipo de archivo")
		return
	}
	fileType := tipos[finalModel.Cursor]

	// Obtener lista de carpetas del tipo seleccionado
	folders := listFolders(fileType)
	if len(folders) == 0 {
		fmt.Println("❌ Error: No se pudieron obtener las carpetas")
		return
	}

	// Agregar opción para crear nueva carpeta
	folders = append(folders, "📁 Crear nueva carpeta")

	// Seleccionar carpeta
	model = inputs.SelectionModel(folders, "Carpetas disponibles", "Selecciona una carpeta:")
	p = tea.NewProgram(model)
	result, _ = p.Run()
	finalModel = result.(inputs.SelectionModelS)

	if !finalModel.Selected || finalModel.Cursor >= len(folders) {
		fmt.Println("❌ Error: Debes seleccionar una carpeta")
		return
	}

	var selectedFolder string
	if finalModel.Cursor == len(folders)-1 {
		// Usuario seleccionó crear nueva carpeta
		selectedFolder = inputs.GetInput("Ingresa el nombre de la nueva carpeta:")
		if selectedFolder == "" {
			fmt.Println("❌ Error: El nombre de la carpeta es requerido")
			return
		}
	} else {
		selectedFolder = folders[finalModel.Cursor]
	}

	// Obtener nombre del archivo
	fileName := inputs.GetInput("Ingresa el nombre del archivo:")
	if fileName == "" {
		fmt.Println("❌ Error: El nombre del archivo es requerido")
		return
	}

	// Obtener descripción
	description := inputs.GetInput("Ingresa la descripción:")
	if description == "" {
		fmt.Println("❌ Error: La descripción es requerida")
		return
	}

	// Obtener ruta del archivo
	filePath := inputs.GetInput("Ingresa la ruta del archivo:")
	if filePath == "" {
		fmt.Println("❌ Error: La ruta del archivo es requerida")
		return
	}

	// Limpiar comillas si las tiene
	filePath = strings.Trim(filePath, `"'`)

	// Verificar que el archivo existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Println("❌ Error: El archivo no existe")
		return
	}

	// Verificar extensión según el tipo de archivo
	ext := strings.ToLower(filepath.Ext(filePath))
	switch fileType {
	case "images":
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".gif" {
			fmt.Println("❌ Error: El archivo debe ser una imagen (PNG, JPG, JPEG, GIF)")
			return
		}
	case "pdfs":
		if ext != ".pdf" {
			fmt.Println("❌ Error: El archivo debe ser PDF")
			return
		}
	case "docx":
		if ext != ".docx" && ext != ".doc" {
			fmt.Println("❌ Error: El archivo debe ser DOCX o DOC")
			return
		}
	case "configs":
		// Para configs aceptamos cualquier archivo
	}

	// Subir archivo
	err := uploadFileToAPI(fileType, selectedFolder, fileName, description, filePath)
	if err != nil {
		fmt.Printf("❌ Error al subir archivo: %v\n", err)
		return
	}

	fmt.Println("✅ Archivo subido exitosamente")
}

func updateImage() {
	fmt.Println("\n=== Modificar Archivo ===")

	// Seleccionar tipo de archivo
	tipos := []string{"images", "pdfs", "docx", "configs"}
	model := inputs.SelectionModel(tipos, "Tipo de archivo", "Selecciona el tipo de archivo:")
	p := tea.NewProgram(model)
	result, _ := p.Run()
	finalModel := result.(inputs.SelectionModelS)

	if !finalModel.Selected || finalModel.Cursor >= len(tipos) {
		fmt.Println("❌ Error: Debes seleccionar un tipo de archivo")
		return
	}
	fileType := tipos[finalModel.Cursor]

	// Obtener lista de carpetas del tipo seleccionado
	folders := listFolders(fileType)
	if len(folders) == 0 {
		fmt.Println("❌ Error: No se pudieron obtener las carpetas")
		return
	}

	// Seleccionar carpeta
	model = inputs.SelectionModel(folders, "Carpetas disponibles", "Selecciona una carpeta:")
	p = tea.NewProgram(model)
	result, _ = p.Run()
	finalModel = result.(inputs.SelectionModelS)

	if !finalModel.Selected || finalModel.Cursor >= len(folders) {
		fmt.Println("❌ Error: Debes seleccionar una carpeta")
		return
	}
	selectedFolder := folders[finalModel.Cursor]

	// Obtener nombre del archivo
	fileName := inputs.GetInput("Ingresa el nombre del archivo:")
	if fileName == "" {
		fmt.Println("❌ Error: El nombre del archivo es requerido")
		return
	}

	// Preguntar qué modificar
	fmt.Println("¿Qué deseas modificar?")
	fmt.Println("1. Solo descripción")
	fmt.Println("2. Solo archivo")
	fmt.Println("3. Ambos")

	choice := inputs.GetInput("Selecciona una opción (1-3):")

	var description string
	var filePath string

	switch choice {
	case "1":
		description = inputs.GetInput("Ingresa la nueva descripción:")
	case "2":
		filePath = inputs.GetInput("Ingresa la ruta del nuevo archivo:")
		filePath = strings.Trim(filePath, `"'`)
	case "3":
		description = inputs.GetInput("Ingresa la nueva descripción:")
		filePath = inputs.GetInput("Ingresa la ruta del nuevo archivo:")
		filePath = strings.Trim(filePath, `"'`)
	}

	// Verificar archivo si se proporciona
	if filePath != "" {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Println("❌ Error: El archivo no existe")
			return
		}

		// Verificar extensión según el tipo de archivo
		ext := strings.ToLower(filepath.Ext(filePath))
		switch fileType {
		case "images":
			if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".gif" {
				fmt.Println("❌ Error: El archivo debe ser una imagen (PNG, JPG, JPEG, GIF)")
				return
			}
		case "pdfs":
			if ext != ".pdf" {
				fmt.Println("❌ Error: El archivo debe ser PDF")
				return
			}
		case "docx":
			if ext != ".docx" && ext != ".doc" {
				fmt.Println("❌ Error: El archivo debe ser DOCX o DOC")
				return
			}
		case "configs":
			// Para configs aceptamos cualquier archivo
		}
	}

	// Modificar archivo
	err := updateFileInAPI(fileType, selectedFolder, fileName, description, filePath)
	if err != nil {
		fmt.Printf("❌ Error al modificar archivo: %v\n", err)
		return
	}

	fmt.Println("✅ Archivo modificado exitosamente")
}

func downloadImage() {
	fmt.Println("\n=== Obtener Archivo ===")

	// Seleccionar tipo de archivo
	tipos := []string{"images", "pdfs", "docx", "configs"}
	model := inputs.SelectionModel(tipos, "Tipo de archivo", "Selecciona el tipo de archivo:")
	p := tea.NewProgram(model)
	result, _ := p.Run()
	finalModel := result.(inputs.SelectionModelS)

	if !finalModel.Selected || finalModel.Cursor >= len(tipos) {
		fmt.Println("❌ Error: Debes seleccionar un tipo de archivo")
		return
	}
	fileType := tipos[finalModel.Cursor]

	// Obtener lista de carpetas del tipo seleccionado
	folders := listFolders(fileType)
	if len(folders) == 0 {
		fmt.Println("❌ Error: No se pudieron obtener las carpetas")
		return
	}

	// Seleccionar carpeta
	model = inputs.SelectionModel(folders, "Carpetas disponibles", "Selecciona una carpeta:")
	p = tea.NewProgram(model)
	result, _ = p.Run()
	finalModel = result.(inputs.SelectionModelS)

	if !finalModel.Selected || finalModel.Cursor >= len(folders) {
		fmt.Println("❌ Error: Debes seleccionar una carpeta")
		return
	}
	selectedFolder := folders[finalModel.Cursor]

	// Obtener nombre del archivo
	fileName := inputs.GetInput("Ingresa el nombre del archivo:")
	if fileName == "" {
		fmt.Println("❌ Error: El nombre del archivo es requerido")
		return
	}

	// Preguntar dimensiones (solo para imágenes)
	var width, height string
	if fileType == "images" {
		width = inputs.GetInput("¿Ancho (opcional, presiona Enter para omitir):")
		height = inputs.GetInput("¿Alto (opcional, presiona Enter para omitir):")
	}

	// Descargar archivo
	err := downloadFileFromAPI(fileType, selectedFolder, fileName, width, height)
	if err != nil {
		fmt.Printf("❌ Error al descargar archivo: %v\n", err)
		return
	}

	fmt.Println("✅ Archivo descargado exitosamente en ~/Downloads/")
}

func searchImages() {
	fmt.Println("\n=== Buscar Archivos ===")

	// Seleccionar tipo de archivo
	tipos := []string{"images", "pdfs", "docx", "configs"}
	model := inputs.SelectionModel(tipos, "Tipo de archivo", "Selecciona el tipo de archivo:")
	p := tea.NewProgram(model)
	result, _ := p.Run()
	finalModel := result.(inputs.SelectionModelS)

	if !finalModel.Selected || finalModel.Cursor >= len(tipos) {
		fmt.Println("❌ Error: Debes seleccionar un tipo de archivo")
		return
	}
	fileType := tipos[finalModel.Cursor]

	// Obtener query
	query := inputs.GetInput("Ingresa el término de búsqueda:")
	if query == "" {
		fmt.Println("❌ Error: El término de búsqueda es requerido")
		return
	}

	// Buscar archivos
	results, err := searchFilesInAPI(fileType, query)
	if err != nil {
		fmt.Printf("❌ Error al buscar archivos: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Println("No se encontraron archivos")
		return
	}

	fmt.Printf("\nSe encontraron %d archivos:\n", len(results))
	for i, result := range results {
		fmt.Printf("%d. Nombre: %s, Descripción: %s, Ruta: %s\n",
			i+1, result.Name, result.Description, result.Filepath)
	}
}

func uploadFileToAPI(fileType, folder, fileName, description, filePath string) error {
	// Crear el archivo multipart
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Crear buffer para el multipart
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Agregar campos de texto (file_id y description) al multipart
	err = writer.WriteField("file_id", fileName)
	if err != nil {
		return err
	}

	err = writer.WriteField("description", description)
	if err != nil {
		return err
	}

	// Crear el campo de archivo
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	writer.Close()

	// Crear request con el tipo de archivo y folder en la URL
	url := fmt.Sprintf("https://img.orgmapp.com/%s/%s", fileType, folder)
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return err
	}

	// Establecer el Content-Type del multipart
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Agregar headers adicionales
	req.Header.Set("Accept", "application/json")

	// Enviar request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func updateFileInAPI(fileType, folder, fileName, description, filePath string) error {
	// Crear el archivo multipart
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Agregar campos si se proporcionan
	if description != "" {
		writer.WriteField("description", description)
	}

	// Agregar archivo si se proporciona
	if filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		part, err := writer.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			return err
		}

		_, err = io.Copy(part, file)
		if err != nil {
			return err
		}
	}

	writer.Close()

	// Crear request con file_id en la URL
	url := fmt.Sprintf("https://img.orgmapp.com/%s/%s/%s", fileType, folder, fileName)
	req, err := http.NewRequest("PATCH", url, &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Enviar request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func downloadFileFromAPI(fileType, folder, fileName, width, height string) error {
	// Construir URL con file_id en la URL
	url := fmt.Sprintf("https://img.orgmapp.com/%s/%s/%s", fileType, folder, fileName)

	// Agregar parámetros de dimensiones si se proporcionan
	if width != "" && height != "" {
		url += fmt.Sprintf("?width=%s&height=%s", width, height)
	}

	// Crear request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Enviar request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Obtener directorio de descargas
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	downloadsDir := filepath.Join(homeDir, "Downloads")

	// Crear directorio si no existe
	if err := os.MkdirAll(downloadsDir, 0755); err != nil {
		return err
	}

	// Crear nombre de archivo
	filename := fileName
	if fileType == "images" && width != "" && height != "" {
		// Para imágenes con dimensiones, agregar sufijo
		baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		ext := filepath.Ext(fileName)
		if ext == "" {
			ext = ".png" // extensión por defecto para imágenes
		}
		filename = fmt.Sprintf("%s_%sx%s%s", baseName, width, height, ext)
	}

	filePath := filepath.Join(downloadsDir, filename)

	// Crear archivo
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copiar contenido
	_, err = io.Copy(file, resp.Body)
	return err
}

type SearchResult struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Filepath    string `json:"filepath"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

func searchFilesInAPI(fileType, query string) ([]SearchResult, error) {
	// Construir URL
	url := fmt.Sprintf("https://img.orgmapp.com/search/%s?query=%s", fileType, query)

	// Crear request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Enviar request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Leer respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parsear JSON
	var response SearchResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response.Results, nil
}
