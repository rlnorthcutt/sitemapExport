package writer

import (
	"fmt"
	"os"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

// WriteToFile handles writing formatted content to a file based on the selected format.
func WriteToFile(filename, content, format string) error {
	switch format {
	case "txt", "md", "json", "jsonl":
		return writeTextFile(filename+"."+format, content)
	case "pdf":
		return writePDF(filename+".pdf", content)
	default:
		return fmt.Errorf("unsupported file format: %s", format)
	}
}

// writeTextFile writes the content as a plain text, markdown, or JSON file.
func writeTextFile(filepath, content string) error {
	// Write the content to the file as plain text (or JSON or markdown).
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", filepath, err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("error writing to file %s: %w", filepath, err)
	}

	return nil
}

// writePDF generates a PDF file with the provided content.
func writePDF(outputPath string, content string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Sanitize content: remove unsupported characters or replace with a placeholder
	content = sanitizeText(content)

	// Add a page
	pdf.AddPage()

	// Set font and size
	pdf.SetFont("Arial", "", 12)

	// Split content into chunks that fit the PDF width
	width := 190.0 // Width in mm for A4 size PDF
	lines := pdf.SplitText(content, width)

	// Add each line to the PDF
	for _, line := range lines {
		pdf.Cell(0, 10, line)
		pdf.Ln(-1) // Line break
	}

	// Output to file
	err := pdf.OutputFileAndClose(outputPath)
	if err != nil {
		return fmt.Errorf("error writing PDF: %w", err)
	}

	return nil
}

// sanitizeText ensures that the content doesn't contain any problematic characters.
func sanitizeText(content string) string {
	// Remove non-ASCII characters (or replace with a placeholder)
	sanitized := strings.Map(func(r rune) rune {
		if r > 127 { // Non-ASCII characters
			return '?' // Or return -1 to remove the character
		}
		return r
	}, content)

	return sanitized
}
