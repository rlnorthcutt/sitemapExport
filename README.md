# sitemapExport

`sitemapExport` is a Go-based CLI tool that crawls a sitemap or RSS feed, extracts content from web pages using CSS selectors, and compiles the data into various formats such as `txt`, `json`, `jsonl`, and `md`.

The primary use case is to extract content into a file that can be used as contextual data for AI. For example, extracting your docs site as a structured text or JSON file to power a solid AI support chatbot ([tutorial here](https://community.appsmith.com/tutorial/4-easy-steps-build-ai-powered-support-bot-knows-your-docs)).

## Features

- Crawl a sitemap or RSS feed to extract content from pages.
- Extract page content using a specified CSS selector.
- Optionally filter URLs with a simple pattern (e.g., `blog/*`).
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
- Supports non-interactive usage via flags or environment variables.

## Installation

### Easy: Run the binary

Grab the pre-built [`sitemapExport`](https://github.com/rlnorthcutt/sitemapExport/releases/) binary from the releases page, make it executable, and run it.

### Build from source

Requires [Go](https://golang.org/doc/install) 1.23 or later.

1. Clone the repository:
   ```bash
   git clone https://github.com/rlnorthcutt/sitemapExport.git
   cd sitemapExport
   ```

2. Build the binary:
   ```bash
   go build
   ```
   For a smaller binary:
   ```bash
   go build -ldflags="-s -w"
   ```

## Usage

The tool supports interactive prompts, command-line flags, and environment variables. Flags take priority over environment variables, which take priority over interactive prompts.

### Interactive

```bash
./sitemapExport
```

Example session:

```
Enter the Sitemap or RSS feed URL or file path (required): https://example.com/sitemap.xml
Enter the CSS selector to extract content (default: body):
Enter the output filename (default: output):
Enter the URL filter pattern (default: *):
Enter the output file type (txt, json, jsonl, md) (default: txt): jsonl

Export data with the following settings:
  Input:           https://example.com/sitemap.xml
  CSS Selector:    body
  URL Filter:      *
  Output Filename: output
  Output Filetype: jsonl
  Format:          txt

Do you want to proceed with these settings? (y/n): y
```

### Non-interactive (flags)

```bash
./sitemapExport --input="https://example.com/sitemap.xml" --css="article" --filename="output" --type="jsonl" --filter="blog/*"
```

Short flags are also supported:

```bash
./sitemapExport -i "https://example.com/sitemap.xml" -c "article" -n "output" -t "jsonl"
```

### Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--input` | `-i` | _(required)_ | Sitemap or RSS feed URL, or local file path |
| `--css` | `-c` | `body` | CSS selector to extract content |
| `--filename` | `-n` | `output` | Output filename (without extension) |
| `--type` | `-t` | `txt` | Output format: `txt`, `json`, `jsonl`, `md` |
| `--filter` | | `*` | Only include URLs matching this pattern (e.g., `blog/*`) |
| `--verbose` | `-v` | `false` | Enable verbose output |

### Environment variables

All inputs can also be set via environment variables, which are checked when a flag is not explicitly provided:

| Variable | Corresponding flag |
|---|---|
| `SITEMAP_INPUT` | `--input` |
| `SITEMAP_CSS` | `--css` |
| `SITEMAP_FILENAME` | `--filename` |
| `SITEMAP_TYPE` | `--type` |
| `SITEMAP_FILTER` | `--filter` |

### Output formats

- `txt` — plain text, one page per section
- `json` — pretty-printed JSON array
- `jsonl` — one JSON object per line (ideal for AI ingestion pipelines)
- `md` — Markdown with headings and content

### Example output

**JSON (`output.json`)**:
```json
[
  {
    "Title": "Home",
    "URL": "https://example.com",
    "Description": "Welcome to our homepage",
    "Content": "Welcome to our site!"
  },
  {
    "Title": "About Us",
    "URL": "https://example.com/about",
    "Description": "Learn more about our company",
    "Content": "We are a company..."
  }
]
```

**JSON Lines (`output.jsonl`)**:
```
{"Title":"Home","URL":"https://example.com/","Description":"Welcome to our homepage","Content":"Welcome to our site!"}
{"Title":"About Us","URL":"https://example.com/about","Description":"Learn more about our company","Content":"We are a company..."}
```

## Project structure

```
sitemapExport/
├── main.go           # CLI entry point, flag definitions, orchestration
├── crawler/          # Sitemap and RSS crawling, page extraction
│   └── crawler.go
├── formatter/        # Formats extracted pages into output formats
│   └── formatter.go
├── writer/           # Writes formatted content to disk
│   └── writer.go
├── feed/             # Feed type detection (sitemap vs RSS)
│   └── feed.go
├── html2text/        # HTML to plain text conversion
│   └── html2text.go
├── go.mod
└── go.sum
```

## Dependencies

- [`github.com/PuerkitoBio/goquery`](https://github.com/PuerkitoBio/goquery) — HTML parsing and CSS selector queries
- [`github.com/JohannesKaufmann/html-to-markdown`](https://github.com/JohannesKaufmann/html-to-markdown) — HTML to Markdown conversion
- [`github.com/spf13/pflag`](https://github.com/spf13/pflag) — CLI flag parsing with short and long flag support
- [`fortio.org/progressbar`](https://github.com/fortio/progressbar) — Zero-dependency progress bar
- [`github.com/rlnorthcutt/cmdkit`](https://github.com/rlnorthcutt/cmdkit) — Interactive prompts, env var resolution, and colored logging

## Contributing

Feel free to open issues or submit pull requests for new features, bug fixes, or general improvements.

## License

This project is licensed under the MIT License.
