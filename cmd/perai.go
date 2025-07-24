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
	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Perplexity AI Request/Response structures
type peraiRequest struct {
	Text                string                   `json:"text"`
	Model               string                   `json:"model"`
	SearchRecencyFilter *string                  `json:"search_recency_filter,omitempty"`
	MessagesHistory     []map[string]interface{} `json:"messages_history,omitempty"`
}

type peraiResponse struct {
	Response            string  `json:"response"`
	ModelUsed           string  `json:"model_used"`
	SearchRecencyFilter *string `json:"search_recency_filter"`
	HasHistory          bool    `json:"has_history"`
}

type peraiModelsResponse struct {
	Models []string `json:"models"`
}

// peraiCmd represents the perai command
var peraiCmd = &cobra.Command{
	Use:   "perai",
	Short: "Perplexity AI interaction",
	Long:  `Interact with Perplexity AI models with conversation management, search filters and history`,
	Run: func(cmd *cobra.Command, args []string) {
		startPerAIConversation(args)
	},
}

var peraiTxtCmd = &cobra.Command{
	Use:   "txt",
	Short: "Export PerAI conversation to TXT",
	Long:  `Export a Perplexity AI conversation to TXT format`,
	Run: func(cmd *cobra.Command, args []string) {
		exportPerAIConversationToTXT()
	},
}

var peraiDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete PerAI conversations",
	Long:  `Delete PerAI conversations with options to delete all or select specific ones`,
	Run: func(cmd *cobra.Command, args []string) {
		deletePerAIConversations()
	},
}

func startPerAIConversation(args []string) {
	// Step 1: Ask if user wants new conversation or continue existing
	conversationChoice := selectConversationType("PerAI", "¿Qué quieres hacer?")
	if conversationChoice == "" {
		return
	}

	var conversation *Conversation
	var configMessages []map[string]interface{}

	if conversationChoice == "Nueva conversación" {
		// Step 2: Select configuration for new conversation
		configName := selectPerAIConfigFile()
		if configName == "" {
			return
		}

		// Load configuration messages
		configMessages = loadPerAIConfigMessages(configName)
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
			AIType:     "perai",
			Model:      "", // Will be set when model is selected
			ConfigUsed: configName,
			Messages:   configMessages,
		}
	} else {
		// Continue existing conversation
		conversation = selectExistingConversation("perai")
		if conversation == nil {
			return
		}

		// Limpiar el historial para asegurar alternancia correcta
		conversation.Messages = cleanMessageHistory(conversation.Messages)
	}

	// Step 4: Select model (only if not set or if user wants to change)
	var model string
	if conversation.Model == "" {
		model = selectPerAIModel()
		if model == "" {
			return
		}
		conversation.Model = model
	} else {
		// Ask if user wants to keep the same model or change it
		changeModel := askToChangeModel(conversation.Model)
		if changeModel {
			model = selectPerAIModel()
			if model == "" {
				return
			}
			conversation.Model = model
		} else {
			model = conversation.Model
		}
	}

	// Step 5: Select search recency filter
	recencyFilter := selectRecencyFilter()

	// Step 6: Get user prompt (con historial mejorado)
	promptText := getUserPromptWithHistory(args, conversation)
	if promptText == "" {
		return
	}

	// Step 7: Make API request ANTES de agregar el mensaje del usuario
	// Esto evita que se rompa la alternancia si hay un error
	response := makePerAIRequest(promptText, model, recencyFilter, conversation.Messages)
	if response == "" {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error: No se pudo obtener respuesta de la API"))
		return
	}

	// Solo si la API responde exitosamente, agregamos los mensajes
	// Step 8: Add user message to conversation
	userMessage := map[string]interface{}{
		"role":    "user",
		"content": promptText,
	}
	conversation.Messages = append(conversation.Messages, userMessage)

	// Step 9: Add assistant response to conversation
	assistantMessage := map[string]interface{}{
		"role":    "assistant",
		"content": response,
	}
	conversation.Messages = append(conversation.Messages, assistantMessage)

	// Step 10: Save conversation
	saveConversation(conversation, "perai")

	// Step 11: Display current question and response (renderizado en markdown)
	fmt.Printf("\n%s\n", inputs.TitleStyle.Render("=== NUEVA INTERACCIÓN ==="))
	fmt.Printf("%s\n", inputs.CursorStyle.Render("👤 TU PREGUNTA:"))
	fmt.Printf("%s\n\n", inputs.ItemStyle.Render(promptText))

	fmt.Printf("%s\n", inputs.CheckedStyle.Render("🤖 RESPUESTA DEL ASISTENTE (PerplexityAI):"))
	markdownResponse := renderMarkdown(response)
	fmt.Print(markdownResponse)
}

func selectPerAIModel() string {
	models := getPerAIModels()

	fmt.Printf("%s\n", inputs.InfoStyle.Render(fmt.Sprintf("Modelos PerAI disponibles: %v", models)))

	if len(models) == 0 {
		fmt.Println(inputs.ErrorStyle.Render("No se pudieron obtener los modelos disponibles"))
		return "sonar-pro" // default fallback
	}

	// Usar el nuevo modelo de selección con filtrado
	filterableModel := NewFilterableSelectionModel(models, "SELECCIONA MODELO", "Selecciona un modelo de Perplexity AI (presiona '/' para filtrar)")
	p := tea.NewProgram(filterableModel, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running menu: "+err.Error()))
		return "sonar-pro" // default fallback
	}

	if model, ok := m.(FilterableSelectionModel); ok {
		if model.quitting || !model.selected {
			return "sonar-pro" // default fallback
		}
		return model.GetSelectedValue()
	}
	return "sonar-pro" // default fallback
}

