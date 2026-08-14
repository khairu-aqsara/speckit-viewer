package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// bannerLines spells SPECKIT in the compact Calvin S figlet font — the same
// box-drawing aesthetic as the official Spec Kit CLI banner, at 3 rows.
var bannerLines = []string{
	"╔═╗╔═╗╔═╗╔═╗╦╔═╦╔╦╗",
	"╚═╗╠═╝║╣ ║  ╠╩╗║ ║",
	"╚═╝╩  ╚═╝╚═╝╩ ╩╩ ╩",
}

// bannerColors is a per-line slice of the Spec Kit CLI gradient:
// bright_blue, bright_cyan, bright_white.
var bannerColors = []string{"12", "14", "15"}

const bannerMinWidth = 46
const bannerMinHeight = 16

// showBanner reports whether the terminal is large enough for the banner.
// Small windows keep every row for content.
func showBanner(width, height int) bool {
	return width >= bannerMinWidth && height >= bannerMinHeight
}

// bannerHeight is the row count the banner occupies.
func bannerHeight(width, height int) int {
	if !showBanner(width, height) {
		return 0
	}
	return len(bannerLines)
}

// banner renders the compact art top-left, with the tagline beside it.
func banner(width int, th theme) string {
	styled := make([]string, len(bannerLines))
	for i, line := range bannerLines {
		color := bannerColors[i%len(bannerColors)]
		styled[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(line)
	}
	art := strings.Join(styled, "\n")
	tagline := lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("11")).
		Render("GitHub Spec Kit viewer")
	return lipgloss.JoinHorizontal(lipgloss.Center, art, "  "+tagline)
}
