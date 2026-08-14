package parser

import "strings"

// Title extracts the document title from the first "# " heading, stripping
// the Spec Kit prefixes "Feature Specification:" and "Implementation Plan:".
// It returns "" when no H1 exists.
func Title(content string) string {
	for _, line := range strings.Split(content, "\n") {
		if !strings.HasPrefix(line, "# ") {
			continue
		}
		title := strings.TrimSpace(strings.TrimPrefix(line, "# "))
		for _, prefix := range []string{"Feature Specification:", "Implementation Plan:"} {
			if rest, ok := strings.CutPrefix(title, prefix); ok {
				title = strings.TrimSpace(rest)
				break
			}
		}
		return title
	}
	return ""
}
