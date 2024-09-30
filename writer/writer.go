package writer

import (
	"fmt"
	"os"
	"strings"

	"github.com/jung-kurt/gofpdf"
)

// WriteToFile writes formatted content to a file based on the selected format.
func WriteToFile(filename, content, format string) error {
	filepath := filename + "." + format

	switch format {
	case "txt", "md", "json", "jsonl":
		return writeTextFile(filepath, content)
	case "pdf":
		return writePDF(filepath, content)
	default:
		return fmt.Errorf("unsupported file format: %s", format)
	}
}

// writeTextFile writes content as plain text, markdown, or JSON file.
func writeTextFile(filepath, content string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", filepath, err)
	}
	defer file.Close()

	if _, err = file.WriteString(content); err != nil {
		return fmt.Errorf("error writing to file %s: %w", filepath, err)
	}

	return nil
}

// writePDF generates a PDF file with the provided content.
func writePDF(filepath, content string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Add a page and handle potential errors
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "", 12)

	// Sanitize content by removing unsupported characters
	content = sanitizeText(content)

	// Define PDF layout width for content fitting
	width := 190.0 // Effective content width for A4 paper in mm

	// Split content into lines that fit within the width
	lines := pdf.SplitText(content, width)

	// Add each line to the PDF
	for _, line := range lines {
		pdf.Cell(0, 10, line)
		pdf.Ln(-1) // Line break
	}

	// Output the PDF to file and handle errors
	if err := pdf.OutputFileAndClose(filepath); err != nil {
		return fmt.Errorf("error writing PDF file: %w", err)
	}

	return nil
}

// sanitizeText removes or replaces non-ASCII characters from the content.
func sanitizeText(content string) string {
	// Replace non-ASCII characters with '?' or remove them
	return strings.Map(func(r rune) rune {
		if r > 127 {
			return -1 // Return -1 if you want to remove non-ASCII characters
		}
		return r
	}, content)
}
