package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sitemapExport/crawler"
	"sitemapExport/formatter"
	"sitemapExport/writer"
	"strings"

	"github.com/spf13/cobra"
)

var sitemapURL, cssSelector, outputFilename, format string

func main() {
	rootCmd := &cobra.Command{
		Use:   "sitemapExport",
		Short: "Crawl a sitemap and extract content from the pages.",
		Run:   executeCrawlAndExport, // Function to execute
	}

	// Define flags
	rootCmd.Flags().StringVarP(&sitemapURL, "sitemap", "s", "", "Sitemap URL to crawl (required)")
	rootCmd.Flags().StringVarP(&cssSelector, "css", "c", "", "CSS selector to extract content (default: body)")
	rootCmd.Flags().StringVarP(&outputFilename, "output", "o", "", "Filename for the output (default: sitemap)")
	rootCmd.Flags().StringVarP(&format, "format", "f", "", "File output format (txt, json, jsonl, md, pdf) (default: txt)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// executeCrawlAndExport prompts the user for missing input (if flags are not provided) and runs the main export logic.
func executeCrawlAndExport(cmd *cobra.Command, args []string) {
	// Prompt for sitemap URL if not provided as a flag
	if sitemapURL == "" {
		sitemapURL = promptUser("Enter the sitemap URL (required): ")
		if sitemapURL == "" {
			log.Fatal("Sitemap URL is required.")
		}
	}

	// Default the CSS selector to "body" if not provided as a flag
	if cssSelector == "" {
		cssSelector = promptUser("Enter the CSS selector to extract content (default: body): ")
		if cssSelector == "" {
			cssSelector = "body"
		}
	}

	// Prompt for output filename if not provided as a flag
	if outputFilename == "" {
		outputFilename = promptUser("Enter the output filename (default: sitemap): ")
		if outputFilename == "" {
			outputFilename = "sitemap" // Default value
		}
	}

	// Prompt for format if not provided as a flag
	if format == "" {
		format = promptUser("Enter the output format (txt, json, jsonl, md, pdf) [default: txt]: ")
		if format == "" {
			format = "txt" // Default value
		}
	}

	// Crawl the sitemap and extract pages
	pages, err := crawler.CrawlSitemap(sitemapURL, cssSelector)
	if err != nil {
		log.Fatalf("Error crawling sitemap: %v", err)
	}

	// Format the extracted pages into the desired output format
	formattedContent, err := formatter.FormatPages(pages, format)
	if err != nil {
		log.Fatalf("Error formatting pages: %v", err)
	}

	// Write the formatted content to the specified output file
	err = writer.WriteToFile(outputFilename, formattedContent, format)
	if err != nil {
		log.Fatalf("Error writing to file: %v", err)
	}

	fmt.Printf("Successfully saved output to %s.%s\n", outputFilename, format)
}

// promptUser is a helper function to ask the user for input interactively if a flag is not provided.
func promptUser(message string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
