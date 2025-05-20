package inputs

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Bubble tea model for Config menu
type ItemS struct {
	Title string
}

type ModelS struct {
	Choices  []ItemS
	Cursor   int
	Selected bool
	Quitting bool
	Title    string
	Subtitle string
}

func InitialModelS(list []string, title, subtitle string) ModelS {

	choices := []ItemS{}
	for _, i := range list {
		choices = append(choices, ItemS{Title: i})
	}

	return ModelS{
		Choices:  choices,
		Title:    title,
		Subtitle: subtitle,
	}
}

func (m ModelS) Init() tea.Cmd {
	return nil
}

func (m ModelS) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			m.Selected = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m ModelS) View() string {
	s := TitleStyle.Render(m.Title) + "\n\n"
	s += SubtitleStyle.Render(m.Subtitle) + "\n\n"

	for i, choice := range m.Choices {
		cursor := " "
		if m.Cursor == i {
			cursor = CursorStyle.Render(">")
		}

		itemText := choice.Title
		if m.Cursor == i {
			itemText = SelectedItemStyle.Render(itemText)
		} else {
			itemText = ItemStyle.Render(itemText)
		}

		s += fmt.Sprintf("%s %s\n", cursor, itemText)
	}

	s += "\n" + HelpStyle.Render("Press q to quit, enter to select") + "\n"

	return s
}
