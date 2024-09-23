package extractor

import (
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/kennygrant/sanitize"
)

// List of allowed HTML attributes to keep
var allowedAttributes = []string{"href", "src", "size", "width", "alt", "title", "colspan"}

// ExtractContent extracts the content from the HTML document based on the provided CSS selector.
func ExtractContent(doc *goquery.Document, cssSelector string) (string, error) {
	selection := doc.Find(cssSelector)
	if selection.Length() == 0 {
		return "", fmt.Errorf("CSS selector %s not found", cssSelector)
	}

	// Extract the HTML content for the selected element
	content, err := selection.Html()
	if err != nil {
		return "", fmt.Errorf("error extracting HTML: %w", err)
	}

	// Sanitize the HTML content by allowing only the specified attributes
	sanitizedContent, err := sanitize.HTMLAllowing(content, allowedAttributes)
	if err != nil {
		return "", fmt.Errorf("error sanitizing HTML: %w", err)
	}

	return sanitizedContent, nil
}
