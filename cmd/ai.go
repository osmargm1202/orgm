/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/osmargm1202/orgm/inputs"
	"github.com/spf13/cobra"
)

// aiCmd represents the ai command
var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "AI command",
	Long:  `interaction with OpenAI usign a specific config file. api endpoint AI. default config file is -c terminal`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("ai called")
	// },
}

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Prompt command",
	Long:  `Prompt to send to the AI`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(inputs.CommandStyle.Render(Prompt(args, "")))
	},
}

var terminalCmd = &cobra.Command{
	Use:   "terminal",
	Short: "Terminal command from AI",
	Long:  `get a terminal command from windows or linux asking to the AI`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(inputs.CommandStyle.Render(Prompt(args, "terminal")))
	},
}

var configsCmd = &cobra.Command{
	Use:   "configs",
	Short: "Configs command",
	Long:  `Get the configs for the AI`,
	Run: func(cmd *cobra.Command, args []string) {
		GetConfigs()
	},
}

// DivisaRequest represents the request body for currency conversion
type aiRequest struct {
	Prompt string `json:"text"`
	Config string `json:"config"`
}

type aiResponse struct {
	Response string `json:"response"`
}

func Prompt(prompt []string, config string) string {
	if config == "" {
		config = SelectConfig("CONFIG SELECTOR", "Select a configfile on server to use")
	}
	// Join prompt words into a single string
	promptText := strings.Join(prompt, " ")

	// If no prompt provided, ask user for input
	if promptText == "" {
		p := tea.NewProgram(inputs.TextInput("Enter your AI request: ", "how to install orgm?"))
		m, err := p.Run()
		if err != nil {
			fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running menu: "+err.Error()))
			promptText = ""
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

	requestBody := aiRequest{
		Prompt: promptText,
		Config: config,
	}

	// Get API URL and headers
	apiURL, headers := InitializeApi()
	if apiURL == "" {
		fmt.Println("Could not initialize API connection")
		return ""
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error marshalling request body:", err)
		return ""
	}

	req, err := http.NewRequest("POST", apiURL+"/ai/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return ""
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// Make request with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return ""
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		fmt.Println("API returned status code:", resp.StatusCode)
		return ""
	}

	// Parse response
	var response aiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Println("Error decoding response:", err)
		return ""
	}

	return response.Response

}

func GetConfigs() []string {
	apiURL, headers := InitializeApi()
	if apiURL == "" {
		fmt.Println("Could not initialize API connection")
		return []string{}
	}
	req, err := http.NewRequest("GET", apiURL+"/ai/configs", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
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
		fmt.Println("Error making request:", err)
		return []string{}
	}
	defer resp.Body.Close()

	var configsList []string
	if err := json.NewDecoder(resp.Body).Decode(&configsList); err != nil {
		fmt.Println("Error decoding response:", err)
		return []string{}
	}

	// fmt.Println(configsList)

	return configsList
}

func SelectConfig(title, subtitle string) string {
	lista := GetConfigs()
	p := tea.NewProgram(inputs.InitialModelS(lista, title, subtitle))
	m, err := p.Run()
	if err != nil {
		fmt.Printf("%s\n", inputs.ErrorStyle.Render("Error running menu: "+err.Error()))
		return ""
	}

	if model, ok := m.(inputs.ModelS); ok {
		if model.Quitting {
			return ""
		}
		if model.Selected {
			return model.Choices[model.Cursor].Title
		}
	}

	return ""
}

func init() {
	RootCmd.AddCommand(aiCmd)
	aiCmd.AddCommand(promptCmd)
	aiCmd.AddCommand(configsCmd)
	aiCmd.AddCommand(terminalCmd)
}
