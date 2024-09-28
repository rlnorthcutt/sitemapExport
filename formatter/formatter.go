package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sitemapExport/crawler"
)

// FormatPages formats pages based on the selected format (json, jsonl, txt, md, pdf).
func FormatPages(pages []crawler.Page, format string) (string, error) {
	switch format {
	case "json":
		return formatJSON(pages)
	case "jsonl":
		return formatJSONLines(pages)
	case "txt", "md", "pdf": // Treat txt, md, and pdf the same
		return formatTextBased(pages)
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

// formatTextBased formats the pages for txt, md, and pdf as simple text output.
func formatTextBased(pages []crawler.Page) (string, error) {
	var buffer bytes.Buffer
	for _, page := range pages {
		buffer.WriteString(fmt.Sprintf("# %s\n", page.Title))
		buffer.WriteString(fmt.Sprintf("URL: %s\n", page.URL))
		buffer.WriteString(fmt.Sprintf("Description: %s\n", page.Description))
		buffer.WriteString(fmt.Sprintf("Content:\n%s\n", page.Content))
		buffer.WriteString("\n\n----------------------------------------------\n")
		buffer.WriteString("----------------------------------------------\n\n")
	}
	return buffer.String(), nil
}
