package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	ColorPrimary = lipgloss.AdaptiveColor{Light: "#7b2cbf", Dark: "#c77dff"}
	ColorAccent  = lipgloss.AdaptiveColor{Light: "#e76f51", Dark: "#f4a261"}
	ColorSuccess = lipgloss.AdaptiveColor{Light: "#2a9d8f", Dark: "#64d8cb"}
	ColorWarn    = lipgloss.AdaptiveColor{Light: "#e9c46a", Dark: "#ffd60a"}
	ColorErr     = lipgloss.AdaptiveColor{Light: "#d62828", Dark: "#ff4d6d"}
	ColorDim     = lipgloss.AdaptiveColor{Light: "#4a4e69", Dark: "#7c7a85"}
	ColorBorder  = lipgloss.AdaptiveColor{Light: "#b8b8b8", Dark: "#3a3a3a"}
)

type Theme struct {
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Badge       lipgloss.Style
	BadgeOK     lipgloss.Style
	BadgeWarn   lipgloss.Style
	BadgeErr    lipgloss.Style
	Success     lipgloss.Style
	Warn        lipgloss.Style
	ErrStyle    lipgloss.Style
	Dim         lipgloss.Style
	Panel       lipgloss.Style
	Selected    lipgloss.Style
	TableHeader lipgloss.Style
	Hotkey      lipgloss.Style
	Help        lipgloss.Style
}

func NewTheme() *Theme {
	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 1)

	return &Theme{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary),
		Subtitle: lipgloss.NewStyle().
			Foreground(ColorDim).
			Italic(true),
		Badge: lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true),
		BadgeOK: lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true),
		BadgeWarn: lipgloss.NewStyle().
			Foreground(ColorWarn).
			Bold(true),
		BadgeErr: lipgloss.NewStyle().
			Foreground(ColorErr).
			Bold(true),
		Success: lipgloss.NewStyle().Foreground(ColorSuccess),
		Warn:    lipgloss.NewStyle().Foreground(ColorWarn),
		ErrStyle: lipgloss.NewStyle().
			Foreground(ColorErr).
			Bold(true),
		Dim:   lipgloss.NewStyle().Foreground(ColorDim),
		Panel: panel,
		Selected: lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true),
		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Underline(true),
		Hotkey: lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true),
		Help: lipgloss.NewStyle().
			Foreground(ColorDim),
	}
}
