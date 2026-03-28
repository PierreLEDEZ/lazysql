package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Lazygit-inspired palette
	ColorPrimary      = lipgloss.Color("#7D56F4")
	ColorSecondary    = lipgloss.Color("#6C71C4")
	ColorBorder       = lipgloss.Color("#555555")
	ColorActiveBorder = lipgloss.Color("#7D56F4")
	ColorError        = lipgloss.Color("#FF5555")
	ColorSuccess      = lipgloss.Color("#50FA7B")
	ColorMuted        = lipgloss.Color("#666666")
	ColorHighlight    = lipgloss.Color("#282A36")
	ColorText         = lipgloss.Color("#F8F8F2")
	ColorStatusBg     = lipgloss.Color("#282A36")
	ColorWarning      = lipgloss.Color("#FFB86C")
	ColorCyan         = lipgloss.Color("#8BE9FD")
	ColorYellow       = lipgloss.Color("#F1FA8C")
	ColorPink         = lipgloss.Color("#FF79C6")
	ColorOrange       = lipgloss.Color("#FFB86C")
)

// Panel title number badge colors — each panel gets a distinct color
var PanelColors = []lipgloss.Color{
	ColorCyan,    // 1: Connections
	ColorSuccess, // 2: Tables
	ColorYellow,  // 3: Structure
	ColorPink,    // 4: Query
	ColorOrange,  // 5: Results
}

var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary).
		PaddingLeft(1)

	StatusBar = lipgloss.NewStyle().
			Background(ColorStatusBg).
			Foreground(ColorText).
			PaddingLeft(1)

	StatusKey = lipgloss.NewStyle().
			Background(ColorStatusBg).
			Foreground(ColorSuccess).
			Bold(true)

	StatusDesc = lipgloss.NewStyle().
			Background(ColorStatusBg).
			Foreground(ColorMuted)

	ErrorText = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	SuccessText = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	MutedText = lipgloss.NewStyle().
			Foreground(ColorMuted)

	WarningText = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	// Modal styles
	ModalBorder = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(ColorPrimary).
			Padding(1, 2)

	ModalTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			MarginBottom(1)

	FormLabel = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true).
			Width(12)

	ConfirmPrompt = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWarning).
			MarginBottom(1)

	// Help styles
	HelpSection = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			MarginTop(1)

	HelpKey = lipgloss.NewStyle().
		Foreground(ColorSuccess).
		Width(16)

	HelpDesc = lipgloss.NewStyle().
		Foreground(ColorText)

	// Vim mode indicator
	VimNormal = lipgloss.NewStyle().
			Foreground(ColorCyan).
			Bold(true)

	VimInsert = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)
)

// PanelTitle renders a numbered panel title like lazygit: "1 Connections (2)"
func PanelTitle(num int, name string, active bool) string {
	color := ColorMuted
	if num > 0 && num <= len(PanelColors) {
		color = PanelColors[num-1]
	}

	badge := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#282A36")).
		Background(color).
		Padding(0, 1).
		Render(string(rune('0' + num)))

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(color)
	if !active {
		titleStyle = titleStyle.Foreground(ColorMuted)
	}

	return badge + " " + titleStyle.Render(name)
}

func PanelStyle(active bool, width, height int) lipgloss.Style {
	borderColor := ColorBorder
	if active {
		borderColor = ColorActiveBorder
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width).
		Height(height)
}
