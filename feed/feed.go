package feed

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

// DetectFeedType detects whether the URL is an RSS feed or sitemap based on the XML root element.
func DetectFeedType(feedURL string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second, // Set a timeout to avoid long-running requests
	}

	// Fetch the feed URL
	res, err := client.Get(feedURL)
	if err != nil {
		return "", fmt.Errorf("error fetching URL %s: %w", feedURL, err)
	}
	defer res.Body.Close()

	// Check for non-200 status codes
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected HTTP status: %d %s", res.StatusCode, http.StatusText(res.StatusCode))
	}

	// Parse the XML root element to detect the feed type
	var root struct {
		XMLName xml.Name
	}
	decoder := xml.NewDecoder(res.Body)
	if err := decoder.Decode(&root); err != nil {
		return "", fmt.Errorf("error decoding XML from %s: %w", feedURL, err)
	}

	// Detect based on the root element's XML name
	switch root.XMLName.Local {
	case "urlset":
		return "sitemap", nil
	case "rss":
		return "rss", nil
	default:
		return "", fmt.Errorf("unknown feed type for URL %s", feedURL)
	}
}
