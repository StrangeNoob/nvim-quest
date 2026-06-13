package ui

import "github.com/charmbracelet/lipgloss"

type Palette struct {
	Primary lipgloss.Color
	Accent  lipgloss.Color
}

// Per-act palettes: dojo greens, crypt embers, neon grid.
var palettes = map[int]Palette{
	1: {Primary: lipgloss.Color("114"), Accent: lipgloss.Color("230")},
	2: {Primary: lipgloss.Color("214"), Accent: lipgloss.Color("203")},
	3: {Primary: lipgloss.Color("213"), Accent: lipgloss.Color("51")},
}

func paletteFor(act int) Palette {
	if p, ok := palettes[act]; ok {
		return p
	}
	return palettes[1]
}

// PrimaryStyle is a lipgloss style foregrounded with the palette's primary color.
func (p Palette) PrimaryStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(p.Primary)
}

func actName(act int) string {
	switch act {
	case 1:
		return "THE CURSOR DOJO"
	case 2:
		return "THE MOTION CRYPTS"
	case 3:
		return "THE NEON GRID"
	}
	return ""
}

// actSummary is the short "what you'll learn here" line shown when an act unlocks.
func actSummary(act int) string {
	switch act {
	case 1:
		return "h/j/k/l · w/b · 0/$ · gg/G"
	case 2:
		return "i/a · x · dw · dd · yy/p · cw · cc"
	case 3:
		return "/ search · n · counts (4w, 3dd)"
	}
	return ""
}

var (
	cursorStyle  = lipgloss.NewStyle().Reverse(true)
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	dangerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
)
