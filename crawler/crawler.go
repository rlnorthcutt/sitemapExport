package crawler

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"sitemapExport/extractor"

	"github.com/PuerkitoBio/goquery"
)

// Sitemap represents the structure of the XML sitemap.
type Sitemap struct {
	URLs []SitemapURL `xml:"url"`
}

// SitemapURL represents each <url> entry in the sitemap.
type SitemapURL struct {
	Loc string `xml:"loc"`
}

// Page represents the extracted data for a single page.
type Page struct {
	Title       string
	URL         string
	Description string
	MetaTags    []string
	Content     string
}

// CrawlSitemap fetches the sitemap, parses it, and crawls each page to extract content.
func CrawlSitemap(sitemapURL, cssSelector string) ([]Page, error) {
	var pages []Page

	// Fetch the sitemap
	res, err := http.Get(sitemapURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sitemap: %w", err)
	}
	defer res.Body.Close()

	// Read the sitemap XML
	sitemapData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read sitemap: %w", err)
	}

	// Parse the XML into Sitemap structure
	var sitemap Sitemap
	err = xml.Unmarshal(sitemapData, &sitemap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sitemap XML: %w", err)
	}

	// Loop through each URL in the sitemap and crawl the page
	for _, sitemapURL := range sitemap.URLs {
		page, err := extractPage(sitemapURL.Loc, cssSelector)
		if err != nil {
			fmt.Printf("Error extracting page %s: %v\n", sitemapURL.Loc, err)
			continue
		}
		pages = append(pages, page)
	}

	return pages, nil
}

// extractPage fetches the page at the given URL and extracts the content based on the CSS selector.
func extractPage(url, cssSelector string) (Page, error) {
	res, err := http.Get(url)
	if err != nil {
		return Page{}, fmt.Errorf("error visiting URL %s: %w", url, err)
	}
	defer res.Body.Close()

	// Parse the HTML document from the page
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

	// Extract content based on the provided CSS selector
	content, err := extractor.ExtractContent(doc, cssSelector)
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
