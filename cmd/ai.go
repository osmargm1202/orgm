/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// AI Request/Response structures for new API
type aiRequest struct {
	Text            string                   `json:"text"`
	Model           string                   `json:"model"`
	MessagesHistory []map[string]interface{} `json:"messages_history,omitempty"`
}

type aiResponse struct {
	Response   string `json:"response"`
	ModelUsed  string `json:"model_used"`
	HasHistory bool   `json:"has_history"`
}

type aiModelsResponse struct {
	Models []string `json:"models"`
}

// Conversation structure for saving/loading
type Conversation struct {
	Name       string                   `json:"name"`
	CreatedAt  string                   `json:"created_at"`
	AIType     string                   `json:"ai_type"`     // "ai" or "perai"
	Model      string                   `json:"model"`       // Current model being used
	ConfigUsed string                   `json:"config_used"` // Configuration file used
	Messages   []map[string]interface{} `json:"messages"`
}

// aiCmd represents the ai command
var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "OpenAI GPT interaction",
	Long:  `Interact with OpenAI GPT models with conversation management and history`,
	Run: func(cmd *cobra.Command, args []string) {
		startAIConversation(args)
	},
}

var aiTxtCmd = &cobra.Command{
	Use:   "txt",
	Short: "Export AI conversation to TXT",
	Long:  `Export an AI conversation to TXT format`,
	Run: func(cmd *cobra.Command, args []string) {
		exportAIConversationToTXT()
	},
}

var aiDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete AI conversations",
	Long:  `Delete AI conversations with options to delete all or select specific ones`,
	Run: func(cmd *cobra.Command, args []string) {
		deleteAIConversations()
	},
}

func startAIConversation(args []string) {
	// Step 1: Ask if user wants new conversation or continue existing
	conversationChoice := selectConversationType("AI", "¿Qué quieres hacer?")
	if conversationChoice == "" {
		return
	}

	var conversation *Conversation
	var configMessages []map[string]interface{}

	if conversationChoice == "Nueva conversación" {
		// Step 2: Select configuration for new conversation
		configName := selectConfigFile()
		if configName == "" {
			return
		}

		// Load configuration messages
		configMessages = loadConfigMessages(configName)
		if configMessages == nil {
			return
		}

		// Step 3: Ask for conversation name
		conversationName := getConversationName()
		if conversationName == "" {
			return
		}

		// Create new conversation
		conversation = &Conversation{
			Name:       conversationName,
			CreatedAt:  time.Now().Format("2006-01-02_15-04-05"),
			AIType:     "ai",
			Model:      "", // Will be set when model is selected
			ConfigUsed: configName,
			Messages:   configMessages,
		}
	} else {
		// Continue existing conversation
		conversation = selectExistingConversation("ai")
		if conversation == nil {
			return
		}

		// Limpiar el historial para asegurar alternancia correcta
		conversation.Messages = cleanMessageHistory(conversation.Messages)
	}

	// Step 4: Select model (only if not set or if user wants to change)
	var model string
	if conversation.Model == "" {
		model = selectAIModel()
		if model == "" {
			return
		}
		conversation.Model = model
	} else {
		// Ask if user wants to keep the same model or change it
		changeModel := askToChangeModel(conversation.Model)
		if changeModel {
			model = selectAIModel()
			if model == "" {
				return
			}
			conversation.Model = model
		} else {
			model = conversation.Model
		}
	}

	// Step 5: Start conversation loop
	continueConversationLoop(args, conversation, model, "ai")
}

