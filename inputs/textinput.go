package inputs

// A simple program demonstrating the text input component from the Bubbles
// component library.

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// func main() {
// 	p := tea.NewProgram(initialModel())
// 	if _, err := p.Run(); err != nil {
// 		log.Fatal(err)
// 	}
// }

type TextInputModel struct {
	TextInput textinput.Model
	Err       error
	Question  string
}

func TextInput(question, placeholder string) TextInputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 80

	return TextInputModel{
		TextInput: ti,
		Err:       nil,
		Question:  question,
	}
}

func (m TextInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TextInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case errMsg:
		m.Err = msg
		return m, nil
	}

	m.TextInput, cmd = m.TextInput.Update(msg)
	return m, cmd
}
func (m TextInputModel) View() string {
	var sb strings.Builder
	sb.WriteString(m.Question)
	sb.WriteString("\n\n")
	sb.WriteString(m.TextInput.View())
	if m.TextInput.Focused() {
		sb.WriteString("\n\n(esc to quit)")
	}
	sb.WriteString("\n")
	return sb.String()
}

// GetInput presents a text input prompt and returns the user's input
func GetInput(question string) string {
	model := TextInput(question, "")
	p := tea.NewProgram(model, tea.WithAltScreen())
	result, _ := p.Run()
	finalModel := result.(TextInputModel)
	return strings.TrimSpace(finalModel.TextInput.Value())
}

// GetInputWithDefault presents a text input prompt with a default value and returns the user's input
func GetInputWithDefault(question, defaultValue string) string {
	model := TextInput(question, defaultValue)
	model.TextInput.SetValue(defaultValue)
	p := tea.NewProgram(model, tea.WithAltScreen())
	result, _ := p.Run()
	finalModel := result.(TextInputModel)
	value := strings.TrimSpace(finalModel.TextInput.Value())
	if value == "" {
		return defaultValue
	}
	return value
}
