package ui

import (
	"path/filepath"
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

func key(s string) tea.KeyPressMsg {
	r := []rune(s)[0]
	return tea.KeyPressMsg{Code: r, Text: s}
}

func enter() tea.KeyPressMsg { return tea.KeyPressMsg{Code: tea.KeyEnter} }

// pump feeds msg into Update and then executes returned commands like the
// Bubble Tea runtime does, feeding resulting messages back until quiet.
// Commands run with a short deadline: instant ones (filter results) come
// back, timer-based ones (cursor blink) are dropped so pump terminates.
func pump(a *App, msg tea.Msg) {
	run := func(c tea.Cmd) tea.Msg {
		out := make(chan tea.Msg, 1)
		go func() { out <- c() }()
		select {
		case m := <-out:
			return m
		case <-time.After(50 * time.Millisecond):
			return nil
		}
	}
	queue := []tea.Msg{msg}
	for steps := 0; len(queue) > 0 && steps < 100; steps++ {
		m := queue[0]
		queue = queue[1:]
		if batch, ok := m.(tea.BatchMsg); ok {
			for _, c := range batch {
				if c != nil {
					if out := run(c); out != nil {
						queue = append(queue, out)
					}
				}
			}
			continue
		}
		_, cmd := a.Update(m)
		if cmd != nil {
			if out := run(cmd); out != nil {
				queue = append(queue, out)
			}
		}
	}
}

func TestFilterFlow(t *testing.T) {
	root, _ := filepath.Abs(filepath.Join("..", "..", "testdata", "full-project"))
	a, err := New(root, "test")
	if err != nil {
		t.Fatal(err)
	}
	pump(a, tea.WindowSizeMsg{Width: 120, Height: 40})

	// Enter on dashboard opens browse with nav focused.
	pump(a, enter())
	if a.screen != screenBrowse {
		t.Fatalf("screen = %v, want browse", a.screen)
	}

	// Tab to content, then "/" must start filtering from content focus.
	pump(a, tea.KeyPressMsg{Code: tea.KeyTab})
	if a.focus != focusContent {
		t.Fatalf("focus = %v, want content", a.focus)
	}
	pump(a, key("/"))
	if got := a.nav.FilterState(); got != list.Filtering {
		t.Fatalf("after /: FilterState = %v, want Filtering", got)
	}

	for _, ch := range "trivial" {
		pump(a, key(string(ch)))
	}
	visible := a.nav.VisibleItems()
	if len(visible) != 1 {
		for _, it := range visible {
			t.Logf("  visible: %q", it.(navEntry).title)
		}
		t.Fatalf("visible while filtering = %d, want 1", len(visible))
	}

	// Enter applies the filter; the match stays selected.
	pump(a, enter())
	if got := a.nav.FilterState(); got != list.FilterApplied {
		t.Fatalf("after apply: FilterState = %v, want FilterApplied", got)
	}
	sel, ok := a.nav.SelectedItem().(navEntry)
	if !ok || sel.kind != navFeature {
		t.Fatalf("selected after apply = %#v, want the trivial feature", a.nav.SelectedItem())
	}

	// Enter on the filtered feature expands it; sidebar must not be empty
	// and the screen must stay browse.
	pump(a, enter())
	if a.screen != screenBrowse {
		t.Fatalf("after open: screen = %v, want browse", a.screen)
	}
	items := a.nav.VisibleItems()
	if len(items) < 5 {
		t.Fatalf("sidebar has %d visible items after open, want full list with expansion", len(items))
	}
	sel2, _ := a.nav.SelectedItem().(navEntry)
	if sel2.kind != navFeature || sel2.featureIdx != sel.featureIdx {
		t.Fatalf("cursor after expand = %#v, want same feature", sel2)
	}
}
