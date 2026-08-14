// Package ui implements the Bubble Tea application: a dashboard table
// screen and a two-pane browse screen (navigator list + content viewport).
package ui

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/wenkhairu/speckit-viewer/internal/model"
	"github.com/wenkhairu/speckit-viewer/internal/parser"
	"github.com/wenkhairu/speckit-viewer/internal/phase"
	"github.com/wenkhairu/speckit-viewer/internal/scanner"
)

type screen int

const (
	screenDashboard screen = iota
	screenBrowse
)

type focusZone int

const (
	focusNav focusZone = iota
	focusContent
)

const sidebarWidth = 40

// featureVM is a Feature enriched with parsed tasks and inferred phase.
type featureVM struct {
	model.Feature
	Tasks       *parser.TasksDocument
	PhaseResult phase.Result
}

// App is the root Bubble Tea model.
type App struct {
	root     string
	version  string
	project  model.Project
	features []featureVM

	screen screen
	focus  focusZone
	theme  theme

	width  int
	height int

	nav      list.Model
	dash     table.Model
	vp       viewport.Model
	expanded map[int]bool

	// current content state
	currentPath  string // breadcrumb
	currentRaw   string // file content as read
	currentKind  model.FileKind
	currentTasks *parser.TasksDocument
	rawMode      bool // for tasks.md: true = Markdown, false = checklist
}

// New scans root and builds the app model. version is the build version
// shown in the status bar.
func New(root, version string) (*App, error) {
	a := &App{root: root, version: version, theme: newTheme(true), expanded: map[int]bool{}}
	if err := a.rescan(); err != nil {
		return nil, err
	}
	a.initWidgets()
	return a, nil
}

func (a *App) rescan() error {
	project, err := scanner.Scan(a.root)
	if err != nil {
		return err
	}
	a.project = project
	a.features = a.features[:0]
	for _, f := range project.Features {
		vm := featureVM{Feature: f}
		if tf := f.File(model.FileTasks); tf != nil {
			if content, err := os.ReadFile(tf.Abs); err == nil {
				doc := parser.ParseTasks(string(content))
				vm.Tasks = &doc
			}
		}
		vm.PhaseResult = phase.Infer(f.PresentKinds(), vm.Tasks)
		a.features = append(a.features, vm)
	}
	return nil
}

func (a *App) initWidgets() {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)

	a.nav = list.New(buildNavEntries(a.features, a.project.ConstitutionPath != "", a.expanded), delegate, 0, 0)
	a.nav.SetShowTitle(false)
	a.nav.SetShowStatusBar(false)
	a.nav.SetShowHelp(false)
	a.nav.DisableQuitKeybindings()

	a.dash = table.New(
		table.WithColumns(dashboardColumns(80)),
		table.WithRows(dashboardRows(a.features)),
		table.WithFocused(true),
	)
	a.vp = viewport.New()
}

func (a *App) Init() tea.Cmd {
	return tea.RequestBackgroundColor
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		a.theme = newTheme(msg.IsDark())
		a.refreshContent()
		return a, nil

	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		a.layout()
		a.refreshContent()
		return a, nil

	case tea.KeyPressMsg:
		return a.handleKey(msg)

	case tea.MouseWheelMsg:
		return a.handleWheel(msg)
	}

	// Non-key messages (list filter results, cursor blinks) belong to the
	// nav list: its filtering runs asynchronously and delivers matches via
	// its own message type, which must reach the widget or / never narrows.
	var cmd tea.Cmd
	a.nav, cmd = a.nav.Update(msg)
	return a, cmd
}

// handleWheel routes mouse-wheel events by screen and pointer position.
// Only the viewport handles the wheel natively; the list and table get the
// wheel translated to cursor moves.
func (a *App) handleWheel(msg tea.MouseWheelMsg) (tea.Model, tea.Cmd) {
	up := msg.Button == tea.MouseWheelUp
	if a.screen == screenDashboard {
		if up {
			a.dash.MoveUp(3)
		} else {
			a.dash.MoveDown(3)
		}
		return a, nil
	}
	if msg.X < sidebarWidth {
		if up {
			a.nav.CursorUp()
		} else {
			a.nav.CursorDown()
		}
		return a, nil
	}
	var cmd tea.Cmd
	a.vp, cmd = a.vp.Update(msg)
	return a, cmd
}

