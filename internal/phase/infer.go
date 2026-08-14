// Package phase infers a feature's workflow phase. Spec Kit has no reliable
// per-feature status field, so phase comes from which files exist plus the
// tasks.md checkbox ratio.
package phase

import (
	"fmt"

	"github.com/wenkhairu/speckit-viewer/internal/model"
	"github.com/wenkhairu/speckit-viewer/internal/parser"
)

// Phase is a feature's inferred workflow phase.
type Phase int

const (
	Unknown Phase = iota
	Specified
	Planned
	TasksGenerated
	Implementing
	Implemented
)

// Result carries the phase plus task progress when tasks.md exists.
type Result struct {
	Phase   Phase
	Checked int
	Total   int
}

// Label renders the badge text, e.g. "Implementing 72/128".
func (r Result) Label() string {
	switch r.Phase {
	case Specified:
		return "Specified"
	case Planned:
		return "Planned"
	case TasksGenerated:
		return "Tasks Generated"
	case Implementing:
		return fmt.Sprintf("Implementing %d/%d", r.Checked, r.Total)
	case Implemented:
		return "Implemented"
	default:
		return "Unknown"
	}
}

// Infer applies the heuristic. tasks may be nil when tasks.md is absent.
func Infer(present map[model.FileKind]bool, tasks *parser.TasksDocument) Result {
	if present[model.FileTasks] && tasks != nil {
		checked, total := tasks.Checked(), tasks.Total()
		r := Result{Checked: checked, Total: total}
		switch {
		case total == 0 || checked == 0:
			r.Phase = TasksGenerated
		case checked == total:
			r.Phase = Implemented
		default:
			r.Phase = Implementing
		}
		return r
	}
	if present[model.FilePlan] || present[model.FileResearch] || present[model.FileDataModel] {
		return Result{Phase: Planned}
	}
	if present[model.FileSpec] {
		return Result{Phase: Specified}
	}
	return Result{Phase: Unknown}
}
