package crawler

import (
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/url"
	"os"
	"regexp"
	"sitemapExport/html2text"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/PuerkitoBio/goquery"
	"github.com/kennygrant/sanitize"
	"github.com/schollz/progressbar/v3"
)

// RSSItem represents an RSS item with relevant fields.
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

// RSSFeed represents the structure of an RSS feed.
type RSSFeed struct {
	Items []RSSItem `xml:"channel>item"`
}

// Page represents the extracted data for a single page.
type Page struct {
	Title       string   `json:"Title"`
	URL         string   `json:"URL"`
	Description string   `json:"Description,omitempty"`
	Tags        []string `json:"Tags,omitempty"`
	Content     string   `json:"Content"`
}

// List of allowed HTML attributes and tags.
var allowedAttributes = []string{"href", "src", "size", "width", "alt", "title", "colspan"}
var allowedTags = []string{"h1", "h2", "h3", "h4", "h5", "h6", "hr", "p", "br", "b", "i", "strong", "em", "ol", "ul", "li", "a", "img", "pre", "code", "blockquote", "tr", "td", "th", "table"}

var reExcessNewlines = regexp.MustCompile(`\n{3,}`)
var reCollapseSpaces = regexp.MustCompile(`\s{2,}`)

// CrawlSitemap fetches and processes a sitemap from a URL or file to extract page content, showing progress.
// Only URLs matching the provided filter pattern are included.
func CrawlSitemap(sitemapSource, cssSelector, format, filter string) ([]Page, error) {
	var pages []Page

	var (
		reader io.ReadCloser
		err    error
	)

	if strings.HasPrefix(sitemapSource, "http://") || strings.HasPrefix(sitemapSource, "https://") {
		// Fetch the sitemap from URL
		res, err := Client.Get(sitemapSource)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch sitemap: %w", err)
		}
		reader = res.Body
	} else {
		// Open sitemap from local file
		reader, err = os.Open(sitemapSource)
		if err != nil {
			return nil, fmt.Errorf("failed to open sitemap file: %w", err)
		}
	}
	defer reader.Close()

	// Parse the XML sitemap
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("error parsing sitemap: %w", err)
	}

	// Get the number of URLs for the progress bar
	urls := doc.Find("url loc")
	totalURLs := urls.Length()

	// Initialize the progress bar
	bar := progressbar.NewOptions(totalURLs, progressbar.OptionSetDescription("Fetching sitemap pages"))

	// Extract each URL from the sitemap and crawl the page
	urls.Each(func(i int, s *goquery.Selection) {
		pageURL := s.Text()
		if !matchesFilter(pageURL, filter) {
			bar.Add(1)
			return
		}
		page, err := extractPage(pageURL, cssSelector, format)
		if err != nil {
			fmt.Printf("Error extracting page %s: %v\n", pageURL, err)
			bar.Add(1)
			return
		}
		pages = append(pages, page)
		bar.Add(1) // Increment the progress bar
	})

	return pages, nil
}

// CrawlRSS fetches and processes an RSS feed from a URL or file to extract page content, showing progress.
// Only URLs matching the provided filter pattern are included.
func CrawlRSS(rssSource, cssSelector, format, filter string) ([]Page, error) {
	var pages []Page

	var (
		reader io.ReadCloser
		err    error
	)

	if strings.HasPrefix(rssSource, "http://") || strings.HasPrefix(rssSource, "https://") {
		// Fetch the RSS feed from URL
		res, err := Client.Get(rssSource)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
		}
		reader = res.Body
	} else {
		// Open RSS feed from local file
		reader, err = os.Open(rssSource)
		if err != nil {
			return nil, fmt.Errorf("failed to open RSS file: %w", err)
		}
	}
	defer reader.Close()

	// Parse the RSS feed using encoding/xml
	var rss RSSFeed
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(&rss); err != nil {
		return nil, fmt.Errorf("error decoding RSS feed: %w", err)
	}

	// Initialize the progress bar
	totalItems := len(rss.Items)
	bar := progressbar.NewOptions(totalItems, progressbar.OptionSetDescription("Fetching RSS pages"))

	// Process each RSS item
	for _, item := range rss.Items {
		if item.Link == "" {
			fmt.Println("Error: RSS item missing URL. Skipping item.")
			bar.Add(1)
			continue
		}
		if !matchesFilter(item.Link, filter) {
			bar.Add(1)
			continue
		}
		page, err := extractPage(item.Link, cssSelector, format)
		if err != nil {
			fmt.Printf("Error extracting page %s: %v\n", item.Link, err)
			bar.Add(1)
			continue
		}

		// Set description from the RSS feed
		page.Description = item.Description
		pages = append(pages, page)
		bar.Add(1) // Increment the progress bar
	}

	return pages, nil
}

