# shiplog

Export Claude Code sessions to beautiful, self-contained HTML chat pages.

Turn your AI-assisted development sessions into shareable documents — perfect for team reviews, portfolio pieces, or showing your CEO what you built with Claude.

## Features

- Exports Claude Code JSONL transcripts to self-contained HTML
- Beautiful chat interface with user/assistant bubbles
- Embedded screenshots and images
- Tool usage indicators (grouped for readability)
- Client-side markdown rendering (code blocks, tables, lists)
- Dark sidebar with session metadata
- Project-scoped session discovery (auto-detects current project)
- Fuzzy session search by name or UUID

## Install

### Homebrew (macOS/Linux)

```bash
brew install HabibPro1999/tap/shiplog
```

### Go

```bash
go install github.com/HabibPro1999/shiplog@latest
```

### Direct download

```bash
curl -fsSL https://raw.githubusercontent.com/HabibPro1999/shiplog/main/install.sh | sh
```

Or download a binary from [GitHub Releases](https://github.com/HabibPro1999/shiplog/releases).

## Usage

```bash
# List sessions (scoped to current project)
shiplog

# List all sessions across all projects
shiplog -a

# Export a session by name
shiplog "resume builder"

# Export with custom output path
shiplog -o ~/Desktop/session.html "resume builder"

# Export by session UUID
shiplog --session-id 42b2caf1
```

### Session Discovery

When run from a project directory, only sessions for that project are shown. Use `-a` to see all sessions.

Sessions are read from `~/.claude/projects/*/*.jsonl`.

## CLI Reference

| Flag           | Short | Description                              |
| -------------- | ----- | ---------------------------------------- |
| `--list`       | `-l`  | List sessions                            |
| `--all`        | `-a`  | Show all sessions (ignore project scope) |
| `--output`     | `-o`  | Output HTML file path                    |
| `--session-id` |       | Export by UUID prefix                    |
| `--version`    | `-v`  | Show version                             |

## How It Works

1. Scans `~/.claude/projects/` for JSONL session files
2. Parses entries, filtering out system messages and tool internals
3. Groups consecutive tool calls into compact indicators
4. Renders a self-contained HTML page with inline CSS/JS
5. Embeds base64 images directly in the HTML

## License

MIT
