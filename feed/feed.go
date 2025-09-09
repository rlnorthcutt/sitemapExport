package feed

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"sitemapExport/crawler"
)

// DetectFeedType detects whether the provided source is an RSS feed or sitemap
// based on the XML root element. The source can be either a URL or a file path.
func DetectFeedType(feedSource string) (string, error) {
	var (
		reader io.ReadCloser
		err    error
	)

	if strings.HasPrefix(feedSource, "http://") || strings.HasPrefix(feedSource, "https://") {
		// Fetch the feed URL
		res, err := crawler.Client.Get(feedSource)
		if err != nil {
			return "", fmt.Errorf("error fetching URL %s: %w", feedSource, err)
		}
		if res.StatusCode != http.StatusOK {
			res.Body.Close()
			return "", fmt.Errorf("unexpected HTTP status: %d %s", res.StatusCode, http.StatusText(res.StatusCode))
		}
		reader = res.Body
	} else {
		// Open local file
		reader, err = os.Open(feedSource)
		if err != nil {
			return "", fmt.Errorf("error opening file %s: %w", feedSource, err)
		}
	}
	defer reader.Close()

	// Parse the XML root element to detect the feed type
	var root struct {
		XMLName xml.Name
	}
	decoder := xml.NewDecoder(reader)
	if err := decoder.Decode(&root); err != nil {
		return "", fmt.Errorf("error decoding XML from %s: %w", feedSource, err)
	}

	// Detect based on the root element's XML name
	switch root.XMLName.Local {
	case "urlset":
		return "sitemap", nil
	case "rss":
		return "rss", nil
	default:
		return "", fmt.Errorf("unknown feed type for %s", feedSource)
	}
}
