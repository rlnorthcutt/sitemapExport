package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sitemapExport/crawler"
)

// FormatPages formats pages based on the selected format (json, jsonl, txt, md, pdf).
// It returns the formatted string or an error if the format is unsupported.
func FormatPages(pages []crawler.Page, format string) (string, error) {
	switch format {
	case "json":
		return formatJSON(pages)
	case "jsonl":
		return formatJSONLines(pages)
	case "txt", "md", "pdf": // Text-based formats are handled together
		return formatTextBased(pages)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// formatJSON formats the pages as pretty-printed JSON.
func formatJSON(pages []crawler.Page) (string, error) {
	data, err := json.MarshalIndent(pages, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

// formatJSONLines formats each page as a single JSON object per line (JSONL format).
func formatJSONLines(pages []crawler.Page) (string, error) {
	var buffer bytes.Buffer
	for _, page := range pages {
		data, err := json.Marshal(page)
		if err != nil {
			return "", fmt.Errorf("failed to marshal JSON: %w", err)
		}
		buffer.Write(data)     // Avoid conversion to string for better performance
		buffer.WriteByte('\n') // Write new line after each JSON object
	}
	return buffer.String(), nil
}

// formatTextBased formats the pages as text-based output (txt, md, pdf).
// The same format is used for all these cases as plain text.
func formatTextBased(pages []crawler.Page) (string, error) {
	var buffer bytes.Buffer
	for _, page := range pages {
		// Writing directly to buffer with fmt.Fprint instead of fmt.Sprintf
		fmt.Fprintf(&buffer, "# %s\n", page.Title)
		fmt.Fprintf(&buffer, "URL: %s\n", page.URL)
		fmt.Fprintf(&buffer, "Description: %s\n", page.Description)
		fmt.Fprintf(&buffer, "Content:\n%s\n", page.Content)
		buffer.WriteString("\n\n----------------------------------------------\n")
		buffer.WriteString("----------------------------------------------\n\n")
	}
	return buffer.String(), nil
}
