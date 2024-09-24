package crawler

import (
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"sitemapExport/html2text"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/kennygrant/sanitize"
)

// RSSItem represents an RSS item with the fields we care about
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

// RSSFeed represents the RSS feed structure
type RSSFeed struct {
	Items []RSSItem `xml:"channel>item"`
}

// Page represents the extracted data for a single page.
// The `omitempty` tags will prevent empty fields from being included in the JSON output.
type Page struct {
	Title       string   `json:"Title"`
	URL         string   `json:"URL"`
	Description string   `json:"Description,omitempty"` // Omits if empty
	Tags        []string `json:"Tags,omitempty"`        // Omits if empty
	Content     string   `json:"Content"`
}

// List of allowed HTML attributes to keep
var allowedAttributes = []string{"href", "src", "size", "width", "alt", "title", "colspan"}

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

// CrawlRSS fetches the RSS feed, parses it using encoding/xml, and extracts content.
func CrawlRSS(rssURL, cssSelector, format string) ([]Page, error) {
	var pages []Page

	// Fetch the RSS feed
	res, err := http.Get(rssURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
	}
	defer res.Body.Close()

	// Step 1: Parse the RSS feed using encoding/xml
	var rss RSSFeed
	decoder := xml.NewDecoder(res.Body)
	if err := decoder.Decode(&rss); err != nil {
		return nil, fmt.Errorf("error decoding RSS feed: %w", err)
	}

	// Step 2: Process each RSS item and use extractPage for each link
	for _, item := range rss.Items {
		// Check if the link is empty
		if item.Link == "" {
			fmt.Println("Error: RSS item missing URL. Skipping item.")
			continue
		}

		// Extract the page using the URL from the RSS <link> tag
		page, err := extractPage(item.Link, cssSelector, format)
		if err != nil {
			fmt.Printf("Error extracting page %s: %v\n", item.Link, err)
			continue
		}

		// Set the description from the RSS feed
		page.Description = item.Description

		// Add the page to the result list
		pages = append(pages, page)
	}

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

	// Extract the title
	title := doc.Find("title").Text()

	// Conditionally extract the description meta tag
	description, _ := doc.Find("meta[name=description]").Attr("content")

	// Conditionally extract the "tags" meta tag
	tags, tagsExists := doc.Find("meta[name=tags]").Attr("content")

	// Prepare a slice for meta tags and add only the ones that exist
	var metaTags []string
	if tagsExists && tags != "" {
		metaTags = append(metaTags, tags)
	}

	// Extract and transform content based on the format
	content, err := extractAndTransformContent(doc, cssSelector, format)
	if err != nil {
		return Page{}, err
	}

	// Return the extracted page data
	return Page{
		Title:       title,
		URL:         url,
		Description: description, // Will be omitted if empty
		Tags:        metaTags,    // Will be omitted if empty
		Content:     content,
	}, nil
}

// extractAndTransformContent extracts the content and applies the appropriate transformation (HTML, MD, or TXT).
func extractAndTransformContent(doc *goquery.Document, cssSelector, format string) (string, error) {
	// Find the HTML content using the CSS selector
	selection := doc.Find(cssSelector)
	if selection.Length() == 0 {
		return "", fmt.Errorf("CSS selector %s not found", cssSelector)
	}

	// Extract the raw HTML content
	content, err := selection.Html()
	if err != nil {
		return "", fmt.Errorf("error extracting HTML: %w", err)
	}

	return extractAndTransformContentFromText(content, format)
}

// extractAndTransformContentFromText applies the appropriate transformation (HTML, MD, TXT) to text content.
func extractAndTransformContentFromText(content, format string) (string, error) {
	// Step 1: Decode HTML entities before sanitization
	decodedContent := html.UnescapeString(content)

	// Step 2: Sanitize the HTML and remove unwanted elements
	sanitizedContent, err := sanitize.HTMLAllowing(decodedContent, allowedAttributes)
	if err != nil {
		return "", fmt.Errorf("error sanitizing HTML: %w", err)
	}

	// Step 3: Handle the content format (HTML, MD, TXT)
	switch format {
	case "html":
		// Return sanitized HTML
		return sanitizedContent, nil
	case "md":
		// Convert sanitized HTML to Markdown using html-to-markdown
		converter := md.NewConverter("", true, nil) // Using the correct alias 'md'
		mdContent, err := converter.ConvertString(sanitizedContent)
		if err != nil {
			return "", fmt.Errorf("error converting HTML to Markdown: %w", err)
		}
		return mdContent, nil
	case "txt":
		// Convert to plain text using the html2text package
		textContent, err := html2text.Convert(sanitizedContent)
		if err != nil {
			return "", fmt.Errorf("error converting HTML to text: %w", err)
		}
		return textContent, nil
	default:
		// Unsupported format
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}
