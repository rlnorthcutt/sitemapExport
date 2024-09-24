package crawler

import (
	"fmt"
	"net/http"
	"sitemapExport/html2text"

	"github.com/PuerkitoBio/goquery"
	"github.com/kennygrant/sanitize"
	"github.com/russross/blackfriday/v2"
)

// Page represents the extracted data for a single page.
type Page struct {
	Title       string
	URL         string
	Description string
	MetaTags    []string
	Content     string
}

// CrawlSitemap fetches the sitemap, parses it, and crawls each page to extract content.
func CrawlSitemap(sitemapURL, cssSelector, format string) ([]Page, error) {
	var pages []Page

	// Fetch the sitemap
	res, err := http.Get(sitemapURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sitemap: %w", err)
	}
	defer res.Body.Close()

	// Parse the XML sitemap (assuming standard XML format here)
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error parsing sitemap: %w", err)
	}

	// Loop through each URL and crawl the page
	doc.Find("url loc").Each(func(i int, s *goquery.Selection) {
		url := s.Text()
		page, err := extractPage(url, cssSelector, format)
		if err != nil {
			fmt.Printf("Error extracting page %s: %v\n", url, err)
			return
		}
		pages = append(pages, page)
	})

	return pages, nil
}

// extractPage fetches the page at the given URL and extracts the content based on the CSS selector.
func extractPage(url, cssSelector, format string) (Page, error) {
	res, err := http.Get(url)
	if err != nil {
		return Page{}, fmt.Errorf("error visiting URL %s: %w", url, err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return Page{}, fmt.Errorf("error parsing HTML from %s: %w", url, err)
	}

	// Extract the title, meta description, and meta tags
	title := doc.Find("title").Text()
	description, _ := doc.Find("meta[name=description]").Attr("content")
	metaTags := doc.Find("meta").Map(func(i int, s *goquery.Selection) string {
		name, _ := s.Attr("name")
		content, _ := s.Attr("content")
		return fmt.Sprintf("%s: %s", name, content)
	})

	// Extract and transform content based on the format
	content, err := extractAndTransformContent(doc, cssSelector, format)
	if err != nil {
		return Page{}, err
	}

	return Page{
		Title:       title,
		URL:         url,
		Description: description,
		MetaTags:    metaTags,
		Content:     content,
	}, nil
}

// extractAndTransformContent extracts the content and applies the appropriate transformation (HTML, MD, or TXT).
func extractAndTransformContent(doc *goquery.Document, cssSelector, format string) (string, error) {
	selection := doc.Find(cssSelector)
	if selection.Length() == 0 {
		return "", fmt.Errorf("CSS selector %s not found", cssSelector)
	}

	content, err := selection.Html()
	if err != nil {
		return "", fmt.Errorf("error extracting HTML: %w", err)
	}

	// Sanitize the HTML
	sanitizedContent, err := sanitize.HTMLAllowing(content)
	if err != nil {
		return "", fmt.Errorf("error sanitizing HTML: %w", err)
	}

	switch format {
	case "html":
		return sanitizedContent, nil
	case "md":
		// Convert to Markdown using blackfriday
		input := []byte(sanitizedContent)
		return string(blackfriday.Run(input)), nil
	case "txt":
		// Convert to plain text
		textContent, err := html2text.Convert(doc, cssSelector)
		if err != nil {
			return "", fmt.Errorf("error converting HTML to text: %w", err)
		}
		return textContent, nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}
