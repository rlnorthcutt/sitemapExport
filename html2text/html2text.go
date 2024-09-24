package html2text

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Convert transforms sanitized HTML content into plain text with custom formatting.
//
// @param sanitizedHTML string The sanitized HTML content.
// @return string The converted plain text content.
// @return error Any error encountered during the conversion process.
func Convert(sanitizedHTML string) (string, error) {
	// 1. Create a new goquery Document from the sanitized HTML
	sanitizedDoc, err := goquery.NewDocumentFromReader(strings.NewReader(sanitizedHTML))
	if err != nil {
		return "", fmt.Errorf("error parsing sanitized HTML: %w", err)
	}

	// 2. Initialize a builder for collecting the text content
	var contentBuilder strings.Builder

	// 3. Loop through the contents and apply handleElement for formatting
	sanitizedDoc.Contents().Each(func(i int, s *goquery.Selection) {
		handleElement(&contentBuilder, s)
	})

	return contentBuilder.String(), nil
}

// handleElement formats different HTML elements as text.
// This function converts the selected HTML elements into plain text representations.
//
// @param contentBuilder *strings.Builder The string builder to append formatted content to.
// @param s *goquery.Selection The selected HTML element to format.
func handleElement(contentBuilder *strings.Builder, s *goquery.Selection) {
	switch goquery.NodeName(s) {
	case "p":
		contentBuilder.WriteString(s.Text() + "\n\n") // Paragraphs with double line breaks
	case "li":
		contentBuilder.WriteString("- " + s.Text() + "\n") // List items with bullet points
	case "tr":
		handleTableRow(contentBuilder, s) // Handle table rows
	case "br":
		contentBuilder.WriteString("\n") // Line breaks
	case "h1", "h2", "h3", "h4", "h5", "h6":
		contentBuilder.WriteString("\n* " + s.Text() + "\n") // Headings with an asterisk (*) in front
	case "a":
		handleAnchor(contentBuilder, s) // Handle <a> tags
	case "img":
		handleImage(contentBuilder, s) // Handle <img> tags
	default:
		contentBuilder.WriteString(s.Text()) // Default handling for other tags
	}
}

// handleTableRow formats table rows and cells.
// Converts table rows into a pipe-separated plain text format.
//
// @param contentBuilder *strings.Builder The string builder to append the formatted table content to.
// @param s *goquery.Selection The selected row element to process.
func handleTableRow(contentBuilder *strings.Builder, s *goquery.Selection) {
	s.Find("td, th").Each(func(i int, cell *goquery.Selection) {
		if i > 0 {
			contentBuilder.WriteString(" | ") // Add column separator for cells
		}
		contentBuilder.WriteString(cell.Text())
	})
	contentBuilder.WriteString("\n") // Line break after each row
}

// handleAnchor formats <a> tags as "text (URL)".
// Converts anchor tags to a plain text link format: `text (URL)`.
//
// @param contentBuilder *strings.Builder The string builder to append the formatted link to.
// @param s *goquery.Selection The selected anchor element to process.
func handleAnchor(contentBuilder *strings.Builder, s *goquery.Selection) {
	href, exists := s.Attr("href")
	text := s.Text()
	if exists {
		contentBuilder.WriteString(fmt.Sprintf("%s (%s)", text, href)) // Format link as "text (URL)"
	} else {
		contentBuilder.WriteString(text) // If no href, just append the text
	}
}

// handleImage formats <img> tags as "![alt text](src)" or "Image: (src)".
// Converts image tags to a plain text format: `![alt](src)` or `Image: (src)`.
//
// @param contentBuilder *strings.Builder The string builder to append the formatted image to.
// @param s *goquery.Selection The selected image element to process.
func handleImage(contentBuilder *strings.Builder, s *goquery.Selection) {
	src, srcExists := s.Attr("src")
	alt, altExists := s.Attr("alt")

	if srcExists && altExists {
		contentBuilder.WriteString(fmt.Sprintf("![%s](%s)", alt, src)) // Markdown-like format for images
	} else if srcExists {
		contentBuilder.WriteString(fmt.Sprintf("Image: (%s)", src)) // Handle images with no alt text
	}
	contentBuilder.WriteString("\n") // Add a line break after each image
}
