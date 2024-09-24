# sitemapExport

`sitemapExport` is a Go-based CLI tool that crawls a sitemap, extracts content from web pages using CSS selectors, and compiles the data into various formats such as `txt`, `json`, `jsonl`, `md`, and `pdf`.

## Features

- Crawl a sitemap and extract content from pages.
- Extract page content using a specified CSS selector.
- Generate a structured list of pages with:
  - Page title
  - URL
  - Meta description (if available)
  - Meta tags (if available)
  - Extracted content
- Output formats supported:
  - Plain text (`txt`)
  - JSON (`json`)
  - JSON Lines (`jsonl`)
  - Markdown (`md`)
  - PDF (`pdf`)
  
## Installation

### Use the generated binary

### Build from source
To build `sitemapExport`, you'll need [Go](https://golang.org/doc/install) installed.

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/sitemapExport.git
   cd sitemapExport
   ```

2. Build the CLI tool:
   ```bash
   go build
   ```
   However, we can create a smaller binary with this command:
   ```bash
   go build -ldflags="-s -w"
  ```

   This will generate the `sitemapExport` binary.

## Usage

Once built, you can run the tool from the command line. The tool will prompt you for input values interactively:

```bash
./sitemapExport
```

### Example Interactive Prompts

```bash
$ ./sitemapExport
Enter the sitemap URL (required): https://example.com/sitemap.xml
Enter the CSS selector to extract content (default: body):
Enter the output filename (default: sitemap): output
Enter the output format (txt, json, jsonl, md, pdf) [default: txt]: jsonl
Successfully saved output to output.jsonl
```

This will crawl the provided sitemap, extract content from each page using the CSS selector, and save the output to the chosen format (`jsonl` in this case).

### Command-Line Options (Not Interactive)
If you prefer to pass flags instead of interactive prompts, you can run:

```bash
./sitemapExport --sitemap="https://example.com/sitemap.xml" --css="body" --output="output" --format="txt"
```

### Supported Formats

- `txt`: Plain text format
- `json`: JSON with pretty-printing
- `jsonl`: JSON Lines format (one JSON object per line)
- `md`: Markdown format
- `pdf`: PDF format

## Example Output

**JSON Output (`output.json`)**:
```json
[
  {
    "Title": "Home",
    "URL": "https://example.com",
    "Description": "Welcome to our homepage",
    "MetaTags": ["name: description, content: Welcome to our site"],
    "Content": "<div>Welcome to our site!</div>"
  },
  {
    "Title": "About Us",
    "URL": "https://example.com/about",
    "Description": "Learn more about our company",
    "MetaTags": ["name: description, content: About Us"],
    "Content": "<p>We are a company...</p>"
  }
]
```

**Markdown Output (`output.md`)**:
```markdown
# Home

URL: https://example.com

Description: Welcome to our homepage

<div>Welcome to our site!</div>

---

# About Us

URL: https://example.com/about

Description: Learn more about our company

<p>We are a company...</p>

---
```

**JSON Lines Output (`output.jsonl`)**:
```jsonl
{"Title":"Home","URL":"https://example.com/","Description":"Welcome to our homepage","MetaTags":["name: description, content: Welcome"],"Content":"<div>Welcome to our site!</div>"}
{"Title":"About Us","URL":"https://example.com/about","Description":"Learn more about our company","MetaTags":["name: description, content: About Us"],"Content":"<p>We are a company...</p>"}
```

## Project Structure

```bash
sitemapExport/
├── main.go           # CLI entry point
├── crawler/          # Handles sitemap crawling and page extraction
│   └── crawler.go
├── extractor/        # Extracts HTML content based on CSS selectors
│   └── extractor.go
├── formatter/        # Formats extracted content into different file formats
│   └── formatter.go
├── writer/           # Writes formatted content to files (txt, json, md, pdf)
│   └── writer.go
├── go.mod            # Go module file with dependencies
├── go.sum            # Go module dependency checksum
└── README.md         # Project documentation
```

## Dependencies

`sitemapExport` uses the following Go packages:

- [`github.com/PuerkitoBio/goquery`](https://github.com/PuerkitoBio/goquery) - For parsing and manipulating HTML documents.
- [`github.com/gocolly/colly/v2`](https://github.com/gocolly/colly) - For crawling the sitemap.
- [`github.com/kennygrant/sanitize`](https://github.com/kennygrant/sanitize) - For sanitizing HTML content.
- [`github.com/spf13/cobra`](https://github.com/spf13/cobra) - For CLI command management.
- [`github.com/jung-kurt/gofpdf`](https://github.com/jung-kurt/gofpdf) - For PDF generation.
- [`github.com/russross/blackfriday/v2`](https://github.com/russross/blackfriday) - For converting HTML to Markdown.

## Contributing

Feel free to open issues or submit pull requests for new features, bug fixes, or general improvements.

## License

This project is licensed under the MIT License.
