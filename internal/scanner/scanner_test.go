package scanner

import (
	"path/filepath"
	"testing"

	"github.com/wenkhairu/speckit-viewer/internal/model"
)

func TestClassifyFile(t *testing.T) {
	cases := []struct {
		rel  string
		want model.FileKind
	}{
		{"spec.md", model.FileSpec},
		{"plan.md", model.FilePlan},
		{"tasks.md", model.FileTasks},
		{"research.md", model.FileResearch},
		{"data-model.md", model.FileDataModel},
		{"quickstart.md", model.FileQuickstart},
		{"checklists/requirements.md", model.FileChecklist},
		{"contracts/payslip.md", model.FileContract},
		{"contracts/deep/nested.md", model.FileOther},
		{"notes.md", model.FileOther},
		{"diagram.png", model.FileOther},
	}
	for _, c := range cases {
		if got := ClassifyFile(c.rel); got != c.want {
			t.Errorf("ClassifyFile(%q) = %v, want %v", c.rel, got, c.want)
		}
	}
}

func TestBuildFeature(t *testing.T) {
	f, ok := BuildFeature("031-employee-payroll", []string{"tasks.md", "spec.md", "plan.md"})
	if !ok {
		t.Fatal("BuildFeature rejected a valid feature folder")
	}
	if f.ID != "031" || f.Slug != "employee-payroll" || f.Name != "Employee Payroll" {
		t.Errorf("feature = %+v", f)
	}
	if len(f.Files) != 3 || f.Files[0].Rel != "plan.md" {
		t.Errorf("files must be sorted, got %+v", f.Files)
	}
	if f.File(model.FileTasks) == nil {
		t.Error("File(FileTasks) = nil")
	}
	if !f.PresentKinds()[model.FileSpec] {
		t.Error("PresentKinds missing FileSpec")
	}

	if _, ok := BuildFeature("not-a-feature", nil); ok {
		t.Error("folder without numeric prefix must be rejected")
	}
	if _, ok := BuildFeature("templates", nil); ok {
		t.Error("plain folder name must be rejected")
	}
}

func TestScan(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "full-project")
	project, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if project.ConstitutionPath == "" {
		t.Error("constitution not found")
	}
	if len(project.Features) != 2 {
		t.Fatalf("features = %d, want 2", len(project.Features))
	}
	payroll := project.Features[0]
	if payroll.ID != "003" || len(payroll.Files) != 8 {
		t.Errorf("payroll = %s with %d files, want 003 with 8", payroll.ID, len(payroll.Files))
	}
	for _, file := range payroll.Files {
		if file.Abs == "" {
			t.Errorf("file %s has no absolute path", file.Rel)
		}
	}
	trivial := project.Features[1]
	if trivial.ID != "007" || len(trivial.Files) != 1 || trivial.Files[0].Kind != model.FileSpec {
		t.Errorf("trivial = %+v, want only spec.md", trivial)
	}

	if _, err := Scan(filepath.Join("..", "..", "testdata", "nope")); err == nil {
		t.Error("Scan of a non-project must fail")
	}
}
