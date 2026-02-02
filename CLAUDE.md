# CLAUDE.md

## Project Overview

**mohak.tui** - A cyberpunk-themed SSH-accessible terminal portfolio built with Go and TypeScript.

- **Runtime**: Bun 1.3.2 (Node.js compatible)
- **Languages**: Go 1.21+, TypeScript 5.x
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
- **Telemetry via `internal/telemetry/`** - PostHog analytics + structured logging
- **Dotenv support** - Loads `.env` file at startup

### apps/ai-gateway (TypeScript)

Streaming AI proxy:

- **Hono** - HTTP server framework
- **Vercel AI SDK** - `streamText` with `@ai-sdk/gateway`
- In-memory rate limiting (10 req/min default)
- SSE streaming responses
- **Intent detection** - Classifies queries (greeting, about, experience, skills, projects, etc.)
- **Structured logging** - via `lib/logger.ts`
- **PostHog analytics** - via `lib/analytics.ts`

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

| Variable                | Required | Default                    | Description               |
| ----------------------- | -------- | -------------------------- | ------------------------- |
| `AI_GATEWAY_API_KEY`    | Yes      | -                          | Vercel AI Gateway API key |
| `AI_GATEWAY_MODEL`      | No       | `openai/gpt-oss-20b`       | AI model                  |
| `AI_GATEWAY_PORT`       | No       | `3001`                     | Gateway port              |
| `AI_GATEWAY_RATE_LIMIT` | No       | `10`                       | Requests per minute       |
| `AI_GATEWAY_MAX_TOKENS` | No       | `1024`                     | Max tokens in AI response |
| `AI_TEMPERATURE`        | No       | `0.7`                      | Response creativity (0-1) |
| `SSH_HOST`              | No       | `0.0.0.0`                  | SSH bind host             |
| `SSH_PORT`              | No       | `2222`                     | SSH bind port             |
| `AI_GATEWAY_URL`        | No       | `http://localhost:3001`    | Gateway URL for TUI       |
| `POSTHOG_API_KEY`       | No       | -                          | PostHog analytics key     |
| `POSTHOG_HOST`          | No       | `https://us.i.posthog.com` | PostHog instance URL      |
| `LOG_LEVEL`             | No       | `info`                     | debug, info, warn, error  |
| `LOG_FORMAT`            | No       | `pretty`                   | pretty (colored) or json  |

## TUI Commands

Users interact via slash commands:

- `/help` - Show available commands
- `/about` - Bio view
- `/projects` - Project list
- `/open <id>` - Project detail
- `/resume` - Resume view
- `/exp` - Experience view
- `/clear` - Reset chat
- `/exit` - Disconnect

Direct text input sends messages to AI chat.

## Keyboard Shortcuts

| Shortcut | Action            |
| -------- | ----------------- |
| `Alt+H`  | Help              |
| `Alt+A`  | About / Profile   |
| `Alt+P`  | Projects list     |
| `Alt+R`  | Resume            |
| `Alt+E`  | Experience        |
| `Alt+W`  | Home / Welcome    |
| `Alt+C`  | Clear chat        |
| `Alt+Q`  | Quit              |
| `Alt+M`  | Toggle mouse mode |
| `Ctrl+U` | Clear input line  |
| `ESC`    | Back / Cancel     |

## Code Patterns

### Go TUI

- Follow Bubble Tea's `Init() → Update() → View()` pattern
- Messages are `tea.Msg` types; define custom `XxxMsg` structs
- Use `tea.Batch()` for multiple commands
- **IMPORTANT**: Styles via `theme.Manager.Styles()` - NEVER create ad-hoc styles
- **IMPORTANT**: All identifiers in telemetry must be SHA256 hashed for PII safety

### TypeScript

- Hono handlers return `c.json()` or streaming responses
- Use `Bun.env` for environment variables
- Type all request/response bodies explicitly

## Telemetry

### Logging

Both services use structured logging:

```bash
# Development
LOG_LEVEL=debug LOG_FORMAT=pretty

# Production
LOG_LEVEL=info LOG_FORMAT=json
```

### PostHog Analytics

Events tracked (all PII-safe with hashed identifiers):

**TUI Server:**

- `tui_session_connected` / `tui_session_disconnected`
- `tui_view_changed`, `tui_command_executed`
- `tui_chat_sent` / `tui_chat_received`

**AI Gateway:**

- `ai_gateway_chat_request` / `ai_gateway_chat_response`
- `ai_gateway_chat_error`, `ai_gateway_rate_limit_hit`

## Important Notes

- SSH server creates `.ssh/id_ed25519` host key on first run
- Go server loads `.env` file at startup via godotenv
- AI gateway health check runs async on TUI startup (non-blocking)
- Chat history maintained per session, lost on disconnect
- Markdown rendering in TUI uses custom renderer (not glamour)
- `ESC` key cancels streaming or returns to chat view
- Content path auto-detected relative to Go binary location
- All analytics identifiers are SHA256 hashed for privacy
