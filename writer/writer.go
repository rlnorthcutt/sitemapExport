package writer

import (
	"fmt"
	"os"

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

// writePDF creates a PDF file from the content.
func writePDF(filepath, content string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)

	// Break the content into lines to fit into the PDF.
	lines := pdf.SplitText(content, 190) // 190mm width for A4 page

	for _, line := range lines {
		pdf.Cell(0, 10, line)
		pdf.Ln(10) // Newline after each cell
	}

	err := pdf.OutputFileAndClose(filepath)
	if err != nil {
		return fmt.Errorf("error writing PDF file %s: %w", filepath, err)
	}

	return nil
}
