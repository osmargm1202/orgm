package inputs

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// Item represents an item in a selection list
type Item struct {
	ID   string
	Name string
	Desc string
}

// Custom implementation for list.Item interface
type listItemDelegate struct {
	item Item
}

func newListItem(item Item) listItemDelegate {
	return listItemDelegate{item: item}
}

func (i listItemDelegate) FilterValue() string { return i.item.Name }
func (i listItemDelegate) Title() string       { return i.item.Name }
func (i listItemDelegate) Description() string { return i.item.Desc }

// Bubble tea model for Config menu
type SelectionItemS struct {
	Title string
}

type SelectionModelS struct {
	Choices  []SelectionItemS
	Cursor   int
	Selected bool
	Quitting bool
	Title    string
	Subtitle string
}

func SelectionModel(list []string, title, subtitle string) SelectionModelS {

	choices := []SelectionItemS{}
	for _, i := range list {
		choices = append(choices, SelectionItemS{Title: i})
	}

	return SelectionModelS{
		Choices:  choices,
		Title:    title,
		Subtitle: subtitle,
	}
}

func (m SelectionModelS) Init() tea.Cmd {
	return nil
}

func (m SelectionModelS) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m SelectionModelS) View() string {
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

// SelectList displays a list of items and returns the selected item
func SelectList(title string, items []Item) Item {
	var listItems []list.Item
	// Create a map to store the original items by their list item counterparts
	itemMap := make(map[string]Item)

	for _, item := range items {
		listItem := newListItem(item)
		listItems = append(listItems, listItem)
		itemMap[item.ID] = item
	}

	// Configurar un tamaño adecuado para la lista
	width, height := 80, 15 // Tamaño explícito para mejor visibilidad

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true // Mostrar descripción para más información
	delegate.SetHeight(2)           // Dos líneas por elemento

	l := list.New(listItems, delegate, width, height)
	l.Title = title
	l.SetShowStatusBar(true)
	l.SetShowHelp(true)
	l.Styles.Title = l.Styles.Title.MarginLeft(2).Bold(true)

	m := listModel{
		list:    l,
		itemMap: itemMap,
	}
	p := tea.NewProgram(m)

	result, _ := p.Run()
	finalModel := result.(listModel)

	if finalModel.selected {
		selectedListItem, ok := finalModel.list.SelectedItem().(listItemDelegate)
		if ok {
			return finalModel.itemMap[selectedListItem.item.ID]
		}
	}

	// Return empty item if no selection was made
	return Item{}
}

type listModel struct {
	list     list.Model
	selected bool
	itemMap  map[string]Item
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			m.selected = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listModel) View() string {
	return m.list.View()
}
