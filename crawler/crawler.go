package crawler

import (
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"sitemapExport/html2text"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
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

// List of allowed HTML tags to keep
var allowedTags = []string{"h1", "h2", "h3", "h4", "h5", "h6", "hr", "p", "br", "b", "i", "strong", "em", "ol", "ul", "li", "a", "img", "pre", "code", "blockquote", "tr", "td", "th", "table"}

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
// It also ensures that all relative links and image sources are converted to absolute URLs using the hostDomain.
func extractPage(pageURL, cssSelector, format string) (Page, error) {
	res, err := http.Get(pageURL)
	if err != nil {
		return Page{}, fmt.Errorf("error visiting URL %s: %w", pageURL, err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return Page{}, fmt.Errorf("error parsing HTML from %s: %w", pageURL, err)
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

	// Fix relative URLs for links and images
	hostDomain, err := getDomainFromURL(pageURL)
	if err == nil {
		fixRelativeUrls(doc, hostDomain)
	}

	// Extract and transform content based on the format
	content, err := extractAndTransformContent(doc, cssSelector, format)
	if err != nil {
		return Page{}, err
	}

	// Return the extracted page data
	return Page{
		Title:       title,
		URL:         pageURL,
		Description: description, // Will be omitted if empty
		Tags:        metaTags,    // Will be omitted if empty
		Content:     content,
	}, nil
}

// fixRelativeUrls updates all relative links (href) and image sources (src) to absolute URLs based on the hostDomain.
// It skips any links that are in-page anchors (start with "#").
func fixRelativeUrls(doc *goquery.Document, hostDomain string) {
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && isRelativeURL(href) && !isAnchorLink(href) {
			absoluteURL := toAbsoluteURL(hostDomain, href)
			s.SetAttr("href", absoluteURL)
		}
	})

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if exists && isRelativeURL(src) {
			absoluteURL := toAbsoluteURL(hostDomain, src)
			s.SetAttr("src", absoluteURL)
		}
	})
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
	sanitizedContent, err := sanitize.HTMLAllowing(decodedContent, allowedTags, allowedAttributes)
	if err != nil {
		return "", fmt.Errorf("error sanitizing HTML: %w", err)
	}
	// Step 3: Remove excess newlines and carriage returns (more than 2)
	sanitizedContent = removeExcessNewlines(sanitizedContent)

	// Step 4: Handle the content format (HTML, MD, TXT)
	switch format {
	case "html":
		// Return sanitized HTML
		return sanitizedContent, nil
	case "md":
		// Convert sanitized HTML to Markdown using html-to-markdown
		converter := md.NewConverter("", true, nil) // Using the correct alias 'md'
		converter.Use(plugin.Table())
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

// getDomainFromURL extracts the scheme and host (domain) from the given full URL.
func getDomainFromURL(pageURL string) (string, error) {
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Rebuild the domain as scheme + host (e.g., https://www.example.com)
	domain := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	return domain, nil
}

// isRelativeURL checks if the given URL is a relative URL.
func isRelativeURL(link string) bool {
	u, err := url.Parse(link)
	return err == nil && !u.IsAbs()
}

// isAnchorLink checks if the given link is an in-page anchor (starts with "#").
func isAnchorLink(link string) bool {
	return len(link) > 0 && link[0] == '#'
}

// toAbsoluteURL converts a relative URL to an absolute URL using the given host.
func toAbsoluteURL(host, relativeURL string) string {
	u, err := url.Parse(relativeURL)
	if err != nil {
		return relativeURL // Return as-is if parsing fails
	}

	baseURL, err := url.Parse(host)
	if err != nil {
		return relativeURL // Return as-is if base URL parsing fails
	}

	return baseURL.ResolveReference(u).String()
}

// removeExcessNewlines reduces multiple consecutive newlines, trims spaces, and collapses multiple spaces.
func removeExcessNewlines(content string) string {
	// Step 1: Normalize \r\n and \r to \n
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	// Step 2: Trim leading/trailing spaces from each line
	lines := strings.Split(content, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}

	// Step 3: Join lines back together and collapse multiple spaces to one
	content = strings.Join(lines, "\n")
	content = collapseSpaces(content)

	// Step 4: Replace 3 or more consecutive newlines with exactly 2 newlines
	re := regexp.MustCompile(`\n{3,}`)
	content = re.ReplaceAllString(content, "\n\n")

	return content
}

// collapseSpaces reduces multiple spaces within a line to a single space.
func collapseSpaces(content string) string {
	// Replace multiple spaces with a single space
	re := regexp.MustCompile(`\s{2,}`)
	return re.ReplaceAllString(content, " ")
}
