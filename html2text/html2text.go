package html2text

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Convert transforms sanitized HTML content into plain text with custom formatting.
// It removes all line breaks in the input HTML before processing.
//
// Convert returns the plain text content and an error if encountered.
func Convert(sanitizedHTML string) (string, error) {
	// Create a new goquery Document from the cleaned HTML
	sanitizedDoc, err := goquery.NewDocumentFromReader(strings.NewReader(sanitizedHTML))
	if err != nil {
		return "", fmt.Errorf("error parsing sanitized HTML: %w", err)
	}

	// Initialize a builder for collecting the text content
	var contentBuilder strings.Builder

	// Process contents starting from the <body> tag
	sanitizedDoc.Find("body").Contents().Each(func(i int, s *goquery.Selection) {
		handleElement(&contentBuilder, s, 0)
	})

	return contentBuilder.String(), nil
}

// handleElement formats different HTML elements into plain text.
//
// contentBuilder appends formatted content, and indent tracks the depth for nested elements.
func handleElement(contentBuilder *strings.Builder, s *goquery.Selection, indent int) {
	tagName := goquery.NodeName(s)

	// Extract text and clean line breaks for non-preformatted elements
	text := s.Text()
	if tagName != "pre" {
		text = strings.ReplaceAll(text, "\n", "")
		text = strings.ReplaceAll(text, "\r", "")
	}

	// Format based on the tag name
	switch tagName {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		contentBuilder.WriteString("\n" + text + "\n" + strings.Repeat("-", len(text)) + "\n")
	case "p":
		contentBuilder.WriteString(text + "\n\n")
	case "ul":
		handleList(contentBuilder, s, indent, false)
	case "ol":
		handleList(contentBuilder, s, indent, true)
	case "li":
		// List items are handled in handleList
	case "br":
		contentBuilder.WriteString("\n")
	case "table":
		handleTable(contentBuilder, s)
		return
	case "a":
		handleAnchor(contentBuilder, s)
	case "img":
		handleImage(contentBuilder, s)
	case "pre":
		contentBuilder.WriteString("CODE:\n" + s.Text() + "\n")
	default:
		contentBuilder.WriteString(text)
	}

	// Recursively process child elements, skipping <pre>, <table>, <ul>, and <ol>
	if tagName != "pre" && tagName != "table" && tagName != "ul" && tagName != "ol" {
		s.Children().Each(func(i int, child *goquery.Selection) {
			handleElement(contentBuilder, child, indent)
		})
	}
}

// handleList processes ordered and unordered lists.
//
// If isOrdered is true, it formats an ordered list; otherwise, it formats an unordered list.
func handleList(contentBuilder *strings.Builder, s *goquery.Selection, indent int, isOrdered bool) {
	count := 1
	indentStr := strings.Repeat("   ", indent)

	s.Children().Each(func(i int, child *goquery.Selection) {
		if goquery.NodeName(child) == "li" {
			if isOrdered {
				contentBuilder.WriteString(fmt.Sprintf("%s%d. ", indentStr, count))
				count++
			} else {
				contentBuilder.WriteString(indentStr + "- ")
			}

			// Process non-list child elements inside <li>
			child.Contents().Each(func(j int, nestedChild *goquery.Selection) {
				if goquery.NodeName(nestedChild) != "ul" && goquery.NodeName(nestedChild) != "ol" {
					handleElement(contentBuilder, nestedChild, indent)
				}
			})

			contentBuilder.WriteString("\n")

			// Recursively process any nested lists inside <li>
			child.Children().Each(func(j int, nestedChild *goquery.Selection) {
				switch goquery.NodeName(nestedChild) {
				case "ul":
					handleList(contentBuilder, nestedChild, indent+1, false)
				case "ol":
					handleList(contentBuilder, nestedChild, indent+1, true)
				}
			})
		}
	})

	// Add an extra line break after the list if needed
	if indent == 0 && !strings.HasSuffix(contentBuilder.String(), "\n\n") {
		contentBuilder.WriteString("\n")
	}
}

// handleTable formats table elements as plain text with pipe-separated rows.
func handleTable(contentBuilder *strings.Builder, s *goquery.Selection) {
	s.Find("tr").Each(func(i int, row *goquery.Selection) {
		handleTableRow(contentBuilder, row)
		contentBuilder.WriteString("\n")
	})
	contentBuilder.WriteString("\n") // Extra line break after the table
}

// handleTableRow formats a table row.
func handleTableRow(contentBuilder *strings.Builder, s *goquery.Selection) {
	first := true
	s.Children().Each(func(i int, cell *goquery.Selection) {
		if !first {
			contentBuilder.WriteString(" | ")
		}
		handleTableCell(contentBuilder, cell)
		first = false
	})
}

// handleTableCell formats table cells as plain text.
func handleTableCell(contentBuilder *strings.Builder, s *goquery.Selection) {
	contentBuilder.WriteString(strings.TrimSpace(s.Text()))
}

// handleAnchor formats <a> tags as "text (URL)".
func handleAnchor(contentBuilder *strings.Builder, s *goquery.Selection) {
	href, exists := s.Attr("href")
	text := s.Text()

	if exists && !strings.HasPrefix(href, "#") {
		contentBuilder.WriteString(text + " (" + href + ") ")
	} else {
		contentBuilder.WriteString(text)
	}
}

// handleImage formats <img> tags as "Image: alt (src)" or "Image: (src)".
func handleImage(contentBuilder *strings.Builder, s *goquery.Selection) {
	src, srcExists := s.Attr("src")
	alt, altExists := s.Attr("alt")

	if srcExists {
		if altExists {
			contentBuilder.WriteString(fmt.Sprintf("\nImage: %s (%s)\n", alt, src))
		} else {
			contentBuilder.WriteString(fmt.Sprintf("\nImage: (%s)\n", src))
		}
		contentBuilder.WriteString("\n")
	}
}
