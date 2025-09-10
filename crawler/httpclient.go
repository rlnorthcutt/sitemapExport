package crawler

import (
	"net/http"
	"time"
)

var (
	// Client is the shared HTTP client used by crawler and feed packages.
	Client    = &http.Client{Timeout: 10 * time.Second}
	userAgent = "sitemapExport"
)

type userAgentTransport struct {
	base http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	return t.base.RoundTrip(req)
}

func init() {
	Client.Transport = &userAgentTransport{base: http.DefaultTransport}
}

// SetUserAgent allows overriding the User-Agent header for HTTP requests.
func SetUserAgent(ua string) {
	if ua != "" {
		userAgent = ua
	}
}
