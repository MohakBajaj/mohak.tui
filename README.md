# mohak.sh - SSH TUI Portfolio

A cyberpunk-themed SSH-accessible terminal portfolio built with Go, Bubble Tea, and Wish.

```bash
ssh mohak.sh
```

## Tech Stack

- **TUI Server**: Go + Bubble Tea + Lip Gloss + Wish
- **AI Gateway**: Bun + Hono + Vercel AI SDK
- **Monorepo**: Turborepo + Bun

## Project Structure

```
mohak.tui/
├─ apps/
│  ├─ tui-server/        # Go SSH + Bubble Tea TUI
│  └─ ai-gateway/        # Bun AI streaming service
├─ packages/
│  └─ shared-content/    # Resume, projects, bio, theme
├─ turbo.json
└─ package.json
```

## Local Development

### Prerequisites

- [Go](https://go.dev/dl/) 1.21+
- [Bun](https://bun.sh/) 1.0+

### Setup

1. Clone and install dependencies:

```bash
git clone https://github.com/mohakbajaj/mohak-tui.git
cd mohak-tui
bun install
```

2. Configure environment variables:

```bash
cp .env.example .env
# Edit .env with your AI_GATEWAY_API_KEY
```

3. Build the Go TUI server:

```bash
cd apps/tui-server
go build -o bin/tui-server .
```

### Running Locally

**Option 1: Run both services**

Terminal 1 - AI Gateway:

```bash
bun run dev:ai
```

Terminal 2 - TUI Server:

```bash
bun run dev:tui
```

**Option 2: Run with turbo**

```bash
bun run dev
```

### Connect via SSH

```bash
ssh -p 2222 localhost
```

## Available Commands

| Command               | Description             |
| --------------------- | ----------------------- |
| `help`                | Show available commands |
| `about`               | Learn about me          |
| `projects`            | View my projects        |
| `open <id>`           | View project details    |
| `resume`              | View my resume          |
| `chat`                | Chat with AI assistant  |
| `theme <dark\|light>` | Switch color theme      |
| `clear`               | Clear the screen        |
| `exit`                | Exit the session        |

## Environment Variables

### AI Gateway (`apps/ai-gateway`)

| Variable                | Description               | Default                       |
| ----------------------- | ------------------------- | ----------------------------- |
| `AI_GATEWAY_API_KEY`    | Vercel AI Gateway API key | Required                      |
| `AI_GATEWAY_MODEL`      | Model to use              | `anthropic/claude-sonnet-4.5` |
| `AI_GATEWAY_PORT`       | Server port               | `3001`                        |
| `AI_GATEWAY_RATE_LIMIT` | Requests per minute       | `10`                          |
| `AI_GATEWAY_MAX_TOKENS` | Max response tokens       | `1024`                        |

### TUI Server (`apps/tui-server`)

| Variable         | Description            | Default                 |
| ---------------- | ---------------------- | ----------------------- |
| `SSH_HOST`       | SSH server host        | `0.0.0.0`               |
| `SSH_PORT`       | SSH server port        | `2222`                  |
| `AI_GATEWAY_URL` | AI Gateway URL         | `http://localhost:3001` |
| `CONTENT_PATH`   | Path to shared content | Auto-detected           |

## Production Deployment

### Docker (Coming Soon)

```bash
docker compose up -d
```

### Manual Deployment

1. Build the TUI server:

```bash
cd apps/tui-server
go build -o bin/tui-server .
```

2. Run with systemd or supervisor

3. Configure firewall to allow port 22 (or custom SSH port)

## Security Notes

- SSH sessions are isolated per connection
- Rate limiting on AI chat (configurable)
- Max sessions per IP (default: 5)
- Idle timeout (default: 10 minutes)
- No shell access - TUI only

## License

MIT
