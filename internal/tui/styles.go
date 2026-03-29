package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette Symphony
var (
	ColorBrand     = lipgloss.Color("#5F87FF")
	ColorSuccess   = lipgloss.Color("#04B575")
	ColorWarning   = lipgloss.Color("#FFD700")
	ColorDanger    = lipgloss.Color("#FF5F57")
	ColorMuted     = lipgloss.Color("#626262")
	ColorHighlight = lipgloss.Color("#D787FF")
)

// Base styles — semua komponen derive dari sini
var (
	StyleBase = lipgloss.NewStyle()

	StyleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBrand).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(ColorMuted).
			Width(54).
			Padding(0, 1)

	StyleDivider = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StyleSuccess = lipgloss.NewStyle().Foreground(ColorSuccess)
	StyleWarning = lipgloss.NewStyle().Foreground(ColorWarning)
	StyleDanger  = lipgloss.NewStyle().Foreground(ColorDanger)
	StyleMuted   = lipgloss.NewStyle().Foreground(ColorMuted)
	StyleBrand   = lipgloss.NewStyle().Foreground(ColorBrand)
	StyleHighlight = lipgloss.NewStyle().Foreground(ColorHighlight)

	// Label aksi file pada preview
	StyleActionCreate = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
	StyleActionModify = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)
	StyleActionSkip   = lipgloss.NewStyle().Foreground(ColorMuted)
	StyleActionDelete = lipgloss.NewStyle().Foreground(ColorDanger).Bold(true)
)

// Divider — garis pemisah horizontal
func Divider(width int) string {
	return StyleDivider.Render(strings.Repeat("─", width))
}

// Icon helpers
const (
	IconSuccess = "✔"
	IconError   = "✖"
	IconWarning = "⚠"
	IconArrow   = "❯"
	IconDiamond = "◆"
	IconSpinner = "⠸"
)
