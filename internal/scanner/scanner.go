// Package scanner walks a Spec Kit project root into model types. The
// filesystem walk is thin; classification lives in pure functions.
package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/wenkhairu/speckit-viewer/internal/model"
)

var featureDirRe = regexp.MustCompile(`^(\d+)-(.+)$`)

// ClassifyFile maps a feature-relative slash path to a file kind.
// Unrecognized files are FileOther, never an error: Spec Kit templates vary
// across versions, so the scanner tolerates anything.
func ClassifyFile(rel string) model.FileKind {
	dir, base := "", rel
	if i := strings.IndexByte(rel, '/'); i >= 0 {
		dir, base = rel[:i], rel[i+1:]
	}
	if !strings.HasSuffix(base, ".md") || strings.Contains(base, "/") {
		return model.FileOther
	}
	switch dir {
	case "":
		switch base {
		case "spec.md":
			return model.FileSpec
		case "plan.md":
			return model.FilePlan
		case "tasks.md":
			return model.FileTasks
		case "research.md":
			return model.FileResearch
		case "data-model.md":
			return model.FileDataModel
		case "quickstart.md":
			return model.FileQuickstart
		}
		return model.FileOther
	case "checklists":
		return model.FileChecklist
	case "contracts":
		return model.FileContract
	}
	return model.FileOther
}

// BuildFeature builds a Feature from a folder name and its file paths
// (feature-relative, slash-separated). It is pure: no filesystem access.
// It returns false when dirName is not a NNN-name feature folder.
func BuildFeature(dirName string, rels []string) (model.Feature, bool) {
	m := featureDirRe.FindStringSubmatch(dirName)
	if m == nil {
		return model.Feature{}, false
	}
	f := model.Feature{ID: m[1], Slug: m[2], Name: titleFromSlug(m[2])}
	sorted := append([]string(nil), rels...)
	sort.Strings(sorted)
	for _, rel := range sorted {
		f.Files = append(f.Files, model.SpecKitFile{Kind: ClassifyFile(rel), Rel: rel})
	}
	return f, true
}

func titleFromSlug(slug string) string {
	words := strings.Split(slug, "-")
	for i, w := range words {
		if w != "" {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// Scan walks root and returns the project. It fails only when root/specs is
// missing or unreadable; anything odd inside is skipped, not fatal.
func Scan(root string) (model.Project, error) {
	specsDir := filepath.Join(root, "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return model.Project{}, fmt.Errorf("%s does not look like a Spec Kit project: %w", root, err)
	}
	project := model.Project{Root: root}

	constitution := filepath.Join(root, ".specify", "memory", "constitution.md")
	if _, err := os.Stat(constitution); err == nil {
		project.ConstitutionPath = constitution
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		featureDir := filepath.Join(specsDir, entry.Name())
		rels := listFilesRel(featureDir)
		feature, ok := BuildFeature(entry.Name(), rels)
		if !ok {
			continue
		}
		feature.Dir = featureDir
		for i := range feature.Files {
			feature.Files[i].Abs = filepath.Join(featureDir, filepath.FromSlash(feature.Files[i].Rel))
		}
		project.Features = append(project.Features, feature)
	}
	sort.Slice(project.Features, func(i, j int) bool {
		return project.Features[i].ID < project.Features[j].ID
	})
	return project, nil
}

func listFilesRel(dir string) []string {
	var rels []string
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return nil
		}
		rels = append(rels, filepath.ToSlash(rel))
		return nil
	})
	return rels
}
