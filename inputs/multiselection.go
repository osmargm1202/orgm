package inputs

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)


// Bubble tea model for Docker menu
type ItemMS struct {
	Title       string
	Description string
	Value       string
	Checked     bool
}

type ModelMS struct {
	Choices  []ItemMS
	Cursor   int
	Selected map[int]struct{}
	Quitting bool
}

func InitialModelMS(lista []ItemMS) ModelMS {
	choices := lista

	return ModelMS{
		Choices:  choices,
		Selected: make(map[int]struct{}),
	}
}

func (m ModelMS) Init() tea.Cmd {
	return nil
}

func (m ModelMS) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.Quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Choices)-1 {
				m.Cursor++
			}
		case "enter":
			// Check if "Salir" is selected
			if m.Choices[m.Cursor].Value == "exit" {
				m.Quitting = true
				return m, tea.Quit
			}
			// Toggle selection
			if _, ok := m.Selected[m.Cursor]; ok {
				delete(m.Selected, m.Cursor)
				m.Choices[m.Cursor].Checked = false
			} else {
				m.Selected[m.Cursor] = struct{}{}
				m.Choices[m.Cursor].Checked = true
			}
		case " ":
			// Toggle selection
			if _, ok := m.Selected[m.Cursor]; ok {
				delete(m.Selected, m.Cursor)
				m.Choices[m.Cursor].Checked = false
			} else {
				m.Selected[m.Cursor] = struct{}{}
				m.Choices[m.Cursor].Checked = true
			}
		case "tab":
			// Run selected commands
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m ModelMS) View() string {
	// Título y subtítulo
	s := TitleStyle.Render("ORGM DOCKER MENU") + "\n\n"
	s += SubtitleStyle.Render("Select operations to execute (space to select, enter to toggle, tab to run)") + "\n\n"

	for i, choice := range m.Choices {
		// Cursor
		cursor := " "
		if m.Cursor == i {
			cursor = CursorStyle.Render(">")
		} else {
			cursor = " "
		}

		// Checkbox
		checked := " "
		if choice.Checked {
			checked = CheckedStyle.Render("✓")
		} else {
			checked = UncheckedStyle.Render(" ")
		}

		// Item title
		itemText := choice.Title
		if m.Cursor == i {
			itemText = SelectedItemStyle.Render(itemText)
		} else {
			itemText = ItemStyle.Render(itemText)
		}

		// Description
		desc := DescriptionStyle.Render(choice.Description)

		// Combine all parts
		s += fmt.Sprintf("%s [%s] %s - %s\n", cursor, checked, itemText, desc)
	}

	s += "\n" + HelpStyle.Render("Press q to quit, space to select/deselect, and tab to run selected commands.") + "\n"

	return s
}