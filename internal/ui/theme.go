package ui

import (
	"charm.land/lipgloss/v2"

	"github.com/wenkhairu/speckit-viewer/internal/phase"
)

// theme holds the Lip Gloss styles for one background mode. Lip Gloss v2 has
// no adaptive colors, so the app rebuilds the theme when the terminal
// reports its background via tea.BackgroundColorMsg.
type theme struct {
	isDark bool

	paneBorder        lipgloss.Style
	paneBorderFocused lipgloss.Style
	statusBar         lipgloss.Style
	statusHints       lipgloss.Style
	phaseBadge        map[phase.Phase]lipgloss.Style
	checkDone         lipgloss.Style
	checkPending      lipgloss.Style
	checkPhaseHeader  lipgloss.Style
	badgeParallel     lipgloss.Style
	badgeStory        lipgloss.Style
}

func newTheme(isDark bool) theme {
	dim := lipgloss.Color("245")
	if !isDark {
		dim = lipgloss.Color("240")
	}
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(dim)
	focused := border.BorderForeground(lipgloss.Color("62"))

	badge := func(c string) lipgloss.Style {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(c))
	}
	return theme{
		isDark:            isDark,
		paneBorder:        border,
		paneBorderFocused: focused,
		statusBar:         lipgloss.NewStyle().Foreground(dim),
		statusHints:       lipgloss.NewStyle().Foreground(dim),
		phaseBadge: map[phase.Phase]lipgloss.Style{
			phase.Unknown:        badge("240"),
			phase.Specified:      badge("110"),
			phase.Planned:        badge("75"),
			phase.TasksGenerated: badge("179"),
			phase.Implementing:   badge("214"),
			phase.Implemented:    badge("78"),
		},
		checkDone:        badge("78"),
		checkPending:     lipgloss.NewStyle(),
		checkPhaseHeader: lipgloss.NewStyle().Bold(true),
		badgeParallel:    badge("135"),
		badgeStory:       badge("75"),
	}
}
