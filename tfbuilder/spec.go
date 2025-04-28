package tfbuilder

import (
	"strings"
)

func addToSpec(spec, after, newLine string) string {
	lines := strings.Split(spec, "\n")
	var result []string
	inserted := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		result = append(result, line)

		if !inserted && strings.Contains(line, after) {
			indent := getIndentation(line)
			result = append(result, indent+newLine)
			inserted = true
		}
	}

	// If "after" not found, just append at end
	if !inserted {
		result = append(result, newLine)
	}

	return strings.Join(result, "\n")
}

func removeFromSpec(spec string, match string) string {
	lines := strings.Split(spec, "\n")
	var filtered []string
	for _, l := range lines {
		if !strings.Contains(l, match) {
			filtered = append(filtered, l)
		}
	}
	return strings.Join(filtered, "\n")
}

func updateSpec(spec, match, newValue string) string {
	return strings.ReplaceAll(spec, match, newValue)
}

func getIndentation(line string) string {
	indent := ""
	for _, r := range line {
		if r == ' ' || r == '\t' {
			indent += string(r)
		} else {
			break
		}
	}
	return indent
}
