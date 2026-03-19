package writer

import (
	"fmt"
	"os"
)

// WriteToFile writes formatted content to a file based on the selected format.
func WriteToFile(filename, content, format string) error {
	filepath := filename + "." + format

	switch format {
	case "txt", "md", "json", "jsonl":
		return writeTextFile(filepath, content)
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
