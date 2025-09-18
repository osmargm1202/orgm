package inputs

import (
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// TextAreaModel represents a multi-line text input model
type TextAreaModel struct {
	text      []string
	cursor    int
	line      int
	width     int
	height    int
	question  string
	placeholder string
	done      bool
}

// getClipboardContent retrieves content from clipboard
func getClipboardContent() string {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard", "-o")
	case "darwin":
		cmd = exec.Command("pbpaste")
	case "windows":
		cmd = exec.Command("powershell", "Get-Clipboard")
	default:
		return ""
	}
	
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	return strings.TrimSpace(string(output))
}

// TextArea creates a new textarea input model
func TextArea(question, placeholder string) TextAreaModel {
	return TextAreaModel{
		text:       []string{""},
		cursor:     0,
		line:       0,
		width:      80,
		height:     10,
		question:   question,
		placeholder: placeholder,
		done:       false,
	}
}

func (m TextAreaModel) Init() tea.Cmd {
	return nil
}

func (m TextAreaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// Add new line
			newLine := m.text[m.line][:m.cursor] + m.text[m.line][m.cursor:]
			m.text = append(m.text[:m.line+1], append([]string{""}, m.text[m.line+1:]...)...)
			m.text[m.line] = newLine
			m.line++
			m.cursor = 0
		case tea.KeyBackspace:
			if m.cursor > 0 {
				// Delete character
				m.text[m.line] = m.text[m.line][:m.cursor-1] + m.text[m.line][m.cursor:]
				m.cursor--
			} else if m.line > 0 {
				// Move to end of previous line
				prevLineLen := len(m.text[m.line-1])
				m.text[m.line-1] += m.text[m.line]
				m.text = append(m.text[:m.line], m.text[m.line+1:]...)
				m.line--
				m.cursor = prevLineLen
			}
		case tea.KeyDelete:
			if m.cursor < len(m.text[m.line]) {
				m.text[m.line] = m.text[m.line][:m.cursor] + m.text[m.line][m.cursor+1:]
			}
		case tea.KeyLeft:
			if m.cursor > 0 {
				m.cursor--
			} else if m.line > 0 {
				m.line--
				m.cursor = len(m.text[m.line])
			}
		case tea.KeyRight:
			if m.cursor < len(m.text[m.line]) {
				m.cursor++
			} else if m.line < len(m.text)-1 {
				m.line++
				m.cursor = 0
			}
		case tea.KeyUp:
			if m.line > 0 {
				m.line--
				if m.cursor > len(m.text[m.line]) {
					m.cursor = len(m.text[m.line])
				}
			}
		case tea.KeyDown:
			if m.line < len(m.text)-1 {
				m.line++
				if m.cursor > len(m.text[m.line]) {
					m.cursor = len(m.text[m.line])
				}
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			m.done = true
			return m, tea.Quit
		case tea.KeyCtrlS:
			// Save and exit
			m.done = true
			return m, tea.Quit
		case tea.KeyCtrlV:
			// Paste from clipboard
			clipboardContent := getClipboardContent()
			if clipboardContent != "" {
				// Split clipboard content by lines
				lines := strings.Split(clipboardContent, "\n")
				if len(lines) == 1 {
					// Single line - insert at cursor
					m.text[m.line] = m.text[m.line][:m.cursor] + lines[0] + m.text[m.line][m.cursor:]
					m.cursor += len(lines[0])
				} else {
					// Multiple lines - insert at cursor and add new lines
					firstLine := m.text[m.line][:m.cursor] + lines[0]
					lastLine := lines[len(lines)-1] + m.text[m.line][m.cursor:]
					
					// Insert middle lines
					middleLines := lines[1 : len(lines)-1]
					
					// Rebuild text array
					newText := make([]string, 0, len(m.text)+len(middleLines))
					newText = append(newText, m.text[:m.line]...)
					newText = append(newText, firstLine)
					newText = append(newText, middleLines...)
					newText = append(newText, lastLine)
					newText = append(newText, m.text[m.line+1:]...)
					
					m.text = newText
					m.line += len(middleLines)
					m.cursor = len(lastLine)
				}
			}
		default:
			// Insert character
			if len(msg.String()) == 1 {
				char := msg.String()
				m.text[m.line] = m.text[m.line][:m.cursor] + char + m.text[m.line][m.cursor:]
				m.cursor++
			}
		}
	}

	return m, nil
}

func (m TextAreaModel) View() string {
	var sb strings.Builder
	
	// Question
	sb.WriteString(TitleStyle.Render(m.question))
	sb.WriteString("\n\n")
	
	// Instructions
	sb.WriteString(HelpStyle.Render("Enter your prompt (Ctrl+S to save, Ctrl+V to paste, Ctrl+C to cancel):"))
	sb.WriteString("\n\n")
	
	// Text area
	for i, line := range m.text {
		lineText := line
		if lineText == "" && i == m.line {
			lineText = m.placeholder
		}
		
		// Highlight current line
		if i == m.line {
			lineText = SelectedItemStyle.Render(lineText)
		} else {
			lineText = ItemStyle.Render(lineText)
		}
		
		// Add cursor
		if i == m.line {
			if m.cursor < len(line) {
				before := line[:m.cursor]
				after := line[m.cursor:]
				lineText = before + "|" + after
			} else {
				lineText += "|"
			}
		}
		
		sb.WriteString(lineText)
		sb.WriteString("\n")
	}
	
	// Help text
	sb.WriteString("\n")
	sb.WriteString(HelpStyle.Render("Use arrow keys to navigate, Enter for new line"))
	
	return sb.String()
}

// GetTextArea presents a textarea input and returns the user's input
func GetTextArea(question, placeholder string) string {
	model := TextArea(question, placeholder)
	p := tea.NewProgram(model, tea.WithAltScreen())
	result, _ := p.Run()
	finalModel := result.(TextAreaModel)
	
	if !finalModel.done {
		return ""
	}
	
	// Join all lines and trim
	return strings.TrimSpace(strings.Join(finalModel.text, "\n"))
}