func (a *App) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// While the list filter is being typed, every key belongs to the list.
	if a.screen == screenBrowse && a.nav.FilterState() == list.Filtering {
		var cmd tea.Cmd
		a.nav, cmd = a.nav.Update(msg)
		return a, cmd
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return a, tea.Quit
	case "d":
		a.screen = screenDashboard
		return a, nil
	case "c":
		if a.project.ConstitutionPath != "" {
			a.screen = screenBrowse
			a.focus = focusContent
			a.openPath(a.project.ConstitutionPath, model.FileOther, nil, ".specify/memory/constitution.md")
		}
		return a, nil
	case "R":
		if err := a.rescan(); err == nil {
			return a, a.syncAfterRescan()
		}
		return a, nil
	case "tab":
		if a.screen == screenBrowse {
			if a.focus == focusNav {
				a.focus = focusContent
			} else {
				a.focus = focusNav
			}
		}
		return a, nil
	case "/":
		// Start filtering from anywhere in the browse screen, even when
		// the content pane has focus.
		if a.screen == screenBrowse {
			a.focus = focusNav
			var cmd tea.Cmd
			a.nav, cmd = a.nav.Update(msg)
			return a, cmd
		}
		return a, nil
	case "r":
		if a.screen == screenBrowse && a.currentKind == model.FileTasks {
			a.rawMode = !a.rawMode
			a.refreshContent()
		}
		return a, nil
	case "enter":
		return a, a.handleEnter()
	}

	var cmd tea.Cmd
	switch {
	case a.screen == screenDashboard:
		a.dash, cmd = a.dash.Update(msg)
	case a.focus == focusNav:
		a.nav, cmd = a.nav.Update(msg)
	default:
		a.vp, cmd = a.vp.Update(msg)
	}
	return a, cmd
}

func (a *App) handleEnter() tea.Cmd {
	if a.screen == screenDashboard {
		if cursor := a.dash.Cursor(); cursor >= 0 && cursor < len(a.features) {
			a.screen = screenBrowse
			a.focus = focusNav
			a.openFeatureSpec(cursor)
			return a.selectNavFeature(cursor)
		}
		return nil
	}
	entry, ok := a.nav.SelectedItem().(navEntry)
	if !ok {
		return nil
	}
	// Choosing a result ends the search: a stale filter would otherwise
	// hide the rows that a rebuild inserts (e.g. expanded files).
	if a.nav.FilterState() != list.Unfiltered {
		a.nav.ResetFilter()
	}
	switch entry.kind {
	case navDashboard:
		a.screen = screenDashboard
	case navConstitution:
		a.openPath(a.project.ConstitutionPath, model.FileOther, nil, ".specify/memory/constitution.md")
	case navFeature:
		a.expanded[entry.featureIdx] = !a.expanded[entry.featureIdx]
		return a.rebuildNav(entry.key())
	case navFile:
		f := a.features[entry.featureIdx]
		file := f.Files[entry.fileIdx]
		var tasks *parser.TasksDocument
		if file.Kind == model.FileTasks {
			tasks = f.Tasks
		}
		a.openPath(file.Abs, file.Kind, tasks, fmt.Sprintf("specs/%s-%s/%s", f.ID, f.Slug, file.Rel))
	}
	return nil
}

// rebuildNav regenerates the flat list and restores the cursor to the entry
// with the given key — the one piece of state the widget does not manage.
// SetItems returns a command that re-runs an active filter; dropping it
// leaves the filtered view empty, so it must reach the program loop.
func (a *App) rebuildNav(selectKey string) tea.Cmd {
	items := buildNavEntries(a.features, a.project.ConstitutionPath != "", a.expanded)
	cmd := a.nav.SetItems(items)
	for i, item := range items {
		if e, ok := item.(navEntry); ok && e.key() == selectKey {
			a.nav.Select(i)
			break
		}
	}
	return cmd
}

func (a *App) selectNavFeature(featureIdx int) tea.Cmd {
	return a.rebuildNav(navEntry{kind: navFeature, featureIdx: featureIdx}.key())
}

func (a *App) syncAfterRescan() tea.Cmd {
	selected := ""
	if e, ok := a.nav.SelectedItem().(navEntry); ok {
		selected = e.key()
	}
	cmd := a.rebuildNav(selected)
	a.dash.SetRows(dashboardRows(a.features))
	if a.currentPath != "" {
		// Re-read the open file so R also refreshes the content pane.
		a.reloadCurrent()
	}
	return cmd
}

