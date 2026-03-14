# marginalia

A CLI tool for managing plain-text Markdown notes organized into collections.

## Features

- Organize notes into named collections
- Interactive TUI for browsing, creating, editing, and deleting notes
- Per-collection templates using Go's `text/template` syntax
- Automatic git backup and sync
- Respects `$VISUAL` / `$EDITOR` environment variables

## Requirements

- Go 1.21 or later
- A terminal editor (e.g. `nvim`, `vim`, `nano`, or any editor in your `$PATH`)
- Git (optional, for backup sync)

## Installation

```bash
# Build binary to ./dist/margi
make build

# Install binary to ~/.local/bin/margi
make install
```

## Usage

### Browse notes

Running `margi` without arguments opens an interactive picker across all collections.

```bash
margi
```

### Create a new note

```bash
# Choose a collection interactively, then open the new note in your editor
margi new "my note title"

# Skip the picker by specifying the collection directly
margi new journal "my note title"
```

### Edit a note

Fuzzy-search by filename and open the matching note. If multiple files match, a numbered list is shown.

```bash
margi edit "search term"
```

### Delete a note

Open a TUI delete picker filtered by the search term, with a confirmation dialog before deletion.

```bash
margi rm "search term"

# Open the delete picker with no filter
margi rm
```

### List collections

```bash
margi collections
```

### Sync

Pull remote changes and push any local uncommitted notes on demand:

```bash
margi sync
```

### Git sync

Sync is automatic when git backup is configured. On every startup `margi` pulls from the configured remote. After every create, edit, or delete operation it commits and pushes the changes.

## Configuration

The config file is loaded from `~/.config/marginalia/config.toml`. It is created automatically with defaults on first run.

```toml
editor = "nvim"

[backup]
provider = "git"

[backup.git]
repo   = "/path/to/local/repo"
remote = "origin"
branch = "main"
```

**Editor resolution order:** `config.toml` value → `$VISUAL` → `$EDITOR` → `vi`

**Git backup:** When `backup.provider = "git"` and `backup.git.repo` is set, `margi` initializes a git repository in the data directory (if one does not already exist), pulls on startup, and commits + pushes after every write operation.

## Note Templates

Per-collection templates are stored at `~/.config/marginalia/collections/<collection>.md`. They use Go's `text/template` syntax.

Available variables:

| Variable | Description |
|---|---|
| `{{.Title}}` | The note title passed on the command line |
| `{{.Collection}}` | The collection name |
| `{{.Date}}` | Current date and time in RFC3339 format |

Example template (`~/.config/marginalia/collections/journal.md`):

```markdown
# {{.Title}}

Date: {{.Date}}
Collection: {{.Collection}}

---

```

If no template is found for a collection, a minimal default header is used:

```markdown
# <title>

## Created: <timestamp>
## Collection: <collection>
```

## Data Layout

```
~/.local/share/marginalia/
└── collections/
    ├── journal/
    │   ├── 20260101-120000-my-first-entry.md
    │   └── 20260314-093000-another-entry.md
    └── work/
        └── 20260310-150000-meeting-notes.md

~/.config/marginalia/
├── config.toml
└── collections/
    ├── journal.md   # template for the journal collection
    └── work.md      # template for the work collection
```

Note filenames follow the pattern `YYYYMMDD-HHMMSS-slug.md`.
