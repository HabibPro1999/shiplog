<p align="center">
  <h1 align="center">shiplog</h1>
  <p align="center">Export AI coding sessions to beautiful, self-contained HTML pages.</p>
</p>

<p align="center">
  <a href="https://github.com/HabibPro1999/shiplog/releases"><img src="https://img.shields.io/github/v/release/HabibPro1999/shiplog" alt="GitHub Release"></a>
  <a href="https://github.com/HabibPro1999/shiplog/blob/main/LICENSE"><img src="https://img.shields.io/github/license/HabibPro1999/shiplog" alt="License"></a>
  <img src="https://img.shields.io/github/go-mod/go-version/HabibPro1999/shiplog" alt="Go Version">
  <img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux-blue" alt="Platform">
</p>

<p align="center">
  <a href="https://claude.ai"><img src="https://img.shields.io/badge/Claude_Code-D97757?style=flat-square&logo=claude&logoColor=white" alt="Claude Code"></a>
  <a href="https://github.com/openai/codex"><img src="https://img.shields.io/badge/Codex_CLI-74aa9c?style=flat-square&logo=openai&logoColor=white" alt="Codex CLI"></a>
  <a href="https://aider.chat"><img src="https://img.shields.io/badge/Aider-4A90E2?style=flat-square&logoColor=white" alt="Aider"></a>
  <a href="https://github.com/opencode-ai/opencode"><img src="https://img.shields.io/badge/OpenCode-FF6B35?style=flat-square&logoColor=white" alt="OpenCode"></a>
  <a href="https://aws.amazon.com/q/developer/"><img src="https://img.shields.io/badge/Amazon_Q-FF9900?style=flat-square&logo=amazon-aws&logoColor=white" alt="Amazon Q"></a>
</p>

---

Turn your AI-assisted development sessions into shareable documents -- perfect for team reviews, portfolio pieces, or showing stakeholders what you built.

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
shiplog "auth refactor"
#  Found: "auth refactor" (my-project) [Claude Code]
#  Parsing transcript...
#  142 entries
#  38 user messages, 41 assistant messages
#  Generating HTML...
#  Written to: auth_refactor.html (2.3 MB)
```

Open the HTML file in any browser. No server, no dependencies -- just one file.

## Supported Tools

| Tool            | Sessions Location          | Format   |
| --------------- | -------------------------- | -------- |
| **Claude Code** | `~/.claude/projects/`      | JSONL    |
| **Codex CLI**   | `~/.codex/sessions/`       | JSONL    |
| **Aider**       | `.aider.chat.history.md`   | Markdown |
| **OpenCode**    | `~/.local/share/opencode/` | JSON     |
| **Amazon Q**    | `~/.aws/amazonq/history/`  | JSON     |

Shiplog auto-discovers sessions from all installed tools. Use `--tool` to filter by a specific tool.

## Features

- **Multi-tool support** -- works with Claude Code, Codex CLI, Aider, OpenCode, and Amazon Q
- **Self-contained HTML** -- single file with inline CSS, JS, and base64-encoded images
- **Interactive TUI** -- fuzzy-search session picker when multiple matches are found
- **Chat interface** -- clean user/assistant message bubbles with proper styling
- **Markdown rendering** -- code blocks with syntax highlighting, tables, lists
- **Tool call grouping** -- consecutive tool uses collapsed into compact indicators
- **Project-scoped discovery** -- auto-detects your current project's sessions
- **Fuzzy search** -- find sessions by name, project, or UUID prefix

## Usage

### List sessions

```bash
# Interactive session picker for the current project
shiplog

# List all sessions across all projects
shiplog -a

# Filter by tool
shiplog --tool claude
shiplog --tool codex

# Show only recent sessions
shiplog --recent 7
```

### Export a session

```bash
# By name (fuzzy match)
shiplog "auth refactor"

# By UUID prefix
shiplog --session-id 42b2caf1

# Custom output path
shiplog -o ~/Desktop/session.html "resume builder"

# Export from a specific tool
shiplog --tool aider "refactor models"
```

### CLI Reference

| Flag           | Short | Description                                              |
| -------------- | ----- | -------------------------------------------------------- |
| `--list`       | `-l`  | List sessions                                            |
| `--all`        | `-a`  | Show all sessions (ignore project scope)                 |
| `--tool`       | `-t`  | Filter by tool (claude, codex, aider, opencode, amazonq) |
| `--recent`     | `-r`  | Show sessions from last N days                           |
| `--output`     | `-o`  | Output HTML file path                                    |
| `--session-id` |       | Export by session UUID prefix                            |
| `--version`    | `-v`  | Show version                                             |

## How It Works

1. Discovers sessions from all supported tools under `$HOME`
2. Filters by current project directory (or `--all` for global)
3. Parses transcripts, filtering out system messages and tool internals
4. Groups consecutive tool calls into compact indicators
5. Renders a self-contained HTML page with all assets inlined
6. Embeds screenshots and images as base64 directly in the output

## Contributing

Contributions are welcome. Open an issue or submit a pull request at [github.com/HabibPro1999/shiplog](https://github.com/HabibPro1999/shiplog).

## License

[MIT](LICENSE)