func continueConversationLoop(args []string, conversation *Conversation, model string, aiType string) {
	// Mostrar historial si existe y no es una nueva conversación
	if len(conversation.Messages) > 1 { // Más de solo el mensaje del sistema
		displayConversationHistory(conversation)
	}

	// Obtener el primer prompt si se proporcionó en args
	var promptText string
	if len(args) > 0 {
		promptText = strings.Join(args, " ")
	}

	for {
		// Si no hay prompt inicial o ya se usó, pedir uno nuevo
		if promptText == "" {
			fmt.Printf("\n%s\n", inputs.InfoStyle.Render("Escribe tu pregunta (Ctrl+C para terminar):"))
			p := tea.NewProgram(inputs.TextInput("", "¿En qué puedo ayudarte?"))
			m, err := p.Run()
			if err != nil {
				fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error getting input: "+err.Error()))
				return
			}

			if model, ok := m.(inputs.TextInputModel); ok {
				promptText = model.TextInput.Value()
				if promptText == "" {
					fmt.Println(inputs.InfoStyle.Render("Conversación terminada."))
					return
				}
			} else {
				return
			}
		}

		// Hacer consulta a la API
		fmt.Printf("\n%s %s\n", inputs.CursorStyle.Render("👤"), inputs.ItemStyle.Render(promptText))

		response := makeAIRequest(promptText, model, conversation.Messages)
		if response == "" {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error: No se pudo obtener respuesta de la API"))
			promptText = "" // Reset para pedir nuevo input
			continue
		}

		// Agregar mensajes a la conversación
		userMessage := map[string]interface{}{
			"role":    "user",
			"content": promptText,
		}
		conversation.Messages = append(conversation.Messages, userMessage)

		assistantMessage := map[string]interface{}{
			"role":    "assistant",
			"content": response,
		}
		conversation.Messages = append(conversation.Messages, assistantMessage)

		// Guardar conversación
		saveConversation(conversation, aiType)

		// Mostrar respuesta
		fmt.Printf("%s\n", inputs.CheckedStyle.Render("🤖 Respuesta:"))
		markdownResponse := renderMarkdown(response)
		fmt.Print(markdownResponse)
		fmt.Println()

		// Reset prompt para el siguiente loop
		promptText = ""
	}
}

func selectConversationType(aiType, title string) string {
	choices := []string{"Continuar conversación", "Nueva conversación"}
	p := tea.NewProgram(inputs.SelectionModel(choices, title, "Selecciona una opción"), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running menu: "+err.Error()))
		return ""
	}

	if model, ok := m.(inputs.SelectionModelS); ok {
		if model.Quitting || !model.Selected {
			return ""
		}
		return model.Choices[model.Cursor].Title
	}
	return ""
}

func selectConfigFile() string {
	appsPath := viper.GetString("carpetas.apps")
	if appsPath == "" {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error: carpetas.apps not configured"))
		return ""
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error getting home directory: "+err.Error()))
		return ""
	}

	aiConfigPath := filepath.Join(homeDir, appsPath, "ai", "configs", "ai")
	files, err := os.ReadDir(aiConfigPath)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error reading config files: "+err.Error()))
		return ""
	}

	var configFiles []string
	// Agregar opción "default" al inicio
	configFiles = append(configFiles, "default")

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			configName := strings.TrimSuffix(file.Name(), ".json")
			configFiles = append(configFiles, configName)
		}
	}

	if len(configFiles) == 1 { // Solo la opción "default"
		fmt.Println(inputs.InfoStyle.Render("No configuration files found, using default"))
		return "default" // Devolver "default" por defecto
	}

	p := tea.NewProgram(inputs.SelectionModel(configFiles, "SELECCIONA CONFIGURACIÓN", "Selecciona un archivo de configuración o default"), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running menu: "+err.Error()))
		return ""
	}

	if model, ok := m.(inputs.SelectionModelS); ok {
		if model.Quitting || !model.Selected {
			return ""
		}
		return model.Choices[model.Cursor].Title
	}
	return ""
}

func getDefaultConfig() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"role":    "system",
			"content": "Eres un asistente útil y amigable. Responde de manera clara y concisa.",
		},
	}
}

