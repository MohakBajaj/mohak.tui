# CLAUDE.md

## Project Overview

**mohak.tui** - A cyberpunk-themed SSH-accessible terminal portfolio built with Go and TypeScript.

- **Runtime**: Bun 1.3.2 (Node.js compatible)
- **Languages**: Go 1.25.6, TypeScript 5.9
- **Monorepo**: Turborepo with Bun workspaces

## Architecture

```
mohak.tui/
├── apps/
│   ├── tui-server/      # Go: SSH server + Bubble Tea TUI
│   └── ai-gateway/      # TypeScript: Hono + AI SDK streaming
└── packages/
    └── shared-content/  # Shared resume, projects, bio, theme JSON
```

### apps/tui-server (Go)

SSH server using charmbracelet stack:

- **Wish** - SSH server framework
- **Bubble Tea** - TUI framework (tea.Model pattern)
- **Lip Gloss** - Styling
- **Bubbles** - Input, viewport components

Key patterns:

- Session-isolated TUI instances per SSH connection
- Rate limiting: max 5 sessions per IP
- Idle timeout: 10 minutes
- Content loaded from `shared-content` package at startup

### apps/ai-gateway (TypeScript)

Streaming AI proxy:

- **Hono** - HTTP server framework
- **Vercel AI SDK** - `streamText` with `@ai-sdk/gateway`
- In-memory rate limiting (10 req/min default)
- SSE streaming responses

### packages/shared-content

Static content consumed by both apps:

- `resume.json` - Structured resume data
- `projects.json` - Project portfolio
- `bio.md` - Bio markdown
- `theme.json` - Color themes
- `buildSystemPrompt()` - AI system prompt builder

## Development Commands

```bash
# Install all dependencies
bun install

# Run both services (Turbo parallel)
bun run dev

# Run services individually
bun run dev:ai     # AI Gateway on :3001
bun run dev:tui    # SSH server on :2222

# Format code
bun run format     # Prettier

# Build Go TUI server
cd apps/tui-server && go build -o bin/tui-server .

# Connect to local SSH
ssh -p 2222 localhost
```

## Environment Variables

Copy `.env.example` to `.env` and configure:

| Variable                | Required | Default                       | Description               |
| ----------------------- | -------- | ----------------------------- | ------------------------- |
| `AI_GATEWAY_API_KEY`    | Yes      | -                             | Vercel AI Gateway API key |
| `AI_GATEWAY_MODEL`      | No       | `anthropic/claude-sonnet-4.5` | AI model                  |
| `AI_GATEWAY_PORT`       | No       | `3001`                        | Gateway port              |
| `AI_GATEWAY_RATE_LIMIT` | No       | `10`                          | Requests per minute       |
| `SSH_HOST`              | No       | `0.0.0.0`                     | SSH bind host             |
| `SSH_PORT`              | No       | `2222`                        | SSH bind port             |
| `AI_GATEWAY_URL`        | No       | `http://localhost:3001`       | Gateway URL for TUI       |

## TUI Commands

Users interact via slash commands:

- `/help` - Show available commands
- `/about` - Bio view
- `/projects` - Project list
- `/open <id>` - Project detail
- `/resume` - Resume view
- `/clear` - Reset chat
- `/exit` - Disconnect

Direct text input sends messages to AI chat.

## Code Patterns

### Go TUI

- Follow Bubble Tea's `Init() → Update() → View()` pattern
- Messages are `tea.Msg` types; define custom `XxxMsg` structs
- Use `tea.Batch()` for multiple commands
- Styles via `theme.Manager.Styles()` - never create ad-hoc styles

### TypeScript

- Hono handlers return `c.json()` or streaming responses
- Use `Bun.env` for environment variables
- Type all request/response bodies explicitly

## Important Notes

- SSH server creates `.ssh/id_ed25519` host key on first run
- AI gateway health check runs async on TUI startup (non-blocking)
- Chat history maintained per session, lost on disconnect
- Markdown rendering in TUI uses custom renderer (not glamour)
- `ESC` key cancels streaming or returns to chat view
- Content path auto-detected relative to Go binary location
