package html2text

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Convert transforms sanitized HTML content into plain text with custom formatting.
// It also removes all line breaks in the input HTML before processing.
//
// @param sanitizedHTML string The sanitized HTML content.
// @return string The converted plain text content.
// @return error Any error encountered during the conversion process.
func Convert(sanitizedHTML string) (string, error) {
	// 1. Create a new goquery Document from the cleaned HTML
	sanitizedDoc, err := goquery.NewDocumentFromReader(strings.NewReader(sanitizedHTML))
	if err != nil {
		return "", fmt.Errorf("error parsing sanitized HTML: %w", err)
	}

	// 2. Initialize a builder for collecting the text content
	var contentBuilder strings.Builder

	// 3. Loop through the contents and apply handleElement for formatting
	// Note: goquery adds a <html> and <head> tag, so we start from the <body>
	sanitizedDoc.Find("body").Contents().Each(func(i int, s *goquery.Selection) {
		handleElement(&contentBuilder, s, 0) // Start at root level with no indentation
	})

	return contentBuilder.String(), nil
}

// handleElement formats different HTML elements as text, supporting nested lists and tables.
// This function converts the selected HTML elements into plain text representations.
//
// @param contentBuilder *strings.Builder The string builder to append formatted content to.
// @param s *goquery.Selection The selected HTML element to format.
// @param indent int The indentation level for nested elements, automatically increases with recursion.
func handleElement(contentBuilder *strings.Builder, s *goquery.Selection, indent int) {
	tagName := goquery.NodeName(s)

	// Remove line breaks for non-preformatted elements
	text := s.Text()
	if tagName != "pre" {
		text = strings.ReplaceAll(text, "\n", "")
		text = strings.ReplaceAll(text, "\r", "")
	}

	// Handle different HTML tags
	switch tagName {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		// Handle headings with underlines
		contentBuilder.WriteString(fmt.Sprintf("\n"))
		contentBuilder.WriteString(fmt.Sprintf("\n%s\n", text))
		contentBuilder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("-", len(text))))
	case "p":
		contentBuilder.WriteString(text + "\n\n") // Paragraphs with double line breaks
	case "ul":
		// Handle unordered list
		handleList(contentBuilder, s, indent, false) // Pass false for unordered lists (dashes)
	case "ol":
		// Handle ordered list
		handleList(contentBuilder, s, indent, true) // Pass true for ordered lists (numbers)
	case "li":
		// List items are handled in handleList
	case "br":
		contentBuilder.WriteString("\n") // Handle <br> as a line break
	case "table":
		// Handle table as an element and skip default text processing
		handleTable(contentBuilder, s)
		return // Prevent further text processing for table content
	case "a":
		handleAnchor(contentBuilder, s) // Handle <a> tags
	case "img":
		handleImage(contentBuilder, s) // Handle <img> tags
	case "pre":
		// Preserve formatting and line breaks for preformatted text
		contentBuilder.WriteString("CODE:\n")
		contentBuilder.WriteString(s.Text() + "\n")
	default:
		// Default handling for other tags
		contentBuilder.WriteString(text)
	}

	// Continue recursively handling child elements, skipping "pre" elements since we want their text unchanged
	if tagName != "pre" && tagName != "table" && tagName != "ul" && tagName != "ol" {
		s.Children().Each(func(i int, child *goquery.Selection) {
			handleElement(contentBuilder, child, indent)
		})
	}
}

// handleTable processes tables and formats them as text with rows and pipe-separated columns.
//
// @param contentBuilder *strings.Builder The string builder to append the formatted table content to.
// @param s *goquery.Selection The selected table element to process.
func handleTable(contentBuilder *strings.Builder, s *goquery.Selection) {
	s.Find("tr").Each(func(i int, row *goquery.Selection) {
		handleTableRow(contentBuilder, row)
		contentBuilder.WriteString("\n") // Add a line break after each row
	})
	contentBuilder.WriteString("\n") // Add an extra line break after the table
}

