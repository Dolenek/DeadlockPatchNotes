package patches

import "strings"

func parseStructuredSectionHeader(line string) (string, bool) {
	match := structuredSectionHeaderRegex.FindStringSubmatch(strings.TrimSpace(line))
	if len(match) == 2 {
		return strings.ToLower(match[1]), true
	}

	switch strings.ToLower(strings.TrimSpace(line)) {
	case "general":
		return "general", true
	case "items":
		return "items", true
	case "heroes":
		return "heroes", true
	}
	return "", false
}

func parseStructuredPrefixedLine(line string) (string, string, bool) {
	match := structuredPrefixedLineRegex.FindStringSubmatch(strings.TrimSpace(line))
	if len(match) != 3 {
		return "", "", false
	}
	return strings.TrimSpace(match[1]), strings.TrimSpace(match[2]), true
}

func shouldSkipTimelineLine(line string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))
	if lower == "" {
		return true
	}
	if lower == "read more" {
		return true
	}
	if strings.HasPrefix(lower, "deadlock - ") && strings.Contains(lower, "steam news") {
		return true
	}
	return structuredDateHeadingRegex.MatchString(line)
}

func cleanTimelineLine(line string) string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "-")
	line = strings.TrimPrefix(line, "*")
	line = strings.TrimSpace(line)
	return line
}
