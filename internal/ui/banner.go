package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// bannerLines spells SPECKIT in a 2-row half-block mini font.
var bannerLines = []string{
	"▄▀▀ █▀█ █▀▀ ▄▀▀ █▄▀ █ ▀█▀",
	"▄▄▀ █▀▀ █▄▄ ▀▄▄ █ █ █  █",
}

// bannerColors is a per-line slice of the Spec Kit CLI gradient.
var bannerColors = []string{"12", "14"}

const bannerMinWidth = 30
const bannerMinHeight = 14

// showBanner reports whether the terminal is large enough for the banner.
// Small windows keep every row for content.
func showBanner(width, height int) bool {
	return width >= bannerMinWidth && height >= bannerMinHeight
}

// bannerHeight is the row count the banner occupies, including the tagline.
func bannerHeight(width, height int) int {
	if !showBanner(width, height) {
		return 0
	}
	return len(bannerLines) + 1
}

// banner renders the compact art top-left with the tagline and version
// underneath, e.g. "GitHub Spec Kit viewer v1.0.0".
func banner(width int, th theme, version string) string {
	styled := make([]string, len(bannerLines))
	for i, line := range bannerLines {
		color := bannerColors[i%len(bannerColors)]
		styled[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(line)
	}
	tagline := lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("11")).
		Render("GitHub Spec Kit viewer")
	ver := lipgloss.NewStyle().Foreground(lipgloss.Color("14")).
		Render(displayVersion(version))
	return strings.Join(styled, "\n") + "\n" + tagline + " " + ver
}

// displayVersion prefixes release versions with "v"; "dev" stays as-is.
func displayVersion(version string) string {
	if version == "dev" || strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}
