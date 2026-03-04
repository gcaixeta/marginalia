# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build      # Build binary to ./dist/margi
make install    # Install binary to ~/.local/bin/margi
make run        # Run without building (go run ./cmd/margi)
make test       # Run all tests
go test ./internal/collection/...  # Run a single package's tests
```

## Architecture

`marginalia` (`margi`) is a CLI tool for managing plain-text Markdown notes organized into collections.

**Entry point:** `cmd/margi/main.go` — parses CLI args and dispatches to internal packages.

**Data layout on disk:**
- Notes: `~/.local/share/marginalia/collections/<collection>/<timestamp-slug>.md`
- Config: `~/.config/marginalia/config.toml`
- Collection snippets/templates: `~/.config/marginalia/collections/<collection>.md`

**Internal packages:**

| Package | Role |
|---|---|
| `internal/storage` | File I/O: `DataDir()`, `ListAllFiles()`, `FindFiles()`, `FindFilePath()`. All filesystem access goes through here. |
| `internal/collection` | Collection management: list, create, check existence. Delegates FS ops to `storage`. |
| `internal/config` | TOML config load. `Config` struct holds `BackupConfig` (provider, git repo/remote/branch). `Default()` and `Save()` are not yet implemented — this is the current compile error. |
| `internal/editor` | Opens a file in `nvim` (hardcoded). |
| `internal/slug` | Converts titles to filesystem-safe slugs (Unicode normalization + kebab-case). Filenames are `YYYYMMDD-HHMMSS-slug.md`. |
| `internal/snippet` | Reads Go `text/template` snippet files for new notes. Falls back to a minimal default header if no template is found. |
| `internal/ui` | Bubbletea TUI components: `picker.go` (collection picker), `delete_picker.go` (file deletion picker), `confirm.go` (yes/no dialog). |

**CLI commands:**
- `margi new [title]` — interactive collection picker, then opens note in nvim
- `margi new [collection] [title]` — skip picker, create directly
- `margi edit [search_term]` — fuzzy-search files by name, open in nvim
- `margi rm [search_term]` — TUI delete picker + confirmation dialog
- `margi collections` — list all collections with file counts

**Current known issue:** `config.Default()` and `config.Save()` are called in `main.go` but not yet defined in `internal/config/loader.go`. The package only defines `Load()`.

**UI pattern:** All TUI components follow the Bubbletea Model-View-Update pattern. Each component has a `RunXxx()` function that creates a `tea.Program`, runs it, and returns the result.