func (a *App) openFeatureSpec(featureIdx int) {
	f := a.features[featureIdx]
	if spec := f.File(model.FileSpec); spec != nil {
		a.openPath(spec.Abs, model.FileSpec, nil, fmt.Sprintf("specs/%s-%s/spec.md", f.ID, f.Slug))
	}
}

func (a *App) openPath(abs string, kind model.FileKind, tasks *parser.TasksDocument, crumb string) {
	content, err := os.ReadFile(abs)
	if err != nil {
		a.currentRaw = fmt.Sprintf("Cannot read %s: %v", abs, err)
		a.currentKind = model.FileOther
		a.currentTasks = nil
	} else {
		a.currentRaw = string(content)
		a.currentKind = kind
		a.currentTasks = tasks
	}
	a.currentPath = crumb
	a.rawMode = false
	a.refreshContent()
	a.vp.GotoTop()
}

func (a *App) reloadCurrent() {
	if content, err := os.ReadFile(a.absOfCurrent()); err == nil {
		a.currentRaw = string(content)
		if a.currentKind == model.FileTasks {
			doc := parser.ParseTasks(a.currentRaw)
			a.currentTasks = &doc
		}
		a.refreshContent()
	}
}

func (a *App) absOfCurrent() string {
	if a.currentPath == ".specify/memory/constitution.md" {
		return a.project.ConstitutionPath
	}
	return a.root + "/" + a.currentPath
}

// refreshContent re-renders the viewport content. Glamour wrap width is
// render-time state, so this runs on resize and theme change too.
func (a *App) refreshContent() {
	if a.currentPath == "" {
		a.vp.SetContent("Select a file from the navigator.\n\nEnter expands a feature; Enter on a file opens it here.")
		return
	}
	contentWidth := a.contentPaneWidth() - 2
	if a.currentKind == model.FileTasks && !a.rawMode && a.currentTasks != nil {
		a.vp.SetContent(renderChecklist(*a.currentTasks, contentWidth, a.theme))
		return
	}
	a.vp.SetContent(renderMarkdown(a.currentRaw, contentWidth, a.theme.isDark))
}

func (a *App) contentPaneWidth() int {
	return max(20, a.width-sidebarWidth-4)
}

func (a *App) layout() {
	// pane borders + status bar, plus the banner when the window fits it
	innerHeight := max(4, a.height-3-bannerHeight(a.width, a.height))
	a.nav.SetSize(sidebarWidth-2, innerHeight)
	a.dash.SetColumns(dashboardColumns(a.width))
	a.dash.SetWidth(a.width - 2)
	a.dash.SetHeight(innerHeight)
	a.vp.SetWidth(a.contentPaneWidth())
	a.vp.SetHeight(innerHeight)
}

func (a *App) View() tea.View {
	var body string
	switch a.screen {
	case screenDashboard:
		body = a.theme.paneBorderFocused.Render(a.dash.View())
	default:
		navStyle, contentStyle := a.theme.paneBorderFocused, a.theme.paneBorder
		if a.focus == focusContent {
			navStyle, contentStyle = a.theme.paneBorder, a.theme.paneBorderFocused
		}
		body = lipgloss.JoinHorizontal(
			lipgloss.Top,
			navStyle.Render(a.nav.View()),
			contentStyle.Render(a.vp.View()),
		)
	}
	if showBanner(a.width, a.height) {
		body = banner(a.width, a.theme, a.version) + "\n" + body
	}
	v := tea.NewView(body + "\n" + a.statusBar())
	v.AltScreen = true
	v.WindowTitle = "speckit-viewer"
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (a *App) statusBar() string {
	crumb := a.currentPath
	if a.screen == screenDashboard {
		crumb = fmt.Sprintf("%d features", len(a.features))
	}
	version := "speckit " + a.version
	left := a.theme.statusBar.Render(" " + crumb)
	// Drop hint groups from the right until the bar fits the width; the
	// version stays visible even on narrow terminals.
	hintGroups := []string{
		"↑↓ nav", "Enter open", "Tab switch pane", "/ filter", "r raw",
		"R refresh", "d dashboard", "c constitution", "q quit",
	}
	for n := len(hintGroups); n >= 0; n-- {
		parts := append(append([]string{}, hintGroups[:n]...), version)
		right := a.theme.statusHints.Render(strings.Join(parts, " · ") + " ")
		gap := a.width - lipgloss.Width(left) - lipgloss.Width(right)
		if gap >= 1 || n == 0 {
			return left + lipgloss.NewStyle().Width(max(1, gap)).Render("") + right
		}
	}
	return left
}
