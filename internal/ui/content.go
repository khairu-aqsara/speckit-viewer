package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/progress"
	"charm.land/glamour/v2"

	"github.com/wenkhairu/speckit-viewer/internal/parser"
)

// renderMarkdown renders Markdown for the given pane width and background
// mode. Wrap width is render-time state, so callers must re-render on resize.
func renderMarkdown(content string, width int, isDark bool) string {
	style := "light"
	if isDark {
		style = "dark"
	}
	if width < 10 {
		width = 10
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return content
	}
	out, err := r.Render(content)
	if err != nil {
		return content
	}
	return out
}

// renderChecklist renders a parsed tasks.md as the special checklist view:
// phase headers with a progress bar, task rows with checkbox and badges.
func renderChecklist(doc parser.TasksDocument, width int, th theme) string {
	var b strings.Builder
	barWidth := min(24, max(8, width-30))
	bar := progress.New(progress.WithDefaultBlend())
	bar.SetWidth(barWidth)
	bar.ShowPercentage = false

	for _, p := range doc.Phases {
		checked, total := 0, len(p.Tasks)
		for _, t := range p.Tasks {
			if t.Done {
				checked++
			}
		}
		title := p.Title
		if title == "" {
			title = "Tasks"
		}
		ratio := 0.0
		if total > 0 {
			ratio = float64(checked) / float64(total)
		}
		b.WriteString(th.checkPhaseHeader.Render(title))
		b.WriteString(fmt.Sprintf("\n%s %d/%d\n\n", bar.ViewAs(ratio), checked, total))

		for _, t := range p.Tasks {
			mark, style := "░", th.checkPending
			if t.Done {
				mark, style = "✓", th.checkDone
			}
			line := fmt.Sprintf("%s %s", mark, t.ID)
			if t.Parallel {
				line += " " + th.badgeParallel.Render("[P]")
			}
			if t.Story != "" {
				line += " " + th.badgeStory.Render("["+t.Story+"]")
			}
			b.WriteString("  " + style.Render(line+" "+t.Text) + "\n")
		}
		b.WriteString("\n")
	}
	if len(doc.Phases) == 0 {
		b.WriteString("No tasks found in tasks.md\n")
	}
	return b.String()
}
