package tools

import (
	"strings"
)

// cleanHTML removes script and style tags, strips HTML tags,
// unescapes standard entities, and cleans up whitespaces.
func cleanHTML(html string) string {
	// Simple script and style tag remover
	for {
		startIdx := strings.Index(strings.ToLower(html), "<script")
		if startIdx == -1 {
			break
		}
		endIdx := strings.Index(strings.ToLower(html[startIdx:]), "</script>")
		if endIdx == -1 {
			break
		}
		html = html[:startIdx] + html[startIdx+endIdx+len("</script>"):]
	}

	for {
		startIdx := strings.Index(strings.ToLower(html), "<style")
		if startIdx == -1 {
			break
		}
		endIdx := strings.Index(strings.ToLower(html[startIdx:]), "</style>")
		if endIdx == -1 {
			break
		}
		html = html[:startIdx] + html[startIdx+endIdx+len("</style>"):]
	}

	// Remove all HTML tags
	var sb strings.Builder
	inTag := false
	for _, r := range html {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
			sb.WriteRune(' ') // Replace tags with space to avoid merging adjacent texts
		} else if !inTag {
			sb.WriteRune(r)
		}
	}

	// Unescape standard HTML entities
	text := sb.String()
	replacer := strings.NewReplacer(
		"&quot;", "\"",
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&#39;", "'",
		"&nbsp;", " ",
	)
	text = replacer.Replace(text)

	// Clean up whitespace
	lines := strings.Split(text, "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleanedLines = append(cleanedLines, trimmed)
		}
	}
	return strings.Join(cleanedLines, "\n")
}

// extractTitle extracts the <title> tag content from HTML.
func extractTitle(html string) string {
	lower := strings.ToLower(html)
	start := strings.Index(lower, "<title>")
	if start == -1 {
		return ""
	}
	end := strings.Index(lower[start:], "</title>")
	if end == -1 {
		return ""
	}
	title := html[start+len("<title>") : start+end]
	return strings.TrimSpace(title)
}
