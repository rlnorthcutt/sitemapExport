package main

import (
	"fmt"
	"log"
	"sitemapExport/crawler"
	"sitemapExport/feed"
	"sitemapExport/formatter"
	"sitemapExport/writer"
	"strings"

	"github.com/spf13/cobra"
)

var (
	feedURL, cssSelector, outputFilename, outputFiletype, format string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "sitemapExport",
		Short: "Crawl a sitemap or RSS feed and extract content.",
		Run:   executeCrawlAndExport,
	}

	// Define flags
	rootCmd.Flags().StringVarP(&feedURL, "url", "u", "", "Sitemap or RSS feed URL to crawl (required)")
	rootCmd.Flags().StringVarP(&cssSelector, "css", "c", "body", "CSS selector to extract content (for sitemaps)")
	rootCmd.Flags().StringVarP(&outputFilename, "outputName", "o", "output", "Filename for the output (without extension)")
	rootCmd.Flags().StringVarP(&outputFiletype, "outputType", "t", "txt", "File output format (txt, json, jsonl, md, pdf)")
	rootCmd.Flags().StringVarP(&format, "format", "f", "txt", "Content format transformation (html, md, txt)")

	// Validate inputs before execution
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// executeCrawlAndExport performs the entire crawling, formatting, and writing process.
func executeCrawlAndExport(cmd *cobra.Command, args []string) {
	// Validate required flags
	if feedURL == "" {
		log.Fatal("Error: You must provide a URL with --url or -u")
	}

	// Validate supported output formats
	if !isValidOutputType(outputFiletype) {
		log.Fatalf("Error: Unsupported output file type '%s'. Supported types: txt, json, jsonl, md, pdf.", outputFiletype)
	}

	// Validate content format transformation
	if !isValidFormat(format) {
		log.Fatalf("Error: Unsupported content format '%s'. Supported formats: html, md, txt.", format)
	}

	// Step 1: Detect feed type (RSS or Sitemap)
	feedType, err := feed.DetectFeedType(feedURL)
	handleError("detecting feed type", err)

	// Step 2: Crawl the pages based on feed type
	var pages []crawler.Page
	switch feedType {
	case "rss":
		pages, err = crawler.CrawlRSS(feedURL, cssSelector, format)
	case "sitemap":
		pages, err = crawler.CrawlSitemap(feedURL, cssSelector, format)
	default:
		log.Fatal("Error: Unknown feed type detected.")
	}
	handleError("crawling the feed", err)

	// Check if any pages were extracted
	if len(pages) == 0 {
		log.Fatal("Error: No pages found in the feed.")
	}

	// Step 3: Format pages into the desired output format
	formattedContent, err := formatter.FormatPages(pages, outputFiletype)
	handleError("formatting pages", err)

	// Step 4: Automatically append the correct file extension
	outputPath := fmt.Sprintf("%s.%s", outputFilename, outputFiletype)

	// Step 5: Write the formatted content to the output file
	err = writer.WriteToFile(outputPath, formattedContent, outputFiletype)
	handleError("writing to file", err)

	fmt.Printf("\n\nSuccessfully saved output to %s\n", outputPath)
}

// handleError logs and terminates if an error occurs during a specific step.
func handleError(step string, err error) {
	if err != nil {
		log.Fatalf("Error %s: %v", step, err)
	}
}

// isValidOutputType checks if the provided output filetype is supported.
func isValidOutputType(outputType string) bool {
	supportedTypes := []string{"txt", "json", "jsonl", "md", "pdf"}
	for _, t := range supportedTypes {
		if strings.EqualFold(t, outputType) {
			return true
		}
	}
	return false
}

// isValidFormat checks if the provided content format transformation is supported.
func isValidFormat(format string) bool {
	supportedFormats := []string{"html", "md", "txt"}
	for _, f := range supportedFormats {
		if strings.EqualFold(f, format) {
			return true
		}
	}
	return false
}
