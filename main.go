package main

import (
	"fmt"
	"log"
	"sitemapExport/crawler"
	"sitemapExport/feed"
	"sitemapExport/formatter"
	"sitemapExport/writer"

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
	rootCmd.Flags().StringVarP(&outputFilename, "outputName", "o", "output", "Filename for the output")
	rootCmd.Flags().StringVarP(&outputFiletype, "outputType", "t", "txt", "File output format (txt, json, jsonl, md, pdf)")
	rootCmd.Flags().StringVarP(&format, "format", "f", "txt", "Content format transformation (html, md, txt)")

	// @TODO: add checks for unsupported flag values to avoid running the command with invalid inputs

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func executeCrawlAndExport(cmd *cobra.Command, args []string) {
	if feedURL == "" {
		log.Fatal("You must provide a URL")
	}

	// Step 1: Detect if it's an RSS feed or a Sitemap
	feedType, err := feed.DetectFeedType(feedURL)
	if err != nil {
		log.Fatalf("Error detecting feed type: %v", err)
	}

	// Step 2: Fetch and crawl the pages based on the feed type
	var pages []crawler.Page

	switch feedType {
	case "rss":
		// Crawl RSS feed
		pages, err = crawler.CrawlRSS(feedURL, cssSelector, format)
		if err != nil {
			log.Fatalf("Error crawling RSS feed: %v", err)
		}
	case "sitemap":
		// Crawl Sitemap
		pages, err = crawler.CrawlSitemap(feedURL, cssSelector, format)
		if err != nil {
			log.Fatalf("Error crawling sitemap: %v", err)
		}
	default:
		log.Fatalf("Unknown feed type detected")
	}

	// Step 3: Format the extracted pages into the desired output file format
	formattedContent, err := formatter.FormatPages(pages, outputFiletype)
	if err != nil {
		log.Fatalf("Error formatting pages: %v", err)
	}

	// Step 4: Write the formatted content to the specified output file
	err = writer.WriteToFile(outputFilename, formattedContent, outputFiletype)
	if err != nil {
		log.Fatalf("Error writing to file: %v", err)
	}

	fmt.Printf("Successfully saved output to %s.%s\n", outputFilename, outputFiletype)
}
