# sitemapExport

`sitemapExport` is a Go-based CLI tool that crawls a sitemap or RSS feed, extracts content from web pages using CSS selectors, and compiles the data into various formats such as `txt`, `json`, `jsonl`, `md`, and `pdf`.

The primary use case is to extract content into a file that can be used as contextual data for AI. For example, extracting your docs site as a simple PDF to power a solid AI support chatbot ([tutorial here](https://community.appsmith.com/tutorial/4-easy-steps-build-ai-powered-support-bot-knows-your-docs)).

## Features

- Crawl a sitemap or RSS feed to extract content from pages.
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

### Easy : Run the command
1. Just grab the [`sitemapExport`](https://github.com/rlnorthcutt/sitemapExport/releases/) file in the repo.

2. Make the file executable, and run

Thats it. Each time the repo is updated, the executable is rebuilt. However, you can always build it from source if you choose.

### Fun : Build from source
To build `sitemapExport`, you'll need [Go](https://golang.org/doc/install) installed.

1. Clone the repository:
   ```bash
   git clone https://github.com/rlnorthcutt/sitemapExport.git
   cd sitemapExport
   ```

2. Build the CLI tool:
   ```bash
   go build
   ```
   For a smaller binary, you can use:
   ```bash
   go build -ldflags="-s -w"
   ```

   This will generate the `sitemapExport` binary.

## Usage

Once built, you can run the tool from the command line. The tool supports both interactive prompts and command-line flags.

### Interactive Usage

```bash
./sitemapExport
```

### Example Interactive Prompts

```bash
$ ./sitemapExport
Enter the Sitemap or RSS feed URL or file path (required): https://example.com/sitemap.xml
Enter the CSS selector to extract content (default: body):
Enter the output filename (default: output): output
Enter the output file type (txt, json, jsonl, md, pdf) (default: txt): jsonl
Enter the content format (html, md, txt) (default: txt): md
Successfully saved output to output.jsonl
```

This will crawl the provided sitemap, extract content from each page using the CSS selector, and save the output in the chosen format (`jsonl` in this case).

### Command-Line Options (Non-Interactive)
If you prefer to pass flags instead of interactive prompts, you can run:

```bash
./sitemapExport --input="https://example.com/sitemap.xml" --css="body" --filename="output" --type="txt" --format="txt"
```

Or, use the short flags:

```bash
./sitemapExport -i="https://example.com/sitemap.xml" -c="body" -n="output" -t="txt" -f="txt"
```

### Supported Formats

- `txt`: Plain text format
- `json`: JSON with pretty-printing
- `jsonl`: JSON Lines format (one JSON object per line)
- `md`: Markdown format
- `pdf`: PDF format

### Example Output

**JSON Output (`output.json`)**:
```json
[
  {
    "Title": "Home",
    "URL": "https://example.com",
    "Description": "Welcome to our homepage",
    "MetaTags": ["description: Welcome to our site"],
    "Content": "<div>Welcome to our site!</div>"
  },
  {
    "Title": "About Us",
    "URL": "https://example.com/about",
    "Description": "Learn more about our company",
    "MetaTags": ["description: About Us"],
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
{"Title":"Home","URL":"https://example.com/","Description":"Welcome to our homepage","MetaTags":["description: Welcome"],"Content":"<div>Welcome to our site!</div>"}
{"Title":"About Us","URL":"https://example.com/about","Description":"Learn more about our company","MetaTags":["description: About Us"],"Content":"<p>We are a company...</p>"}
```

## Project Structure

```bash
sitemapExport/
├── main.go           # CLI entry point
├── crawler/          # Handles sitemap and RSS crawling and page extraction
│   └── crawler.go
├── formatter/        # Formats extracted content into different file formats
│   └── formatter.go
├── writer/           # Writes formatted content to files (txt, json, md, pdf)
│   └── writer.go
├── feed/             # Detects feed type and handles feed-related tasks
│   └── feed.go
├── go.mod            # Go module file with dependencies
├── go.sum            # Go module dependency checksum
└── README.md         # Project documentation
```

## Dependencies

`sitemapExport` uses the following Go packages:

- [`github.com/PuerkitoBio/goquery`](https://github.com/PuerkitoBio/goquery) - For parsing and manipulating HTML documents.
- [`github.com/kennygrant/sanitize`](https://github.com/kennygrant/sanitize) - For sanitizing HTML content.
- [`github.com/spf13/cobra`](https://github.com/spf13/cobra) - For CLI command management.
- [`github.com/jung-kurt/gofpdf`](https://github.com/jung-kurt/gofpdf) - For PDF generation.
- [`github.com/JohannesKaufmann/html-to-markdown`](https://github.com/JohannesKaufmann/html-to-markdown) - For converting HTML to Markdown.
- [`github.com/schollz/progressbar/v3`](https://github.com/schollz/progressbar) - For showing progress bars during sitemap and RSS crawling.

## Contributing

Feel free to open issues or submit pull requests for new features, bug fixes, or general improvements.

## License

This project is licensed under the MIT License.

