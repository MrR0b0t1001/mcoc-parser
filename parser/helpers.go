package parser

import (
	"html"
	"regexp"
	"slices"
	"strings"
)

func normalizeChampionName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

func cleanText(s string) string {
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "\u00a0", " ") // nbsp
	s = strings.TrimSpace(s)
	spaceRe := regexp.MustCompile(`\s+`)
	s = spaceRe.ReplaceAllString(s, " ")
	return s
}

func splitCSVLike(s string) []string {
	s = cleanText(s)
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func findAfterLabel(text, label string) string {
	i := strings.Index(text, label)
	if i == -1 {
		return ""
	}
	return strings.TrimSpace(text[i+len(label):])
}

func appendUnique(dst []string, v string) []string {
	if slices.Contains(dst, v) {
		return dst
	}
	return append(dst, v)
}

func looksLikeMetaNote(s string) bool {
	// Example: "Passive" lines we *do* want, but “Expert Player Notes, Recommended Masteries...” we don’t.
	// Keep this conservative; adjust as you see pages.
	return strings.HasPrefix(strings.ToLower(s), "expert player notes")
}