func loadConfigMessages(configName string) []map[string]interface{} {
	// Si se seleccionó "default", devolver configuración por defecto
	if configName == "default" {
		return getDefaultConfig()
	}

	appsPath := viper.GetString("carpetas.apps")
	if appsPath == "" {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error: carpetas.apps not configured"))
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error getting home directory: "+err.Error()))
		return nil
	}

	configFile := filepath.Join(homeDir, appsPath, "ai", "configs", "ai", configName+".json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error reading config file: "+err.Error()))
		return nil
	}

	var config struct {
		Messages []map[string]interface{} `json:"messages"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error parsing config file: "+err.Error()))
		return nil
	}

	return config.Messages
}

func getConversationName() string {
	p := tea.NewProgram(inputs.TextInput("Nombre de la conversación: ", "Mi conversación"), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running input: "+err.Error()))
		return ""
	}

	if model, ok := m.(inputs.TextInputModel); ok {
		if model.TextInput.Value() != "" {
			return model.TextInput.Value()
		}
	}
	return ""
}

func selectExistingConversation(aiType string) *Conversation {
	conversations := listConversations(aiType)
	if len(conversations) == 0 {
		fmt.Println(inputs.ErrorStyle.Render("No hay conversaciones guardadas"))
		return nil
	}

	// Sort by creation time (most recent first)
	sort.Slice(conversations, func(i, j int) bool {
		return conversations[i].CreatedAt > conversations[j].CreatedAt
	})

	var choices []string
	for _, conv := range conversations {
		displayName := fmt.Sprintf("%s (%s)", conv.Name, conv.CreatedAt)
		choices = append(choices, displayName)
	}

	p := tea.NewProgram(inputs.SelectionModel(choices, "CONTINUAR CONVERSACIÓN", "Selecciona una conversación"), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running menu: "+err.Error()))
		return nil
	}

	if model, ok := m.(inputs.SelectionModelS); ok {
		if model.Quitting || !model.Selected {
			return nil
		}
		selectedConv := &conversations[model.Cursor]

		return selectedConv
	}
	return nil
}

func listConversations(aiType string) []Conversation {
	configPath := viper.GetString("config_path")
	if configPath == "" {
		configPath = filepath.Join(os.Getenv("HOME"), ".config", "orgm")
	}

	conversationsPath := filepath.Join(configPath, "ai", "conversacion", aiType)
	files, err := os.ReadDir(conversationsPath)
	if err != nil {
		return []Conversation{}
	}

	var conversations []Conversation
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			filePath := filepath.Join(conversationsPath, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}

			var conversation Conversation
			if err := json.Unmarshal(data, &conversation); err != nil {
				continue
			}
			conversations = append(conversations, conversation)
		}
	}
	return conversations
}

func selectAIModel() string {
	models := getAIModels()

	if len(models) == 0 {
		fmt.Println(inputs.ErrorStyle.Render("No se pudieron obtener los modelos disponibles"))
		return "gpt-4o-mini" // default fallback
	}

	// Usar el nuevo modelo de selección con filtrado
	filterableModel := NewFilterableSelectionModel(models, "SELECCIONA MODELO", "Selecciona un modelo de OpenAI (presiona '/' para filtrar)")
	p := tea.NewProgram(filterableModel, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running menu: "+err.Error()))
		return "gpt-4o-mini" // default fallback
	}

	if model, ok := m.(FilterableSelectionModel); ok {
		if model.quitting || !model.selected {
			return "gpt-4o-mini" // default fallback
		}
		return model.GetSelectedValue()
	}
	return "gpt-4o-mini" // default fallback
}

func getAIModels() []string {
	apiURL, headers := InitializeApi()
	if apiURL == "" {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error: apiURL vacío al obtener modelos de AI"))
		return []string{}
	}

	req, err := http.NewRequest("GET", apiURL+"/ai/ai/models", nil)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creando la petición HTTP para obtener modelos de AI: "+err.Error()))
		return []string{}
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error realizando la petición HTTP para obtener modelos de AI: "+err.Error()))
		return []string{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render(fmt.Sprintf("Error: status code %d al obtener modelos de AI", resp.StatusCode)))
		return []string{}
	}

	var response aiModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error decodificando respuesta de modelos de AI: "+err.Error()))
		return []string{}
	}

	return response.Models
}

func askToChangeModel(currentModel string) bool {
	choices := []string{
		fmt.Sprintf("Mantener modelo actual (%s)", currentModel),
		"Cambiar modelo",
	}

	p := tea.NewProgram(inputs.SelectionModel(choices, "MODELO ACTUAL", "¿Qué quieres hacer con el modelo?"), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running menu: "+err.Error()))
		return false
	}

	if model, ok := m.(inputs.SelectionModelS); ok {
		if model.Quitting || !model.Selected {
			return false
		}
		return model.Cursor == 1 // True if "Cambiar modelo" was selected
	}
	return false
}

func getUserPrompt(args []string) string {
	// Join prompt words into a single string
	promptText := strings.Join(args, " ")

	// If no prompt provided, ask user for input
	if promptText == "" {
		p := tea.NewProgram(inputs.TextInput("Tu pregunta: ", "¿En qué puedo ayudarte?"), tea.WithAltScreen())
		m, err := p.Run()
		if err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running input: "+err.Error()))
			return ""
		}

		if model, ok := m.(inputs.TextInputModel); ok {
			if model.TextInput.Value() != "" {
				promptText = model.TextInput.Value()
			}
		}
	}

	if promptText == "" {
		fmt.Println("No request entered. Operation cancelled.")
		return ""
	}

	return promptText
}

// Versión mejorada que muestra el historial durante la entrada de texto
func getUserPromptWithHistory(args []string, conversation *Conversation) string {
	// Join prompt words into a single string
	promptText := strings.Join(args, " ")

	// If no prompt provided, ask user for input
	if promptText == "" {
		// Mostrar historial antes de pedir la nueva pregunta
		if conversation != nil && len(conversation.Messages) > 0 {
			displayConversationHistory(conversation)
		}

		p := tea.NewProgram(inputs.TextInput("Tu nueva pregunta: ", "¿En qué puedo ayudarte?"), tea.WithAltScreen())
		m, err := p.Run()
		if err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running input: "+err.Error()))
			return ""
		}

		if model, ok := m.(inputs.TextInputModel); ok {
			if model.TextInput.Value() != "" {
				promptText = model.TextInput.Value()
			}
		}
	}

	if promptText == "" {
		fmt.Println("No request entered. Operation cancelled.")
		return ""
	}

	return promptText
}

func makeAIRequest(text, model string, messagesHistory []map[string]interface{}) string {
	requestBody := aiRequest{
		Text:            text,
		Model:           model,
		MessagesHistory: messagesHistory,
	}

	apiURL, headers := InitializeApi()
	if apiURL == "" {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error: Could not initialize API connection"))
		return ""
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error marshalling request body: "+err.Error()))
		return ""
	}

	req, err := http.NewRequest("POST", apiURL+"/ai/ai", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creating request: "+err.Error()))
		return ""
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error making request: "+err.Error()))
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Leer el cuerpo de la respuesta para obtener detalles del error
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("%s\n", inputs.ErrorStyle.Render(fmt.Sprintf("API returned status code %d: %s", resp.StatusCode, string(body))))
		return ""
	}

	var response aiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error decoding response: "+err.Error()))
		return ""
	}

	return response.Response
}

func saveConversation(conversation *Conversation, aiType string) {
	configPath := viper.GetString("config_path")
	if configPath == "" {
		configPath = filepath.Join(os.Getenv("HOME"), ".config", "orgm")
	}

	conversationsPath := filepath.Join(configPath, "ai", "conversacion", aiType)
	filename := fmt.Sprintf("%s_%s.json", conversation.Name, conversation.CreatedAt)
	filePath := filepath.Join(conversationsPath, filename)

	// Si la carpeta no existe, créala
	if _, err := os.Stat(conversationsPath); os.IsNotExist(err) {
		if err := os.MkdirAll(conversationsPath, 0755); err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creando carpeta de conversaciones: "+err.Error()))
			return
		}
	}

	data, err := json.MarshalIndent(conversation, "", "  ")
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error saving conversation: "+err.Error()))
		return
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error writing conversation file: "+err.Error()))
		return
	}
}

func exportAIConversationToTXT() {
	conversations := listConversations("ai")
	if len(conversations) == 0 {
		fmt.Println(inputs.ErrorStyle.Render("No hay conversaciones de AI para exportar"))
		return
	}

	// Sort by creation time (most recent first)
	sort.Slice(conversations, func(i, j int) bool {
		return conversations[i].CreatedAt > conversations[j].CreatedAt
	})

	var choices []string
	for _, conv := range conversations {
		displayName := fmt.Sprintf("%s (%s)", conv.Name, conv.CreatedAt)
		choices = append(choices, displayName)
	}

	p := tea.NewProgram(inputs.SelectionModel(choices, "EXPORTAR A TXT", "Selecciona conversación para exportar"), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running menu: "+err.Error()))
		return
	}

	if model, ok := m.(inputs.SelectionModelS); ok {
		if model.Quitting || !model.Selected {
			return
		}

		selectedConversation := &conversations[model.Cursor]
		generateConversationTXT(selectedConversation, "AI")
	}
}

func generateConversationTXT(conversation *Conversation, aiType string) {
	// Create a text file with the conversation
	var content strings.Builder
	content.WriteString(fmt.Sprintf("Conversación %s: %s\n", aiType, conversation.Name))
	content.WriteString(fmt.Sprintf("Fecha de creación: %s\n", conversation.CreatedAt))
	content.WriteString(fmt.Sprintf("Modelo: %s\n", conversation.Model))
	content.WriteString(fmt.Sprintf("Configuración: %s\n\n", conversation.ConfigUsed))
	content.WriteString(strings.Repeat("=", 80) + "\n\n")

	for _, message := range conversation.Messages {
		role := message["role"].(string)
		content_text := message["content"].(string)

		if role == "system" {
			continue // Skip system messages for TXT
		}

		content.WriteString(fmt.Sprintf("[%s]\n", strings.ToUpper(role)))
		content.WriteString(content_text + "\n\n")
		content.WriteString(strings.Repeat("-", 40) + "\n\n")
	}

	// Create directory if it doesn't exist
	configPath := viper.GetString("config_path")
	if configPath == "" {
		configPath = filepath.Join(os.Getenv("HOME"), ".config", "orgm")
	}

	txtPath := filepath.Join(configPath, "ai", "txt")
	if _, err := os.Stat(txtPath); os.IsNotExist(err) {
		if err := os.MkdirAll(txtPath, 0755); err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creando carpeta de TXT: "+err.Error()))
			return
		}
	}

	// Generate filename
	filename := fmt.Sprintf("%s_conversation_%s_%s.txt", strings.ToLower(aiType), conversation.Name, conversation.CreatedAt)
	// Replace spaces and special characters for filename
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, ":", "-")

	filePath := filepath.Join(txtPath, filename)

	if err := os.WriteFile(filePath, []byte(content.String()), 0644); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creating file: "+err.Error()))
		return
	}

	fmt.Printf("%s\n", inputs.SuccessStyle.Render("Conversación exportada a: "+filePath))
}

// Función auxiliar para renderizar markdown usando Glamour
func renderMarkdown(content string) string {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		// Si falla Glamour, devolver el contenido sin formato
		return content
	}

	out, err := r.Render(content)
	if err != nil {
		// Si falla el renderizado, devolver el contenido sin formato
		return content
	}

	return out
}

// Función auxiliar para mostrar el historial de conversación
func displayConversationHistory(conversation *Conversation) {
	if len(conversation.Messages) == 0 {
		return
	}

	fmt.Printf("%s: %s | %s: %s\n",
		inputs.InfoStyle.Render("Conversación"), conversation.Name,
		inputs.InfoStyle.Render("Modelo"), conversation.Model)
	fmt.Println()

	for _, message := range conversation.Messages {
		role := message["role"].(string)
		content := message["content"].(string)

		// Saltar mensajes del sistema para el display
		if role == "system" {
			continue
		}

		// Mostrar el mensaje con estilo
		if role == "user" {
			fmt.Printf("%s %s\n", inputs.CursorStyle.Render("👤"), inputs.ItemStyle.Render(content))
		} else if role == "assistant" {
			fmt.Printf("%s\n", inputs.CheckedStyle.Render("🤖 Respuesta:"))
			// Renderizar la respuesta del asistente en markdown
			markdownContent := renderMarkdown(content)
			fmt.Print(markdownContent)
		}
		fmt.Println()
	}
}

// Función para limpiar el historial y asegurar alternancia correcta
func cleanMessageHistory(messages []map[string]interface{}) []map[string]interface{} {
	if len(messages) == 0 {
		// Si no hay mensajes, devolver configuración por defecto
		return getDefaultConfig()
	}

	cleaned := []map[string]interface{}{}
	var lastRole string
	hasSystemMessage := false

	for _, message := range messages {
		role := message["role"].(string)
		content := message["content"].(string)

		// Verificar si hay mensaje del sistema
		if role == "system" {
			hasSystemMessage = true
		}

		// Saltar mensajes vacíos o que contengan errores
		if content == "" || strings.Contains(content, "Error:") || strings.Contains(content, "API returned status code") {
			continue
		}

		// Saltar mensajes duplicados del mismo rol (excepto system)
		if role == lastRole && role != "system" {
			continue
		}

		cleaned = append(cleaned, message)
		lastRole = role
	}

	// Si no hay mensaje del sistema, agregar uno al inicio
	if !hasSystemMessage {
		defaultConfig := getDefaultConfig()
		cleaned = append(defaultConfig, cleaned...)
	}

	return cleaned
}

// Nuevo tipo para manejar la selección con filtrado
type FilterableSelectionModel struct {
	choices    []string
	filtered   []string
	cursor     int
	selected   bool
	quitting   bool
	title      string
	subtitle   string
	filterMode bool
	filterText string
}

func NewFilterableSelectionModel(choices []string, title, subtitle string) FilterableSelectionModel {
	return FilterableSelectionModel{
		choices:    choices,
		filtered:   choices,
		title:      title,
		subtitle:   subtitle,
		filterMode: false,
	}
}

func (m FilterableSelectionModel) Init() tea.Cmd {
	return nil
}

func (m FilterableSelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filterMode {
			switch msg.String() {
			case "enter":
				m.filterMode = false
				m.applyFilter()
				m.cursor = 0
			case "esc":
				m.filterMode = false
				m.filterText = ""
				m.filtered = m.choices
				m.cursor = 0
			case "backspace":
				if len(m.filterText) > 0 {
					m.filterText = m.filterText[:len(m.filterText)-1]
					m.applyFilter()
					m.cursor = 0
				}
			default:
				if len(msg.String()) == 1 {
					m.filterText += msg.String()
					m.applyFilter()
					m.cursor = 0
				}
			}
		} else {
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			case "/":
				m.filterMode = true
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.filtered)-1 {
					m.cursor++
				}
			case "enter":
				if len(m.filtered) > 0 {
					m.selected = true
					return m, tea.Quit
				}
			}
		}
	}

	return m, nil
}

func (m *FilterableSelectionModel) applyFilter() {
	if m.filterText == "" {
		m.filtered = m.choices
		return
	}

	m.filtered = []string{}
	filterLower := strings.ToLower(m.filterText)
	for _, choice := range m.choices {
		if strings.Contains(strings.ToLower(choice), filterLower) {
			m.filtered = append(m.filtered, choice)
		}
	}
}

func (m FilterableSelectionModel) View() string {
	s := inputs.TitleStyle.Render(m.title) + "\n\n"
	s += inputs.SubtitleStyle.Render(m.subtitle) + "\n\n"

	// Mostrar estado del filtro
	if m.filterMode {
		s += inputs.InfoStyle.Render("Filtro: ") + m.filterText + "█\n\n"
	} else if m.filterText != "" {
		s += inputs.InfoStyle.Render("Filtro activo: ") + m.filterText + " (presiona '/' para editar)\n\n"
	}

	// Mostrar opciones filtradas
	const itemsPerPage = 10
	totalItems := len(m.filtered)

	visibleStart := 0
	visibleEnd := totalItems

	if totalItems > itemsPerPage {
		if m.cursor >= itemsPerPage/2 {
			visibleStart = m.cursor - itemsPerPage/2
		}

		if visibleStart+itemsPerPage < totalItems {
			visibleEnd = visibleStart + itemsPerPage
		} else {
			visibleEnd = totalItems
			visibleStart = totalItems - itemsPerPage
			if visibleStart < 0 {
				visibleStart = 0
			}
		}
	}

	if visibleStart > 0 {
		s += inputs.HelpStyle.Render(fmt.Sprintf("... %d more items above ...", visibleStart)) + "\n\n"
	}

	for i := visibleStart; i < visibleEnd && i < totalItems; i++ {
		choice := m.filtered[i]
		cursor := " "
		if m.cursor == i {
			cursor = inputs.CursorStyle.Render(">")
		}

		itemText := choice
		if m.cursor == i {
			itemText = inputs.SelectedItemStyle.Render(itemText)
		} else {
			itemText = inputs.ItemStyle.Render(itemText)
		}

		s += fmt.Sprintf("%s %s\n", cursor, itemText)
	}

	if visibleEnd < totalItems {
		s += "\n" + inputs.HelpStyle.Render(fmt.Sprintf("... %d more items below ...", totalItems-visibleEnd)) + "\n"
	}

	s += "\n" + inputs.HelpStyle.Render(fmt.Sprintf("Item %d of %d", m.cursor+1, totalItems)) + "\n"

	if m.filterMode {
		s += inputs.HelpStyle.Render("Escribe para filtrar, enter para confirmar, esc para cancelar") + "\n"
	} else {
		s += inputs.HelpStyle.Render("Press ↑/↓ to navigate, enter to select, / to filter, q to quit") + "\n"
	}

	return s
}

func (m FilterableSelectionModel) GetSelectedValue() string {
	if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
		return m.filtered[m.cursor]
	}
	return ""
}

func deleteAIConversations() {
	conversations := listConversations("ai")
	if len(conversations) == 0 {
		fmt.Println(inputs.InfoStyle.Render("No hay conversaciones de AI para eliminar"))
		return
	}

	// Opciones principales
	mainOptions := []string{
		"Eliminar conversaciones específicas",
		"Eliminar TODAS las conversaciones",
		"Cancelar",
	}

	p := tea.NewProgram(inputs.SelectionModel(mainOptions, "ELIMINAR CONVERSACIONES AI", "¿Qué quieres hacer?"), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running menu: "+err.Error()))
		return
	}

	if model, ok := m.(inputs.SelectionModelS); ok {
		if model.Quitting || !model.Selected {
			return
		}

		switch model.Cursor {
		case 0: // Eliminar específicas
			deleteSpecificAIConversations(conversations)
		case 1: // Eliminar todas
			deleteAllAIConversations(conversations)
		case 2: // Cancelar
			fmt.Println(inputs.InfoStyle.Render("Operación cancelada"))
			return
		}
	}
}

func deleteSpecificAIConversations(conversations []Conversation) {
	if len(conversations) == 0 {
		return
	}

	// Sort by creation time (most recent first)
	sort.Slice(conversations, func(i, j int) bool {
		return conversations[i].CreatedAt > conversations[j].CreatedAt
	})

	// Crear lista de choices con información detallada
	var choices []inputs.ItemMS
	for i, conv := range conversations {
		displayName := fmt.Sprintf("%s (%s)", conv.Name, conv.CreatedAt)
		description := fmt.Sprintf("Modelo: %s, Config: %s", conv.Model, conv.ConfigUsed)
		choices = append(choices, inputs.ItemMS{
			Title:       displayName,
			Description: description,
			Value:       fmt.Sprintf("%d", i), // Usar índice como valor
			Checked:     false,
		})
	}

	// Agregar opción de salir
	choices = append(choices, inputs.ItemMS{
		Title:       "Salir",
		Description: "Cancelar la eliminación",
		Value:       "exit",
		Checked:     false,
	})

	// Mostrar menú de selección múltiple
	p := tea.NewProgram(inputs.InitialModelMS(choices), tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running menu: "+err.Error()))
		return
	}

	if model, ok := result.(inputs.ModelMS); ok {
		if model.Quitting {
			return
		}

		// Obtener conversaciones seleccionadas
		var selectedConversations []Conversation
		for _, choice := range model.Choices {
			if choice.Checked && choice.Value != "exit" {
				if idx := choice.Value; idx != "exit" {
					// Convertir el valor a índice
					var convIdx int
					fmt.Sscanf(idx, "%d", &convIdx)
					if convIdx < len(conversations) {
						selectedConversations = append(selectedConversations, conversations[convIdx])
					}
				}
			}
		}

		if len(selectedConversations) == 0 {
			fmt.Println(inputs.InfoStyle.Render("No se seleccionaron conversaciones para eliminar"))
			return
		}

		// Confirmar eliminación
		confirmOptions := []string{"Sí, eliminar", "No, cancelar"}
		confirmMsg := fmt.Sprintf("¿Estás seguro de eliminar %d conversación(es)?", len(selectedConversations))

		p := tea.NewProgram(inputs.SelectionModel(confirmOptions, "CONFIRMAR ELIMINACIÓN", confirmMsg), tea.WithAltScreen())
		m, err := p.Run()
		if err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running confirmation: "+err.Error()))
			return
		}

		if model, ok := m.(inputs.SelectionModelS); ok {
			if model.Quitting || !model.Selected || model.Cursor != 0 {
				fmt.Println(inputs.InfoStyle.Render("Eliminación cancelada"))
				return
			}

			// Eliminar conversaciones seleccionadas
			deleted := 0
			for _, conv := range selectedConversations {
				if deleteConversationFile(conv, "ai") {
					deleted++
				}
			}

			fmt.Printf("%s\n", inputs.SuccessStyle.Render(fmt.Sprintf("Se eliminaron %d conversación(es) exitosamente", deleted)))
		}
	}
}

func deleteAllAIConversations(conversations []Conversation) {
	if len(conversations) == 0 {
		return
	}

	// Confirmar eliminación de todas
	confirmOptions := []string{"Sí, eliminar TODAS", "No, cancelar"}
	confirmMsg := fmt.Sprintf("¿Estás SEGURO de eliminar TODAS las %d conversaciones de AI? Esta acción no se puede deshacer.", len(conversations))

	p := tea.NewProgram(inputs.SelectionModel(confirmOptions, "⚠️  ELIMINAR TODAS LAS CONVERSACIONES", confirmMsg), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running confirmation: "+err.Error()))
		return
	}

	if model, ok := m.(inputs.SelectionModelS); ok {
		if model.Quitting || !model.Selected || model.Cursor != 0 {
			fmt.Println(inputs.InfoStyle.Render("Eliminación cancelada"))
			return
		}

		// Eliminar todas las conversaciones
		deleted := 0
		for _, conv := range conversations {
			if deleteConversationFile(conv, "ai") {
				deleted++
			}
		}

		fmt.Printf("%s\n", inputs.SuccessStyle.Render(fmt.Sprintf("Se eliminaron TODAS las %d conversaciones exitosamente", deleted)))
	}
}

func deleteConversationFile(conversation Conversation, aiType string) bool {
	configPath := viper.GetString("config_path")
	if configPath == "" {
		configPath = filepath.Join(os.Getenv("HOME"), ".config", "orgm")
	}

	conversationsPath := filepath.Join(configPath, "ai", "conversacion", aiType)
	filename := fmt.Sprintf("%s_%s.json", conversation.Name, conversation.CreatedAt)
	filePath := filepath.Join(conversationsPath, filename)

	err := os.Remove(filePath)
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error eliminando "+filename+": "+err.Error()))
		return false
	}
	return true
}

func init() {
	RootCmd.AddCommand(aiCmd)
	aiCmd.AddCommand(aiTxtCmd)
	aiCmd.AddCommand(aiDeleteCmd)
}
