package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// bannerLines spells SPECKIT in the ANSI Shadow figlet font — the same font
// the official GitHub Spec Kit CLI uses for its SPECIFY banner.
var bannerLines = []string{
	"███████╗██████╗ ███████╗ ██████╗██╗  ██╗██╗████████╗",
	"██╔════╝██╔══██╗██╔════╝██╔════╝██║ ██╔╝██║╚══██╔══╝",
	"███████╗██████╔╝█████╗  ██║     █████╔╝ ██║   ██║",
	"╚════██║██╔═══╝ ██╔══╝  ██║     ██╔═██╗ ██║   ██║",
	"███████║██║     ███████╗╚██████╗██║  ██╗██║   ██║",
	"╚══════╝╚═╝     ╚══════╝ ╚═════╝╚═╝  ╚═╝╚═╝   ╚═╝",
}

// bannerColors is the per-line gradient the Spec Kit CLI applies:
// bright_blue, blue, cyan, bright_cyan, white, bright_white.
var bannerColors = []string{"12", "4", "6", "14", "7", "15"}

const bannerMinWidth = 54
const bannerMinHeight = 24

// showBanner reports whether the terminal is large enough for the banner.
// Small windows keep every row for content.
func showBanner(width, height int) bool {
	return width >= bannerMinWidth && height >= bannerMinHeight
}

// bannerHeight is the row count the banner occupies, including its tagline.
func bannerHeight(width, height int) int {
	if !showBanner(width, height) {
		return 0
	}
	return len(bannerLines) + 1
}

// banner renders the ASCII art centered for the terminal width.
func banner(width int, th theme) string {
	styled := make([]string, len(bannerLines))
	for i, line := range bannerLines {
		color := bannerColors[i%len(bannerColors)]
		styled[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(line)
	}
	tagline := lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("11")).
		Render("GitHub Spec Kit viewer")
	block := lipgloss.JoinVertical(lipgloss.Center, strings.Join(styled, "\n"), tagline)
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, block)
}
