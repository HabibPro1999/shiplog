<p align="center">
  <h1 align="center">shiplog</h1>
  <p align="center">Export Claude Code sessions to beautiful, self-contained HTML pages.</p>
</p>

<p align="center">
  <a href="https://github.com/HabibPro1999/shiplog/releases"><img src="https://img.shields.io/github/v/release/HabibPro1999/shiplog" alt="GitHub Release"></a>
  <a href="https://github.com/HabibPro1999/shiplog/blob/main/LICENSE"><img src="https://img.shields.io/github/license/HabibPro1999/shiplog" alt="License"></a>
  <img src="https://img.shields.io/github/go-mod/go-version/HabibPro1999/shiplog" alt="Go Version">
  <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux-blue" alt="Platform">
</p>

---

Turn your AI-assisted development sessions into shareable documents -- perfect for team reviews, portfolio pieces, or showing stakeholders what you built with Claude.

<!-- Add screenshot: shiplog_demo.png -->

## Quick Start

**Install** (pick one):

```bash
# Homebrew (macOS/Linux)
brew install HabibPro1999/tap/shiplog

# Go
go install github.com/HabibPro1999/shiplog@latest

# Shell script
curl -fsSL https://raw.githubusercontent.com/HabibPro1999/shiplog/main/install.sh | sh
```

Or grab a binary from [GitHub Releases](https://github.com/HabibPro1999/shiplog/releases).

**Export a session:**

```bash
shiplog "resume builder"
#  Found: "resume builder" (my-project)
#  Parsing transcript...
#  142 entries
#  38 user messages, 41 assistant messages
#  Generating HTML...
#  Written to: resume_builder.html (2.3 MB)
```

Open the HTML file in any browser. No server, no dependencies -- just one file.

## Features

- **Self-contained HTML** -- single file with inline CSS, JS, and base64-encoded images
- **Chat interface** -- clean user/assistant message bubbles with proper styling
- **Markdown rendering** -- code blocks with syntax highlighting, tables, lists
- **Tool call grouping** -- consecutive tool uses collapsed into compact indicators
- **Project-scoped discovery** -- auto-detects your current project's sessions
- **Fuzzy search** -- find sessions by name or UUID prefix
- **Dark sidebar** -- session metadata displayed in a navigable side panel

## Usage

### List sessions

```bash
# List sessions for the current project
shiplog

# List all sessions across all projects
shiplog -a
```

Example output:

```
  #    Title                     ID         Project                                  Date
  ---- ------------------------- ---------- ---------------------------------------- ------------
  1    resume builder            42b2caf1   my-project                               Jan 15, 2026
  2    auth refactor             8f3e21a0   my-project                               Jan 14, 2026
  3    fix dark mode             c91d44b7   my-project                               Jan 12, 2026

  Total: 3 sessions
```

### Export a session

```bash
# By name (fuzzy match)
shiplog "auth refactor"

# By UUID prefix
shiplog --session-id 42b2caf1

# Custom output path
shiplog -o ~/Desktop/session.html "resume builder"
```

### CLI Reference

| Flag           | Short | Description                              |
| -------------- | ----- | ---------------------------------------- |
| `--list`       | `-l`  | List sessions                            |
| `--all`        | `-a`  | Show all sessions (ignore project scope) |
| `--output`     | `-o`  | Output HTML file path                    |
| `--session-id` |       | Export by session UUID prefix            |
| `--version`    | `-v`  | Show version                             |

## How It Works

1. Scans `~/.claude/projects/` for JSONL session files
2. Parses transcript entries, filtering out system messages and tool internals
3. Groups consecutive tool calls into compact indicators
4. Renders a self-contained HTML page with all assets inlined
5. Embeds screenshots and images as base64 directly in the output

## Contributing

Contributions are welcome. Open an issue or submit a pull request at [github.com/HabibPro1999/shiplog](https://github.com/HabibPro1999/shiplog).

## License

[MIT](LICENSE)
