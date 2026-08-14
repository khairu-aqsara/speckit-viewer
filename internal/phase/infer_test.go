package phase

import (
	"testing"

	"github.com/wenkhairu/speckit-viewer/internal/model"
	"github.com/wenkhairu/speckit-viewer/internal/parser"
)

func kinds(ks ...model.FileKind) map[model.FileKind]bool {
	m := make(map[model.FileKind]bool)
	for _, k := range ks {
		m[k] = true
	}
	return m
}

func tasksDoc(checked, unchecked int) *parser.TasksDocument {
	p := parser.TaskPhase{Title: "Phase 1", Number: 1}
	for range checked {
		p.Tasks = append(p.Tasks, parser.Task{Done: true})
	}
	for range unchecked {
		p.Tasks = append(p.Tasks, parser.Task{})
	}
	return &parser.TasksDocument{Phases: []parser.TaskPhase{p}}
}

func TestInfer(t *testing.T) {
	cases := []struct {
		name      string
		present   map[model.FileKind]bool
		tasks     *parser.TasksDocument
		want      Phase
		wantLabel string
	}{
		{"no files", kinds(), nil, Unknown, "Unknown"},
		{"spec only", kinds(model.FileSpec), nil, Specified, "Specified"},
		{"spec+plan", kinds(model.FileSpec, model.FilePlan), nil, Planned, "Planned"},
		{"spec+research only", kinds(model.FileSpec, model.FileResearch), nil, Planned, "Planned"},
		{"spec+data-model only", kinds(model.FileSpec, model.FileDataModel), nil, Planned, "Planned"},
		{"tasks with zero tasks", kinds(model.FileSpec, model.FilePlan, model.FileTasks), tasksDoc(0, 0), TasksGenerated, "Tasks Generated"},
		{"tasks none checked", kinds(model.FileSpec, model.FilePlan, model.FileTasks), tasksDoc(0, 5), TasksGenerated, "Tasks Generated"},
		{"tasks partly checked", kinds(model.FileSpec, model.FilePlan, model.FileTasks), tasksDoc(72, 56), Implementing, "Implementing 72/128"},
		{"tasks all checked", kinds(model.FileSpec, model.FilePlan, model.FileTasks), tasksDoc(4, 0), Implemented, "Implemented"},
		{"tasks file present but unparsed", kinds(model.FileSpec, model.FileTasks), nil, Specified, "Specified"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Infer(c.present, c.tasks)
			if got.Phase != c.want {
				t.Errorf("Phase = %v, want %v", got.Phase, c.want)
			}
			if got.Label() != c.wantLabel {
				t.Errorf("Label() = %q, want %q", got.Label(), c.wantLabel)
			}
		})
	}
}
