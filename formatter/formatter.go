package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sitemapExport/crawler"
	"strings"

	"github.com/russross/blackfriday/v2"
)

// FormatPages formats pages based on the selected format.
func FormatPages(pages []crawler.Page, format string) (string, error) {
	switch format {
	case "json":
		return formatJSON(pages)
	case "jsonl":
		return formatJSONLines(pages) // New format
	case "md":
		return formatMarkdown(pages)
	case "txt":
		return formatText(pages)
	case "pdf":
		return formatText(pages) // Will be converted to PDF in the writer
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// formatJSON formats the pages as pretty JSON.
func formatJSON(pages []crawler.Page) (string, error) {
	data, err := json.MarshalIndent(pages, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// formatJSONLines formats each page as a single JSON object per line (JSONL format).
func formatJSONLines(pages []crawler.Page) (string, error) {
	var buffer bytes.Buffer
	for _, page := range pages {
		data, err := json.Marshal(page)
		if err != nil {
			return "", err
		}
		buffer.WriteString(string(data) + "\n")
	}
	return buffer.String(), nil
}

// formatMarkdown formats the pages as Markdown.
func formatMarkdown(pages []crawler.Page) (string, error) {
	var buffer bytes.Buffer
	for _, page := range pages {
		buffer.WriteString(fmt.Sprintf("# %s\n\n", page.Title))
		buffer.WriteString(fmt.Sprintf("URL: %s\n\n", page.URL))
		buffer.WriteString(fmt.Sprintf("Description: %s\n\n", page.Description))

		// Convert the byte slice to a string
		buffer.WriteString(string(blackfriday.Run([]byte(page.Content))))
		buffer.WriteString("\n\n---\n\n")
	}
	return buffer.String(), nil
}

// formatText formats the pages as plain text.
func formatText(pages []crawler.Page) (string, error) {
	var buffer bytes.Buffer
	for _, page := range pages {
		buffer.WriteString(fmt.Sprintf("Title: %s\n", page.Title))
		buffer.WriteString(fmt.Sprintf("URL: %s\n", page.URL))
		buffer.WriteString(fmt.Sprintf("Description: %s\n", page.Description))
		buffer.WriteString(fmt.Sprintf("Content:\n%s\n", strings.TrimSpace(page.Content)))
		buffer.WriteString("\n\n---\n\n")
	}
	return buffer.String(), nil
}