// handleTableRow formats table rows and cells.
// Converts table rows into a pipe-separated plain text format.
//
// @param contentBuilder *strings.Builder The string builder to append the formatted table content to.
// @param s *goquery.Selection The selected row element to process.
func handleTableRow(contentBuilder *strings.Builder, s *goquery.Selection) {
	first := true
	s.Children().Each(func(i int, cell *goquery.Selection) {
		if !first {
			contentBuilder.WriteString(" | ") // Add separator for each cell
		}
		handleTableCell(contentBuilder, cell) // Process each cell
		first = false
	})
}

// handleTableCell processes individual table cells (td and th).
//
// @param contentBuilder *strings.Builder The string builder to append the formatted cell content to.
// @param s *goquery.Selection The selected cell element (td or th) to process.
func handleTableCell(contentBuilder *strings.Builder, s *goquery.Selection) {
	contentBuilder.WriteString(strings.TrimSpace(s.Text()))
}

// handleList processes both ordered and unordered lists.
// If isOrdered is true, it will render numbers, otherwise dashes.
// @param contentBuilder *strings.Builder The string builder to append formatted content to.
// @param s *goquery.Selection The selected HTML element to format.
// @param indent int The indentation level for nested elements, automatically increases with recursion.
// @param isOrdered bool Determines whether to render a numbered list (true) or unordered list (false).
func handleList(contentBuilder *strings.Builder, s *goquery.Selection, indent int, isOrdered bool) {
	count := 1 // For numbering ordered lists

	s.Children().Each(func(i int, child *goquery.Selection) {
		if goquery.NodeName(child) == "li" {
			// Add the appropriate prefix (dash or number) based on list type
			if isOrdered {
				contentBuilder.WriteString(fmt.Sprintf("%s%d. ", strings.Repeat("   ", indent), count)) // Write the list prefix (numbered or dashed)
				count++
			} else {
				contentBuilder.WriteString(fmt.Sprintf("%s- ", strings.Repeat("   ", indent))) // Write the list prefix (dashed)
			}

			// Process all inline contents inside the <li> (excluding nested lists)
			child.Contents().Each(func(j int, nestedChild *goquery.Selection) {
				// Only process non-list elements here (like text, links, etc.)
				if goquery.NodeName(nestedChild) != "ul" && goquery.NodeName(nestedChild) != "ol" {
					handleElement(contentBuilder, nestedChild, indent) // Process inline elements within <li>
				}
			})

			contentBuilder.WriteString("\n") // Add a line break after each list item

			// Now, handle any nested lists inside the current list item
			child.Children().Each(func(j int, nestedChild *goquery.Selection) {
				switch goquery.NodeName(nestedChild) {
				case "ul":
					handleList(contentBuilder, nestedChild, indent+1, false) // Unordered list
				case "ol":
					handleList(contentBuilder, nestedChild, indent+1, true) // Ordered list
				}
			})
		}
	})

	// If this is the last list item, add an extra line break
	if contentBuilder.Len() > 0 && strings.HasSuffix(contentBuilder.String(), "\n") && indent == 0 {
		contentBuilder.WriteString("\n")
	}
}

// handleAnchor formats <a> tags as "text (URL)".
// Converts anchor tags to a plain text link format: `text (URL)`.
//
// @param contentBuilder *strings.Builder The string builder to append the formatted link to.
// @param s *goquery.Selection The selected anchor element to process.
func handleAnchor(contentBuilder *strings.Builder, s *goquery.Selection) {
	href, exists := s.Attr("href")
	text := s.Text()

	// If href exists and doesn't start with "#", include the link; otherwise, just output the text
	if exists && !strings.HasPrefix(href, "#") {
		contentBuilder.WriteString(fmt.Sprintf("%s (%s) ", text, href)) // Output "text (URL)"
	} else {
		contentBuilder.WriteString(text) // Output just the text if no href or href starts with "#"
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
		contentBuilder.WriteString(fmt.Sprintf("\nImage: %s (%s)\n", alt, src)) // Markdown-like format for images
	} else if srcExists {
		contentBuilder.WriteString(fmt.Sprintf("\nImage: (%s)\n", src)) // Handle images with no alt text
	}
	contentBuilder.WriteString("\n") // Add a line break after each image
}
