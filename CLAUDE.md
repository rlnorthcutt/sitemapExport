## Purpose

- **Goal**: Help build and maintain a Go-based CLI tool (similar to `sitemapExport`) that crawls a sitemap/RSS feed or other structured sources, extracts content, and exports it into multiple formats for AI-friendly consumption.
- **Behavior**: Prefer small, focused changes that match the existing architecture and style. When in doubt, mirror patterns already used in this repo.

## Development environment

- **Dev container**:
  - Assume development happens inside a VS Code / Cursor dev container based on the official Go image (see `.devcontainer/devcontainer.json`).
  - **Do**: Keep changes compatible with the Go version in the image (currently Go 1.23 image tag).
  - **If adding tools** (linters, generators, etc.): Prefer installing them via Go modules or the dev container image rather than shelling out to the host.

- **Dependencies**:
  - **Respect** existing module management via `go.mod`/`go.sum`.
  - **When adding libraries**: 
    - Use `go get` and let `go mod tidy` manage versions.
    - Favor small, well-maintained libraries over large frameworks.

## Architecture and code organization

- **Overall structure**:
  - **Keep a single CLI entry point** in `main.go` that wires together:
    - feed detection (`feed` package),
    - crawling & extraction (`crawler`),
    - formatting (`formatter`),
    - file output (`writer`),
    - helper utilities (`html2text` and similar).
  - **Do** split new responsibilities into small, focused packages rather than bloating `main.go`.

- **Separation of concerns**:
  - `main`:
    - Defines CLI flags and interactive prompts.
    - Orchestrates the high-level “steps” (detect → crawl → format → write).
  - `feed`:
    - Deals only with feed-type detection and feed-related logic.
  - `crawler`:
    - Handles network IO, parsing documents, and constructing `Page` structs.
  - `formatter`:
    - Converts `[]Page` into a string representation based on the output type.
  - `writer`:
    - Responsible only for writing content to disk (including PDF generation).
  - `html2text`:
    - Converts sanitized HTML into a text representation; can be extended for new HTML → text behaviors.

- **When adding features**:
  - **Prefer** extending existing packages (e.g., new output format → `formatter` + `writer`), rather than mixing logic into `main`.
  - **Keep data structures aligned**: e.g., reuse or extend `crawler.Page` rather than inventing parallel structs unless necessary.

## CLI design and flag handling

- **Use Cobra for commands and flags**:
  - Define a `rootCmd` (and additional subcommands if needed) using `github.com/spf13/cobra`.
  - Attach a single `Run` function that calls a clearly named orchestrator function (e.g., `executeCrawlAndExport`).

- **Flag conventions**:
  - **StringVarP with short and long names** where it improves ergonomics:
    - Example: `--input, -i`, `--css, -c`, `--filename, -n`, `--type, -t`, `--format, -f`.
  - **Defaults**:
    - Provide sensible, documented defaults (e.g., `body` for CSS selector, `output` for filename, `txt` for type).
  - **Required vs optional**:
    - Treat truly required values like `input` as *logically required*:
      - It may have a flag, but the tool should still validate and fail clearly if missing.
    - Optional values must have meaningful defaults and clear prompts.

- **Adding or changing flags**:
  - Update:
    - The Cobra flag definitions in `init()`.
    - The interactive prompts in the main execution function so both paths stay in sync.
    - Any relevant validation helpers (e.g., `isValidOutputType`, `isValidFormat`).
  - Maintain **backwards compatibility**:
    - Avoid breaking existing flag names; add new flags instead of renaming, unless necessary.
    - If you must change behavior, clearly document it in `README.md`.

## Interactive, stepped input flow

- **Pattern**:
  - The CLI supports both:
    - Non-interactive usage via flags only.
    - Interactive usage where missing flags are requested step-by-step.

- **Prompting behavior**:
  - Use a helper like `promptUser(message, defaultValue string) string` to:
    - Display the current value or default in the prompt message.
    - Return the default when the user presses Enter.
  - **Order the prompts logically**:
    1. Core source (e.g., sitemap/RSS URL or file path) — **required**.
    2. Extraction-specific settings (CSS selector, filters).
    3. Output settings (filename, file type, format).

