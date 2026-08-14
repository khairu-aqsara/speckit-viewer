package ui

import (
	"fmt"

	"charm.land/bubbles/v2/list"

	"github.com/wenkhairu/speckit-viewer/internal/model"
)

type navKind int

const (
	navDashboard navKind = iota
	navConstitution
	navFeature
	navFile
)

// navEntry is one row of the navigator. The navigator is a flat list, not a
// tree: expanding a feature inserts indented file rows under it. Expansion
// state lives in the app model, and the list is rebuilt on every toggle.
type navEntry struct {
	kind       navKind
	featureIdx int // index into project.Features for navFeature and navFile
	fileIdx    int // index into Feature.Files for navFile
	title      string
	filter     string
}

func (e navEntry) Title() string       { return e.title }
func (e navEntry) Description() string { return "" }
func (e navEntry) FilterValue() string { return e.filter }

// key identifies an entry across list rebuilds so the cursor can be restored.
func (e navEntry) key() string {
	return fmt.Sprintf("%d:%d:%d", e.kind, e.featureIdx, e.fileIdx)
}

// buildNavEntries produces the flat navigator rows from the scanned project
// plus the expansion set.
func buildNavEntries(features []featureVM, hasConstitution bool, expanded map[int]bool) []list.Item {
	items := []list.Item{
		navEntry{kind: navDashboard, featureIdx: -1, title: "⌂ Dashboard", filter: "dashboard"},
	}
	if hasConstitution {
		items = append(items, navEntry{
			kind: navConstitution, featureIdx: -1, title: "§ Constitution", filter: "constitution",
		})
	}
	for i, f := range features {
		arrow := "▸"
		if expanded[i] {
			arrow = "▾"
		}
		items = append(items, navEntry{
			kind:       navFeature,
			featureIdx: i,
			title:      fmt.Sprintf("%s %s · %s  %s", arrow, f.ID, f.Name, f.PhaseResult.Label()),
			filter:     fmt.Sprintf("%s %s %s", f.ID, f.Slug, f.Name),
		})
		if !expanded[i] {
			continue
		}
		for j, file := range f.Files {
			items = append(items, navEntry{
				kind:       navFile,
				featureIdx: i,
				fileIdx:    j,
				title:      "    " + file.Rel,
				filter:     fmt.Sprintf("%s %s", f.ID, file.Rel),
			})
		}
	}
	return items
}

// fileLabel names a file for the breadcrumb.
func fileLabel(f model.SpecKitFile) string { return f.Rel }
