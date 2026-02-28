# csvmap

A TUI tool for interactive CSV column mapping — date reformatting, string mapping with glob patterns, column reordering, and row filtering.

Built for cleaning up CSV exports (like credit card statements) with persistent mapping files that grow over time.

## Install

```bash
go install github.com/erikburt/csvmap@latest
```

Or build from source:

```bash
git clone https://github.com/erikburt/csvmap
cd csvmap
go build -o csvmap .
```

## Usage

```bash
csvmap <input.csv> [flags]
```

Flags:
- `--has-header` / `--no-header` — specify if CSV has a header row
- `--config-dir <path>` — override config directory (default: `~/.config/csvmap/`)

Output is written to `<input>_mapped.csv`.

## Features

- **Date reformatting** — auto-detects input format, converts to any Go time layout
- **String mapping** — glob patterns (`*AMAZON*` → `Amazon`), case-insensitive, persistent JSON mapping files
- **Drop columns** — remove unwanted columns (press `x` on a column)
- **Filter rows** — interactively review and drop rows
- **Reorder columns** — press `space` to pick up, move with arrows, `space` to drop

## Disclaimer

This project was written entirely by Claude Code (claude-code).
