package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sitemapExport/crawler"
	"sitemapExport/feed"
	"sitemapExport/formatter"
	"sitemapExport/writer"
	"strings"

	"github.com/spf13/cobra"
)

var (
	feedURL        string
	cssSelector    string
	outputFilename string
	outputFiletype string
	format         string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		handleError("executing command", err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "sitemapExport",
	Short: "Crawl a sitemap or RSS feed and extract content.",
	Run:   executeCrawlAndExport, // Main function to run the command
}

func init() {
	// Define flags in the init function
	rootCmd.Flags().StringVarP(&feedURL, "url", "u", "", "Sitemap or RSS feed URL to crawl (required)")
	rootCmd.Flags().StringVarP(&cssSelector, "css", "c", "body", "CSS selector to extract content (for sitemaps)")
	rootCmd.Flags().StringVarP(&outputFilename, "filename", "n", "output", "Filename for the output")
	rootCmd.Flags().StringVarP(&outputFiletype, "type", "t", "txt", "File output format (txt, json, jsonl, md, pdf)")
	rootCmd.Flags().StringVarP(&format, "format", "f", "txt", "Content format transformation (html, md, txt)")
}

// executeCrawlAndExport prompts the user for missing input (if flags are not provided), validates the inputs, and runs the main export logic.
func executeCrawlAndExport(cmd *cobra.Command, args []string) {
	// Prompt for missing user input
	feedURL = promptUser("Enter the Sitemap or RSS feed URL (required): ", feedURL)
	if feedURL == "" {
		handleError("getting feed URL", fmt.Errorf("feed URL is required"))
	}

	cssSelector = promptUser("Enter the CSS selector to extract content (default: 'body'): ", cssSelector)
	outputFilename = promptUser("Enter the output filename (default: 'output'): ", outputFilename)

	// Validate output file type
	outputFiletype = promptUser("Enter the output file type (txt, json, jsonl, md, pdf) (default: 'txt'): ", outputFiletype)
	if !isValidOutputType(outputFiletype) {
		handleError("validating output file type", fmt.Errorf("unsupported output file type: %s", outputFiletype))
	}

	// Validate content format
	format = promptUser("Enter the content format (html, md, txt) (default: 'txt'): ", format)
	if !isValidFormat(format) {
		handleError("validating content format", fmt.Errorf("unsupported content format: %s", format))
	}

	// Confirm the input values with the user before proceeding
	fmt.Printf("\nExport data with the following settings:\n")
	fmt.Printf("URL: %s\n", feedURL)
	fmt.Printf("CSS Selector: %s\n", cssSelector)
	fmt.Printf("Output Filename: %s\n", outputFilename)
	fmt.Printf("Output Filetype: %s\n", outputFiletype)
	fmt.Printf("Format: %s\n", format)

	confirmation := promptUser("Do you want to proceed with these settings? (y/n): ", "y")
	if strings.ToLower(confirmation) != "y" {
		fmt.Println("Operation cancelled.")
		return
	}
	fmt.Print("\n")

	// Step 1: Detect if it's an RSS feed or a Sitemap
	feedType, err := feed.DetectFeedType(feedURL)
	handleError("detecting feed type", err)

	// Step 2: Fetch and crawl the pages based on the feed type
	var pages []crawler.Page
	switch feedType {
	case "rss":
		// Crawl RSS feed
		pages, err = crawler.CrawlRSS(feedURL, cssSelector, format)
		handleError("crawling RSS feed", err)
	case "sitemap":
		// Crawl Sitemap
		pages, err = crawler.CrawlSitemap(feedURL, cssSelector, format)
		handleError("crawling sitemap", err)
	default:
		handleError("processing feed", fmt.Errorf("unknown feed type detected"))
	}

	// Step 3: Format the extracted pages into the desired output file format
	formattedContent, err := formatter.FormatPages(pages, outputFiletype)
	handleError("formatting pages", err)

	// Step 4: Write the formatted content to the specified output file
	err = writer.WriteToFile(outputFilename, formattedContent, outputFiletype)
	handleError("writing to file", err)

	fmt.Printf("Successfully saved output to %s.%s\n", outputFilename, outputFiletype)
}

// promptUser is a helper function that asks for input, providing a default value if none is given.
func promptUser(message string, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	// If no input is provided, use the default value
	if input == "" {
		return defaultValue
	}
	return input
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
