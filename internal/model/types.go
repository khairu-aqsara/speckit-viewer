// Package model holds the shared domain types for a scanned Spec Kit project.
package model

// FileKind classifies a file inside a feature folder.
type FileKind int

const (
	FileOther FileKind = iota
	FileSpec
	FilePlan
	FileTasks
	FileResearch
	FileDataModel
	FileQuickstart
	FileChecklist
	FileContract
)

// SpecKitFile is one file inside a feature folder.
type SpecKitFile struct {
	Kind FileKind
	// Rel is the path relative to the feature folder, slash-separated,
	// e.g. "spec.md" or "checklists/requirements.md".
	Rel string
	// Abs is the absolute path on disk. Empty in pure unit tests.
	Abs string
}

// Feature is one specs/NNN-name folder.
type Feature struct {
	ID    string // "031"
	Slug  string // "employee-payroll"
	Name  string // "Employee Payroll"
	Dir   string // absolute folder path
	Files []SpecKitFile
}

// PresentKinds reports which file kinds exist in the feature.
func (f Feature) PresentKinds() map[FileKind]bool {
	present := make(map[FileKind]bool, len(f.Files))
	for _, file := range f.Files {
		present[file.Kind] = true
	}
	return present
}

// File returns the first file of the given kind, or nil.
func (f Feature) File(kind FileKind) *SpecKitFile {
	for i := range f.Files {
		if f.Files[i].Kind == kind {
			return &f.Files[i]
		}
	}
	return nil
}

// Project is a scanned Spec Kit project root.
type Project struct {
	Root             string
	ConstitutionPath string // "" when .specify/memory/constitution.md is absent
	Features         []Feature
}
