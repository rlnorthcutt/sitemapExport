package html2text

import (
	"fmt"
	"html"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/kennygrant/sanitize"
)

// html2text converts the HTML content into plain text with custom formatting.
func Convert(doc *goquery.Document, cssSelector string) (string, error) {
	// 1. Get the content using the CSS selector
	selection := doc.Find(cssSelector)
	if selection.Length() == 0 {
		return "", fmt.Errorf("CSS selector %s not found", cssSelector)
	}

	// 2. Sanitize the HTML content using HTMLAllowing with default settings
	content, err := selection.Html()
	if err != nil {
		return "", fmt.Errorf("error extracting HTML: %w", err)
	}
	sanitizedContent, err := sanitize.HTMLAllowing(content)
	if err != nil {
		return "", fmt.Errorf("error sanitizing HTML: %w", err)
	}

	// 3. Loop through the content items and apply handleElement to format the text
	var contentBuilder strings.Builder
	sanitizedDoc, err := goquery.NewDocumentFromReader(strings.NewReader(sanitizedContent))
	if err != nil {
		return "", fmt.Errorf("error parsing sanitized HTML: %w", err)
	}

	sanitizedDoc.Contents().Each(func(i int, s *goquery.Selection) {
		handleElement(&contentBuilder, s)
	})

	// 4. Decode HTML entities (e.g., &#34; -> ")
	decodedContent := html.UnescapeString(contentBuilder.String())

	return decodedContent, nil
}

// handleElement formats different HTML elements as text.
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
