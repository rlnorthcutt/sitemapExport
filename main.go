package main

import (
	"fmt"
	"log"
	"sitemapExport/crawler"
	"sitemapExport/formatter"
	"sitemapExport/writer"

	"github.com/spf13/cobra"
)

var sitemapURL, cssSelector, outputFilename, outputFiletype, format string

func main() {
	rootCmd := &cobra.Command{
		Use:   "sitemapExport",
		Short: "Crawl a sitemap and extract content from the pages.",
		Run:   executeCrawlAndExport,
	}

	// Define flags
	rootCmd.Flags().StringVarP(&sitemapURL, "sitemap", "s", "", "Sitemap URL to crawl (required)")
	rootCmd.Flags().StringVarP(&cssSelector, "css", "c", "body", "CSS selector to extract content")
	rootCmd.Flags().StringVarP(&outputFilename, "outputName", "o", "sitemap", "Filename for the output")
	rootCmd.Flags().StringVarP(&outputFiletype, "outputType", "t", "txt", "File output format (txt, json, jsonl, md, pdf)")
	rootCmd.Flags().StringVarP(&format, "format", "f", "txt", "Content format transformation (html, md, txt)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func executeCrawlAndExport(cmd *cobra.Command, args []string) {
	// Crawl the sitemap and extract pages
	pages, err := crawler.CrawlSitemap(sitemapURL, cssSelector, format)
	if err != nil {
		log.Fatalf("Error crawling sitemap: %v", err)
	}

	// Format the extracted pages into the desired output file format
	formattedContent, err := formatter.FormatPages(pages, outputFiletype)
	if err != nil {
		log.Fatalf("Error formatting pages: %v", err)
	}

	// Write the formatted content to the specified output file
	err = writer.WriteToFile(outputFilename, formattedContent, outputFiletype)
	if err != nil {
		log.Fatalf("Error writing to file: %v", err)
	}

	fmt.Printf("Successfully saved output to %s.%s\n", outputFilename, outputFiletype)
}
