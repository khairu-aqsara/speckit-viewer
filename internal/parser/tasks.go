// Package parser extracts structure from Spec Kit Markdown files.
package parser

import (
	"regexp"
	"strconv"
	"strings"
)

// Task is one checklist line in tasks.md.
type Task struct {
	ID       string // "T031"
	Done     bool
	Parallel bool   // [P] marker
	Story    string // "US1" or ""
	Text     string
}

// TaskPhase is one "## ..." section of tasks.md that contains tasks or
// matches the "Phase N" header pattern. Phase numbers can be non-contiguous.
type TaskPhase struct {
	Title  string // header text without the leading "## "
	Number int    // parsed N, or 0 when the header is not "Phase N" shaped
	MVP    bool   // header carries the "🎯 MVP" suffix
	Tasks  []Task
}

// TasksDocument is the parsed shape of a tasks.md file.
type TasksDocument struct {
	Phases []TaskPhase
}

// Total counts all tasks.
func (d TasksDocument) Total() int {
	n := 0
	for _, p := range d.Phases {
		n += len(p.Tasks)
	}
	return n
}

// Checked counts completed tasks.
func (d TasksDocument) Checked() int {
	n := 0
	for _, p := range d.Phases {
		for _, t := range p.Tasks {
			if t.Done {
				n++
			}
		}
	}
	return n
}

var (
	// Checkbox case is inconsistent in real projects: some features use
	// "[x]", others "[X]".
	taskLineRe    = regexp.MustCompile(`^\s*- \[([ xX])\]\s+(T\d{3})\s+(.*)$`)
	parallelRe    = regexp.MustCompile(`^\[P\]\s*`)
	storyRe       = regexp.MustCompile(`^\[(US\d+)\]\s*`)
	phaseHeaderRe = regexp.MustCompile(`^##\s+Phase\s+(\d+)\s*:?\s*(.*)$`)
	anyHeaderRe   = regexp.MustCompile(`^##\s+(.*)$`)
)

// ParseTasks parses tasks.md content with a line-based state machine.
// Sections are kept when they match the "Phase N" pattern or contain at
// least one task line; other sections (Dependencies, notes) are dropped.
// Tasks appearing before any header land in an untitled phase.
func ParseTasks(content string) TasksDocument {
	var doc TasksDocument
	var current *TaskPhase

	flush := func() {
		if current != nil && (len(current.Tasks) > 0 || current.Number > 0) {
			doc.Phases = append(doc.Phases, *current)
		}
		current = nil
	}

	for _, line := range strings.Split(content, "\n") {
		if m := phaseHeaderRe.FindStringSubmatch(line); m != nil {
			flush()
			num, _ := strconv.Atoi(m[1])
			title := strings.TrimSpace(strings.TrimPrefix(line, "##"))
			mvp := strings.Contains(title, "🎯 MVP")
			current = &TaskPhase{Title: title, Number: num, MVP: mvp}
			continue
		}
		if m := anyHeaderRe.FindStringSubmatch(line); m != nil {
			flush()
			current = &TaskPhase{Title: strings.TrimSpace(m[1])}
			continue
		}
		m := taskLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		task := Task{
			ID:   m[2],
			Done: m[1] == "x" || m[1] == "X",
		}
		rest := m[3]
		if loc := parallelRe.FindString(rest); loc != "" {
			task.Parallel = true
			rest = rest[len(loc):]
		}
		if sm := storyRe.FindStringSubmatch(rest); sm != nil {
			task.Story = sm[1]
			rest = rest[len(sm[0]):]
		}
		task.Text = strings.TrimSpace(rest)
		if current == nil {
			current = &TaskPhase{}
		}
		current.Tasks = append(current.Tasks, task)
	}
	flush()
	return doc
}
