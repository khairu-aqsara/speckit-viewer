package parser

import "testing"

// golden mirrors the tasks.md shape captured from a production Spec Kit
// project (terus-crm): mixed-case checkboxes, [P] and [USn] badges, an
// MVP-suffixed phase, a non-phase section, and an appended out-of-order
// phase number.
const golden = `# Tasks: Employee Payroll

**Input**: Design documents from specs/031-employee-payroll/

## Phase 1: Setup

- [x] T001 Create module skeleton in local/payroll
- [X] T002 [P] Configure lint rules

## Phase 3: User Story 1 - Generate payslip 🎯 MVP

- [x] T003 [P] [US1] Add payslip table schema
- [ ] T004 [US1] Implement payslip renderer
- [ ] T005 Wire cron task

## Dependencies

- User Story 1 depends on Setup.

## Phase 2: Foundational

- [ ] T006 [P] Add capability definitions
`

func TestParseTasksGolden(t *testing.T) {
	doc := ParseTasks(golden)

	if got := len(doc.Phases); got != 3 {
		t.Fatalf("phases = %d, want 3 (Dependencies section must be dropped)", got)
	}
	if doc.Total() != 6 {
		t.Errorf("Total() = %d, want 6", doc.Total())
	}
	if doc.Checked() != 3 {
		t.Errorf("Checked() = %d, want 3 (must count both [x] and [X])", doc.Checked())
	}

	p1, p3, p2 := doc.Phases[0], doc.Phases[1], doc.Phases[2]
	if p1.Number != 1 || p3.Number != 3 || p2.Number != 2 {
		t.Errorf("phase numbers = %d,%d,%d, want file order 1,3,2", p1.Number, p3.Number, p2.Number)
	}
	if !p3.MVP {
		t.Error("phase 3 must carry the MVP flag")
	}
	if p1.MVP || p2.MVP {
		t.Error("only phase 3 is MVP")
	}

	t003 := p3.Tasks[0]
	if t003.ID != "T003" || !t003.Done || !t003.Parallel || t003.Story != "US1" {
		t.Errorf("T003 parsed as %+v, want done parallel US1", t003)
	}
	if t003.Text != "Add payslip table schema" {
		t.Errorf("T003 text = %q, badges must be stripped", t003.Text)
	}

	t004 := p3.Tasks[1]
	if t004.Done || t004.Parallel || t004.Story != "US1" {
		t.Errorf("T004 parsed as %+v, want undone non-parallel US1", t004)
	}
	t005 := p3.Tasks[2]
	if t005.Parallel || t005.Story != "" {
		t.Errorf("T005 parsed as %+v, want no badges", t005)
	}
}

func TestParseTasksEdgeCases(t *testing.T) {
	t.Run("empty content", func(t *testing.T) {
		doc := ParseTasks("")
		if doc.Total() != 0 || len(doc.Phases) != 0 {
			t.Errorf("empty file must parse to zero phases, got %+v", doc)
		}
	})

	t.Run("zero-task phase is kept", func(t *testing.T) {
		doc := ParseTasks("## Phase 4: Polish\n\nNothing yet.\n")
		if len(doc.Phases) != 1 || doc.Phases[0].Number != 4 {
			t.Fatalf("phases = %+v, want one empty phase 4", doc.Phases)
		}
	})

	t.Run("tasks before any header", func(t *testing.T) {
		doc := ParseTasks("- [ ] T001 Orphan task\n")
		if doc.Total() != 1 || doc.Phases[0].Title != "" {
			t.Fatalf("orphan task must land in an untitled phase, got %+v", doc)
		}
	})

	t.Run("non-task list lines are ignored", func(t *testing.T) {
		doc := ParseTasks("## Phase 1: Setup\n- plain bullet\n- [ ] no task id here\n- [ ] T010 Real\n")
		if doc.Total() != 1 {
			t.Fatalf("Total() = %d, want 1", doc.Total())
		}
	})
}

func TestTitle(t *testing.T) {
	cases := []struct{ in, want string }{
		{"# Feature Specification: Employee Payroll\n\nBody", "Employee Payroll"},
		{"# Implementation Plan: Employee Payroll\n", "Employee Payroll"},
		{"# Plain Title\n", "Plain Title"},
		{"no heading at all", ""},
	}
	for _, c := range cases {
		if got := Title(c.in); got != c.want {
			t.Errorf("Title(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