func getPerAIModels() []string {
	apiURL, headers := InitializeApi()
	if apiURL == "" {
		return []string{}
	}

	req, err := http.NewRequest("GET", apiURL+"/ai/perai/models", nil)
	if err != nil {
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
		return []string{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render(fmt.Sprintf("Error: status code %d al obtener modelos de PerAI", resp.StatusCode)))
		return []string{}
	}

	var response peraiModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error decodificando respuesta de modelos de PerAI: "+err.Error()))
		return []string{}
	}

	return response.Models
}

func selectPerAIConfigFile() string {
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

	peraiConfigPath := filepath.Join(homeDir, appsPath, "ai", "configs", "perai")
	files, err := os.ReadDir(peraiConfigPath)
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

func getDefaultPerAIConfig() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"role":    "system",
			"content": "Eres un asistente de investigación inteligente con acceso a información actualizada. Proporciona respuestas precisas, bien fundamentadas y cita fuentes cuando sea relevante.",
		},
	}
}

func loadPerAIConfigMessages(configName string) []map[string]interface{} {
	// Si se seleccionó "default", devolver configuración por defecto
	if configName == "default" {
		return getDefaultPerAIConfig()
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

	configFile := filepath.Join(homeDir, appsPath, "ai", "configs", "perai", configName+".json")
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

func selectRecencyFilter() *string {
	recencyOptions := []string{
		"Ninguna",
		"Día",
		"Semana",
		"Mes",
		"Año",
	}

	p := tea.NewProgram(inputs.SelectionModel(recencyOptions, "FILTRO DE RECENCIA", "Selecciona filtro de búsqueda temporal"), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running menu: "+err.Error()))
		return nil
	}

	if model, ok := m.(inputs.SelectionModelS); ok {
		if model.Quitting || !model.Selected {
			return nil
		}

		selectedOption := model.Choices[model.Cursor].Title
		switch selectedOption {
		case "Ninguna":
			return nil
		case "Día":
			filter := "day"
			return &filter
		case "Semana":
			filter := "week"
			return &filter
		case "Mes":
			filter := "month"
			return &filter
		case "Año":
			filter := "year"
			return &filter
		default:
			return nil
		}
	}
	return nil
}

func makePerAIRequest(text, model string, recencyFilter *string, messagesHistory []map[string]interface{}) string {
	requestBody := peraiRequest{
		Text:                text,
		Model:               model,
		SearchRecencyFilter: recencyFilter,
		MessagesHistory:     messagesHistory,
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

	req, err := http.NewRequest("POST", apiURL+"/ai/perai", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error creating request: "+err.Error()))
		return ""
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	client := &http.Client{
		Timeout: 120 * time.Second,
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

	var response peraiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error decoding response: "+err.Error()))
		return ""
	}

	return response.Response
}

func exportPerAIConversationToTXT() {
	conversations := listConversations("perai")
	if len(conversations) == 0 {
		fmt.Println(inputs.ErrorStyle.Render("No hay conversaciones de PerAI para exportar"))
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
		generateConversationTXT(selectedConversation, "PerAI")
	}
}

func deletePerAIConversations() {
	conversations := listConversations("perai")
	if len(conversations) == 0 {
		fmt.Println(inputs.InfoStyle.Render("No hay conversaciones de PerAI para eliminar"))
		return
	}

	// Opciones principales
	mainOptions := []string{
		"Eliminar conversaciones específicas",
		"Eliminar TODAS las conversaciones",
		"Cancelar",
	}

	p := tea.NewProgram(inputs.SelectionModel(mainOptions, "ELIMINAR CONVERSACIONES PERAI", "¿Qué quieres hacer?"), tea.WithAltScreen())
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
			deleteSpecificPerAIConversations(conversations)
		case 1: // Eliminar todas
			deleteAllPerAIConversations(conversations)
		case 2: // Cancelar
			fmt.Println(inputs.InfoStyle.Render("Operación cancelada"))
			return
		}
	}
}

func deleteSpecificPerAIConversations(conversations []Conversation) {
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
				if deleteConversationFile(conv, "perai") {
					deleted++
				}
			}

			fmt.Printf("%s\n", inputs.SuccessStyle.Render(fmt.Sprintf("Se eliminaron %d conversación(es) exitosamente", deleted)))
		}
	}
}

func deleteAllPerAIConversations(conversations []Conversation) {
	if len(conversations) == 0 {
		return
	}

	// Confirmar eliminación de todas
	confirmOptions := []string{"Sí, eliminar TODAS", "No, cancelar"}
	confirmMsg := fmt.Sprintf("¿Estás SEGURO de eliminar TODAS las %d conversaciones de PerAI? Esta acción no se puede deshacer.", len(conversations))

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
			if deleteConversationFile(conv, "perai") {
				deleted++
			}
		}

		fmt.Printf("%s\n", inputs.SuccessStyle.Render(fmt.Sprintf("Se eliminaron TODAS las %d conversaciones exitosamente", deleted)))
	}
}

func init() {
	RootCmd.AddCommand(peraiCmd)
	peraiCmd.AddCommand(peraiTxtCmd)
	peraiCmd.AddCommand(peraiDeleteCmd)
}