// extractPage fetches a page and extracts its content based on a CSS selector and format.
func extractPage(pageURL, cssSelector, format string) (Page, error) {
	res, err := Client.Get(pageURL)
	if err != nil {
		return Page{}, fmt.Errorf("error visiting URL %s: %w", pageURL, err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return Page{}, fmt.Errorf("error parsing HTML from %s: %w", pageURL, err)
	}

	// Extract page details
	title := doc.Find("title").Text()
	description, _ := doc.Find("meta[name=description]").Attr("content")
	tags, _ := doc.Find("meta[name=tags]").Attr("content")

	var metaTags []string
	if tags != "" {
		metaTags = strings.Split(tags, ",")
	}

	// Convert relative URLs to absolute ones
	if hostDomain, err := getDomainFromURL(pageURL); err == nil {
		fixRelativeUrls(doc, hostDomain)
	}

	// Extract and transform content based on format
	content, err := extractAndTransformContent(doc, cssSelector, format)
	if err != nil {
		return Page{}, err
	}

	return Page{
		Title:       title,
		URL:         pageURL,
		Description: description,
		Tags:        metaTags,
		Content:     content,
	}, nil
}

// fixRelativeUrls converts relative URLs in links and images to absolute URLs.
func fixRelativeUrls(doc *goquery.Document, hostDomain string) {
	convertToAbsolute := func(attr, tag string) {
		doc.Find(tag).Each(func(i int, s *goquery.Selection) {
			url, exists := s.Attr(attr)
			if exists && isRelativeURL(url) && !isAnchorLink(url) {
				absoluteURL := toAbsoluteURL(hostDomain, url)
				s.SetAttr(attr, absoluteURL)
			}
		})
	}

	convertToAbsolute("href", "a")  // Convert relative links
	convertToAbsolute("src", "img") // Convert relative images
}

// extractAndTransformContent extracts content and applies HTML, Markdown, or Text transformations.
func extractAndTransformContent(doc *goquery.Document, cssSelector, format string) (string, error) {
	selection := doc.Find(cssSelector)
	if selection.Length() == 0 {
		return "", fmt.Errorf("CSS selector %s not found", cssSelector)
	}

	htmlContent, err := selection.Html()
	if err != nil {
		return "", fmt.Errorf("error extracting HTML: %w", err)
	}

	return extractAndTransformContentFromText(htmlContent, format)
}

// extractAndTransformContentFromText transforms content into HTML, Markdown, or plain text format.
func extractAndTransformContentFromText(content, format string) (string, error) {
	decodedContent := html.UnescapeString(content)
	sanitizedContent, _ := sanitize.HTMLAllowing(decodedContent, allowedTags, allowedAttributes)

	// Clean up excess newlines
	sanitizedContent = removeExcessNewlines(sanitizedContent)

	switch format {
	case "html":
		return sanitizedContent, nil
	case "md":
		converter := md.NewConverter("", true, nil)
		converter.Use(plugin.Table())
		mdContent, err := converter.ConvertString(sanitizedContent)
		if err != nil {
			return "", fmt.Errorf("error converting HTML to Markdown: %w", err)
		}
		return mdContent, nil
	case "txt":
		textContent, err := html2text.Convert(sanitizedContent)
		if err != nil {
			return "", fmt.Errorf("error converting HTML to text: %w", err)
		}
		return textContent, nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// getDomainFromURL extracts the scheme and host from the given URL.
func getDomainFromURL(pageURL string) (string, error) {
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	return fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host), nil
}

// isRelativeURL checks if the provided URL is relative.
func isRelativeURL(link string) bool {
	u, err := url.Parse(link)
	return err == nil && !u.IsAbs()
}

// toAbsoluteURL converts a relative URL to an absolute URL.
func toAbsoluteURL(host, relativeURL string) string {
	u, err := url.Parse(relativeURL)
	if err != nil {
		return relativeURL
	}

	baseURL, err := url.Parse(host)
	if err != nil {
		return relativeURL
	}

	return baseURL.ResolveReference(u).String()
}

// isAnchorLink checks if the given link is an in-page anchor (starts with "#").
func isAnchorLink(link string) bool {
	return len(link) > 0 && link[0] == '#'
}

// removeExcessNewlines normalizes line breaks and removes unnecessary newlines.
func removeExcessNewlines(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	content = collapseSpaces(content)
	return reExcessNewlines.ReplaceAllString(content, "\n\n")
}

// collapseSpaces reduces multiple spaces within text to a single space.
func collapseSpaces(content string) string {
	return reCollapseSpaces.ReplaceAllString(content, " ")
}

// matchesFilter checks if a given URL matches the provided filter pattern.
// The filter supports simple prefix matching with a trailing '*', e.g., "blog/*".
func matchesFilter(pageURL, filter string) bool {
	if filter == "" || filter == "*" {
		return true
	}
	parsed, err := url.Parse(pageURL)
	if err != nil {
		return false
	}
	filter = strings.TrimSuffix(filter, "*")
	path := strings.TrimPrefix(parsed.Path, "/")
	return strings.HasPrefix(path, filter)
}
