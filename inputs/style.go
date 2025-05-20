package inputs

import (
	"github.com/charmbracelet/lipgloss"
)

// Estilos para el menú
var (
	TitleStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4")).Padding(0, 1)
	SubtitleStyle     = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#ABABAB"))
	CursorStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	CheckedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#73F59F"))
	UncheckedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F27878"))
	ItemStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	SelectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#74ACDF"))
	DescriptionStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	HelpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#6A6A6A")).Italic(true)

	// Estilos para los mensajes de salida
	SuccessStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#73F59F"))
	ErrorStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F27878"))
	InfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#2020F4"))
	WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAB26"))
	CommandStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#3B9FEF"))
)

type (
	errMsg error
)
