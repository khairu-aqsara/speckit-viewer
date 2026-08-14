// Command speckit is a two-pane terminal dashboard for GitHub Spec Kit
// projects: it scans a project's specs/ folder, infers each feature's
// workflow phase, and renders the Markdown artifacts.
//
// Usage: speckit [path]   (path defaults to the current directory)
package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/wenkhairu/speckit-viewer/internal/ui"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	root, err := filepath.Abs(root)
	if err != nil {
		fatal(err)
	}

	app, err := ui.New(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "speckit: %v\n", err)
		fmt.Fprintln(os.Stderr, "Point speckit at a Spec Kit project root (a folder containing specs/).")
		os.Exit(1)
	}
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "speckit:", err)
	os.Exit(1)
}