- **Validation and confirmation**:
  - **Validate** user inputs immediately (e.g., supported file types, formats).
  - After collecting inputs:
    - Print a clear summary of all settings.
    - Ask for an explicit `y/n` confirmation before doing any heavy work.
  - On `n` or any non-`y` answer:
    - Cancel the operation gracefully with a short, clear message.

- **Extending interactive flow**:
  - When introducing a new parameter:
    - Add a matching flag.
    - Add a prompt with a thoughtful default.
    - Include it in the pre-run summary so users see what will happen.

## Error handling and user feedback

- **Centralized error handling**:
  - Use an error helper similar to `handleError(step string, err error)` to:
    - Immediately exit on non-recoverable errors.
    - Prefix messages with a clear “step” label (e.g., “detecting feed type”, “formatting pages”).

- **Recoverable vs non-recoverable errors**:
  - For per-item failures inside loops (e.g., a single page failing to crawl):
    - Log the issue with enough context (usually the URL).
    - Skip that item and continue processing others.
  - For configuration or validation errors:
    - Fail fast with a clear message and a non-zero exit code.

- **Progress and transparency**:
  - For operations that may take time (e.g., crawling many URLs):
    - Use a progress indicator (like `github.com/schollz/progressbar/v3`) that:
      - Shows high-level progress (e.g., “Fetching sitemap pages”).
      - Updates per item processed.

## Content extraction and transformation

- **Content model**:
  - Reuse a `Page`-like struct with fields similar to:
    - `Title`, `URL`, `Description`, `Tags`, `Content`.
  - **Do** keep fields JSON-friendly and documented via struct tags.

- **Sanitization & transformation**:
  - Sanitize HTML before converting:
    - Restrict tags and attributes to a safe, useful subset.
    - Normalize excess newlines and whitespace.
  - Support multiple output “content” formats:
    - `html` (sanitized HTML),
    - `md` (Markdown),
    - `txt` (plain text via a converter helper).

- **Adding new transformations or formats**:
  - Place new transformation logic in the appropriate package:
    - HTML/Markdown/Text: `crawler` / `html2text`.
    - File-level formats (json, jsonl, txt, md, pdf, etc.): `formatter` + `writer`.
  - Keep transformation functions pure where possible:
    - Accept inputs, return outputs and errors.
    - Avoid global state.

## Output formats and file writing

- **Formatting**:
  - Let `formatter.FormatPages` be the main entry point for turning `[]Page` into a serialized string.
  - Add cases to the `FormatPages` switch for new formats instead of scattering logic.

- **Writing**:
  - Let `writer.WriteToFile` handle mapping logical formats to on-disk representations:
    - Text formats (`txt`, `md`, `json`, `jsonl`) written directly.
    - PDF or other binary formats handled by a dedicated helper (`writePDF`-style).
  - When adding new binary formats:
    - Keep binary-specific logic isolated from the rest of the pipeline.

## Commenting and code style

- **Comment style**:
  - Use **function-level doc comments** that:
    - State what the function does, what it returns, and important side-effects.
    - Highlight non-obvious behavior (e.g., filtering rules, special cases).
  - Avoid redundant line-by-line narration of obvious code.

- **Naming and structure**:
  - Use clear, descriptive names (`CrawlSitemap`, `DetectFeedType`, `FormatPages`, `WriteToFile`).
  - Keep functions small and single-purpose; factor out helpers instead of growing large monoliths.
  - Prefer explicit parameters over hidden globals (except for Cobra-bound flag variables where it improves ergonomics).

- **Consistency**:
  - Match existing patterns in:
    - Error messages (lowercase, step-prefixed).
    - Command descriptions and usage strings.
    - Progress bar messages.

## How Claude should work in this repo

- **When making changes**:
  - Start from the existing patterns in this codebase; **mirror them first**, then improve incrementally.
  - Ensure new flags, prompts, and behaviors are accurately documented in `README.md`.
  - Keep interactive and non-interactive usage paths equivalent in capability.

- **When adding a new feature** (e.g., new export format, new source type):
  - Identify the right package(s) to extend.
  - Add or extend data structures instead of introducing parallel, redundant ones.
  - Thread new options from flags → prompts → orchestrating function → specific package(s).

- **When unsure**:
  - Prefer conservative changes that maintain backward compatibility and match the existing UX:
    - Step-by-step prompts.
    - Clear defaults and summaries.
    - Helpful error messages and progress feedback.

