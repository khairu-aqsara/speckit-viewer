package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

var bannerLines = []string{
	"███████ ██████  ███████  ██████ ██   ██ ██ ████████",
	"██      ██   ██ ██      ██      ██  ██  ██    ██",
	"███████ ██████  █████   ██      █████   ██    ██",
	"     ██ ██      ██      ██      ██  ██  ██    ██",
	"███████ ██      ███████  ██████ ██   ██ ██    ██",
}

const bannerMinWidth = 54
const bannerMinHeight = 24

// showBanner reports whether the terminal is large enough for the banner.
// Small windows keep every row for content.
func showBanner(width, height int) bool {
	return width >= bannerMinWidth && height >= bannerMinHeight
}

// bannerHeight is the row count the banner occupies, including its subtitle.
func bannerHeight(width, height int) int {
	if !showBanner(width, height) {
		return 0
	}
	return len(bannerLines) + 1
}

// banner renders the ASCII art centered for the terminal width.
func banner(width int, th theme) string {
	art := lipgloss.NewStyle().Foreground(lipgloss.Color("62")).
		Render(strings.Join(bannerLines, "\n"))
	subtitle := th.statusHints.Render("GitHub Spec Kit viewer")
	block := lipgloss.JoinVertical(lipgloss.Center, art, subtitle)
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, block)
}
