package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

type appModel struct {
	width  int
	height int
}

func (m appModel) Init() tea.Cmd {
	return nil
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m appModel) View() tea.View {
	v := tea.NewView(fmt.Sprintf("speckit-viewer — %dx%d — press q to quit", m.width, m.height))
	v.AltScreen = true
	return v
}

func main() {
	if _, err := tea.NewProgram(appModel{}).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
