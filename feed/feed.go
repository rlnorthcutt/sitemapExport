package feed

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

// DetectFeedType detects whether the URL is an RSS feed or sitemap.
func DetectFeedType(feedURL string) (string, error) {
	res, err := http.Get(feedURL)
	if err != nil {
		return "", fmt.Errorf("error fetching URL: %w", err)
	}
	defer res.Body.Close()

	var root struct {
		XMLName xml.Name
	}
	decoder := xml.NewDecoder(res.Body)
	if err := decoder.Decode(&root); err != nil {
		return "", fmt.Errorf("error decoding XML: %w", err)
	}

	switch root.XMLName.Local {
	case "urlset":
		return "sitemap", nil
	case "rss":
		return "rss", nil
	default:
		return "", fmt.Errorf("unknown feed type")
	}
}
