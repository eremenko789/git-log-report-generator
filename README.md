# git-html-report

CLI utility that generates a single static HTML report from `git log` for any revision range.

## Installation

Install latest release:

```bash
curl -fsSL https://raw.githubusercontent.com/eremenko789/git-log-report-generator/main/install.sh | bash
```

Install specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/eremenko789/git-log-report-generator/main/install.sh | VERSION=v1.0.0 bash
```

## Usage

```bash
git-html-report <from-ref> <to-ref> [flags]
```

Examples:

```bash
git-html-report v1.2.0 v1.3.0 --output release-notes.html
git-html-report HEAD~10 HEAD --repo /path/to/repo
git-html-report abc1234 def5678 --output report.html --title "Release 2.0"
```

## Flags

- `--repo`, `-r` path to git repository (default `.`)
- `--output`, `-o` output HTML file path (default `git-report.html`)
- `--title`, `-t` report title (default `Git Commits Report`)
- `--author` filter commits by author name/email
- `--no-files` do not include changed files list
- `--version`, `-v` print version information and exit

## Development

```bash
go test ./...
go build ./cmd/git-html-report
```

## Features

- Parses commits with robust delimiters (`\x1e`, `\x1f`) from `git log`
- Optional per-commit changed files from `git show --name-status`
- Self-contained HTML output with inline CSS/JS
- Light/dark themes with manual toggle and `prefers-color-scheme`
- Commit subject filter in browser without reload
- CI workflow (lint, test, build) and release workflow (GoReleaser)
