package main

import (
	"fmt"
	"strings"

	"github.com/rlnorthcutt/cmdkit/logger"
	"github.com/rlnorthcutt/cmdkit/ui"
	"github.com/spf13/pflag"
	"sitemapExport/crawler"
	"sitemapExport/feed"
	"sitemapExport/formatter"
	"sitemapExport/writer"
)

var (
	feedSource     string
	cssSelector    string
	outputFilename string
	outputFiletype string
	urlFilter      string
	verbose        bool
)

func main() {
	pflag.StringVarP(&feedSource, "input", "i", "", "Sitemap or RSS feed URL or file path to crawl (required)")
	pflag.StringVarP(&cssSelector, "css", "c", "body", "CSS selector to extract content (for sitemaps)")
	pflag.StringVarP(&outputFilename, "filename", "n", "output", "Filename for the output")
	pflag.StringVarP(&outputFiletype, "type", "t", "txt", "File output format (txt, json, jsonl, md)")
	pflag.StringVar(&urlFilter, "filter", "*", "Only include URLs matching this pattern (e.g., blog/*)")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	pflag.Parse()

	executeCrawlAndExport()
}

// executeCrawlAndExport resolves all inputs (via flags, env vars, or interactive prompts),
// confirms settings with the user, then runs the full detect → crawl → format → write pipeline.
func executeCrawlAndExport() {
	log := logger.New(verbose)
	u := ui.New(false).WithLogger(log)

	// Resolve all inputs: flag > env > prompt > default
	u.ResolveString(feedSource, pflag.CommandLine.Changed("input"), "SITEMAP_INPUT", "Enter the Sitemap or RSS feed URL or file path (required)", &feedSource)
	if feedSource == "" {
		log.Fatal("error getting feed source: feed source is required")
	}

	u.ResolveString(cssSelector, pflag.CommandLine.Changed("css"), "SITEMAP_CSS", "Enter the CSS selector to extract content", &cssSelector)
	u.ResolveString(outputFilename, pflag.CommandLine.Changed("filename"), "SITEMAP_FILENAME", "Enter the output filename", &outputFilename)
	u.ResolveString(urlFilter, pflag.CommandLine.Changed("filter"), "SITEMAP_FILTER", "Enter the URL filter pattern", &urlFilter)
	u.ResolveString(outputFiletype, pflag.CommandLine.Changed("type"), "SITEMAP_TYPE", "Enter the output file type (txt, json, jsonl, md)", &outputFiletype)

	if !isValidOutputType(outputFiletype) {
		log.Fatal("error validating output file type: unsupported type: %s", outputFiletype)
	}

	// Automatically determine content format from output type
	format := "txt"
	if strings.EqualFold(outputFiletype, "md") {
		format = "md"
	}

	// Confirm the input values with the user before proceeding
	log.Print("\nExport data with the following settings:")
	log.Print("  Input:           %s", feedSource)
	log.Print("  CSS Selector:    %s", cssSelector)
	log.Print("  URL Filter:      %s", urlFilter)
	log.Print("  Output Filename: %s", outputFilename)
	log.Print("  Output Filetype: %s", outputFiletype)
	log.Print("  Format:          %s\n", format)

	if !u.Confirm("Do you want to proceed with these settings?") {
		log.Print("Operation cancelled.")
		return
	}
	fmt.Println()

	// Step 1: Detect if it's an RSS feed or a Sitemap
	feedType, err := feed.DetectFeedType(feedSource)
	if err != nil {
		log.Fatal("error detecting feed type: %v", err)
	}

	// Step 2: Fetch and crawl the pages based on the feed type
	var pages []crawler.Page
	switch feedType {
	case "rss":
		pages, err = crawler.CrawlRSS(feedSource, cssSelector, format, urlFilter)
		if err != nil {
			log.Fatal("error crawling RSS feed: %v", err)
		}
	case "sitemap":
		pages, err = crawler.CrawlSitemap(feedSource, cssSelector, format, urlFilter)
		if err != nil {
			log.Fatal("error crawling sitemap: %v", err)
		}
	default:
		log.Fatal("error processing feed: unknown feed type detected")
	}

	// Step 3: Format the extracted pages into the desired output file format
	formattedContent, err := formatter.FormatPages(pages, outputFiletype)
	if err != nil {
		log.Fatal("error formatting pages: %v", err)
	}

	// Step 4: Write the formatted content to the specified output file
	err = writer.WriteToFile(outputFilename, formattedContent, outputFiletype)
	if err != nil {
		log.Fatal("error writing to file: %v", err)
	}

	log.Success("Successfully saved output to %s.%s", outputFilename, outputFiletype)
}

// isValidOutputType checks if the provided output filetype is supported.
func isValidOutputType(outputType string) bool {
	supportedTypes := []string{"txt", "json", "jsonl", "md"}
	for _, t := range supportedTypes {
		if strings.EqualFold(t, outputType) {
			return true
		}
	}
	return false
}
