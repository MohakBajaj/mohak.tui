# bmohak.xyz - SSH TUI Portfolio

A cyberpunk-themed SSH-accessible terminal portfolio built with Go, Bubble Tea, and Wish. Features in-process AI chat via Vercel AI Gateway, embedded content, full observability with PostHog, and production-grade logging.

```bash
ssh bmohak.xyz
```

## Features

- **Cyberpunk UI** - Neon colors, box-drawing characters, and terminal aesthetics
- **AI Chat** - Go-native streaming chat with intent-aware responses
- **Full Observability** - PostHog analytics + structured logging
- **Responsive** - Adapts to terminal size with proper text wrapping
- **Keyboard Navigation** - Alt+key shortcuts for quick access
- **Session Management** - Rate limiting, idle timeout, PII-safe logging

## Tech Stack

| Component  | Technology                                   |
| ---------- | -------------------------------------------- |
| TUI Server | Go + Bubble Tea + Lip Gloss + Wish           |
| AI Runtime | Go + Vercel AI Gateway OpenAI-compatible API |
| Analytics  | PostHog Go SDK                               |
| Monorepo   | Turborepo + Bun                              |

## Project Structure

```
mohak.tui/
├── apps/
│   ├── tui-server/           # Go SSH + Bubble Tea TUI + integrated AI
│   │   ├── internal/
│   │   │   ├── app/          # Main Bubble Tea model
│   │   │   ├── ai/           # Prompting + provider abstraction
│   │   │   ├── content/      # Content loaders
│   │   │   ├── telemetry/    # Logging + PostHog analytics
│   │   │   ├── theme/        # Cyberpunk color scheme
│   │   │   └── ui/           # Views + markdown renderer
│   │   └── main.go
├── packages/
│   └── shared-content/       # Resume, projects, bio data
├── turbo.json
└── package.json
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
# Edit .env with your API keys
```

3. Build the Go TUI server:

```bash
cd apps/tui-server
go build -o bin/tui-server .
```

### Running Locally

**Option 1: Run the TUI server from the repo root**

```bash
bun start
```

**Option 2: Run the Go server directly**

```bash
bun run dev:tui
```

### Connect via SSH

```bash
ssh -p 2222 localhost
```

### Termux Build

Build Linux ARM64 artifacts for Termux with:

```bash
bun run build:termux
```

This writes a deployable bundle to `dist/termux` with:

- `tui-server-linux-arm64` - self-contained Go SSH server binary
- `run-termux.sh` - launcher for Termux
- `README.termux.md` - deployment notes

On Termux, run:

```bash
cd dist/termux
chmod +x ./run-termux.sh ./tui-server-linux-arm64
./run-termux.sh
```

The shared content is embedded into the Go binary, so no Bun runtime or extra content files are required on Termux.

## Keyboard Shortcuts

| Shortcut | Action                            |
| -------- | --------------------------------- |
| `Alt+H`  | Help                              |
| `Alt+A`  | About / Profile                   |
| `Alt+P`  | Projects list                     |
| `Alt+R`  | Resume                            |
| `Alt+E`  | Experience                        |
| `Alt+W`  | Home / Welcome                    |
| `Alt+C`  | Clear chat                        |
| `Alt+Q`  | Quit                              |
| `Alt+M`  | Toggle mouse mode                 |
| `Ctrl+U` | Clear input line                  |
| `ESC`    | Back / Cancel                     |
| `1-9`    | Select project (in projects view) |

## Slash Commands

| Command      | Description          |
| ------------ | -------------------- |
| `/help`      | Show help            |
| `/about`     | View profile         |
| `/projects`  | Browse projects      |
| `/open <id>` | View project details |
| `/resume`    | View credentials     |
| `/exp`       | View experience      |
| `/clear`     | Reset chat           |
| `/exit`      | Disconnect           |

## Environment Variables

### Integrated AI + TUI (`.env`)

| Variable                | Description                     | Default                    |
| ----------------------- | ------------------------------- | -------------------------- |
| `AI_GATEWAY_API_KEY`    | Vercel AI Gateway API key       | Required                   |
| `AI_GATEWAY_MODEL`      | Model identifier                | `openai/gpt-oss-20b`       |
| `AI_GATEWAY_RATE_LIMIT` | Requests per minute             | `10`                       |
| `AI_GATEWAY_MAX_TOKENS` | Max response tokens             | `1024`                     |
| `AI_TEMPERATURE`        | Response creativity (0-1)       | `0.7`                      |
| `SSH_HOST`              | SSH server bind address         | `0.0.0.0`                  |
| `SSH_PORT`              | SSH server port                 | `2222`                     |
| `CONTENT_PATH`          | Optional content override path  | Embedded content           |
| `POSTHOG_API_KEY`       | PostHog project API key         | Optional                   |
| `POSTHOG_HOST`          | PostHog instance URL            | `https://us.i.posthog.com` |
| `LOG_LEVEL`             | Logging level                   | `info`                     |
| `LOG_FORMAT`            | Output format (`pretty`/`json`) | `pretty`                   |

## Observability

### Logging

The Go server uses structured logging with configurable levels and formats:

```bash
# Development (pretty output)
LOG_LEVEL=debug LOG_FORMAT=pretty

# Production (JSON for log aggregators)
LOG_LEVEL=info LOG_FORMAT=json
```

**Log Levels:** `debug`, `info`, `warn`, `error`

### PostHog Analytics

Events tracked (all PII-safe with hashed identifiers):

**TUI Server:**

- `tui_session_connected` - User connects via SSH
- `tui_session_disconnected` - User disconnects
- `tui_view_changed` - Navigation between views
- `tui_command_executed` - Slash commands
- `tui_chat_sent` / `tui_chat_received` - Chat interactions

**Integrated AI layer:**

- `ai_gateway_chat_request` - Incoming chat requests
- `ai_gateway_chat_response` - Successful responses
- `ai_gateway_chat_error` - Errors
- `ai_gateway_rate_limit_hit` - Rate limiting events

### Session Data Captured

All identifiers are SHA256 hashed for privacy:

```json
{
  "session_hash": "a1b2c3d4e5f6",
  "user_hash": "f6e5d4c3b2a1",
  "ip_hash": "1a2b3c4d5e6f",
  "terminal": "xterm-256color",
  "width": 120,
  "height": 40,
  "key_type": "ssh-ed25519",
  "term_program": "iTerm.app",
  "shell": "zsh"
}
```

## AI System

The AI assistant (NEURAL) runs inside the Go TUI server and uses intent-aware prompting:

**Query Intents:**

- `greeting` - Hello, hi, hey
- `about` - Who is Mohak, tell me about
- `experience` - Work, jobs, roles
- `skills` - Technologies, stack
- `projects` - What has he built
- `contact` - Email, social links
- `education` - Degree, college
- `achievements` - Awards, competitions
- `meta` - Questions about this portfolio

**Features:**

- Dynamic context injection based on intent
- Message preprocessing (normalizes slang)
- Token-efficient (only loads relevant sections)
- Stop sequences to prevent runaway generation
- Frequency/presence penalties for natural responses

## Security

- **Isolated sessions** - Each SSH connection is sandboxed
- **Rate limiting** - Configurable per-session limits
- **IP throttling** - Max 5 sessions per IP
- **Idle timeout** - 10 minute default
- **No shell access** - TUI only, no command execution
- **PII-safe logging** - All identifiers hashed

## Production Deployment

### Quick Start with Docker

```bash
# Clone repository
git clone https://github.com/mohakbajaj/mohak-tui.git
cd mohak-tui

# Configure environment
cp .env.example .env
# Edit .env with your AI_GATEWAY_API_KEY

# Deploy
docker compose up -d

# Test
ssh -p 2222 localhost
```

### Server Setup (Ubuntu/Debian)

```bash
# Run setup script on fresh server
curl -fsSL https://raw.githubusercontent.com/mohakbajaj/mohak-tui/main/scripts/setup-server.sh | bash

# Configure environment
nano ~/mohak-tui/.env

# Start
sudo systemctl start mohak-tui
```

### Manual Deployment

**1. Build images locally:**

```bash
docker compose build
```

**2. Or use pre-built images:**

```bash
# Pull from GitHub Container Registry and use the production compose file
docker pull ghcr.io/mohakbajaj/mohak-tui/tui-server:latest
docker compose -f docker/docker-compose.prod.yml up -d
```

### CI/CD with GitHub Actions

The repository includes CI/CD workflows:

- **CI** (`.github/workflows/ci.yml`) - Runs on push to main / PRs
  - Lints TypeScript
  - Builds and tests the Go server
  - Builds Docker images

- **Deploy** (`.github/workflows/deploy.yml`) - Runs on tags (`v*`) or manual trigger
  - Builds and pushes the TUI server image (amd64/arm64)
  - Pushes to GitHub Container Registry
  - Deploys to `~/mohak-tui` on production server via SSH

**To trigger deployment:**

```bash
# Create and push a tag
git tag v1.0.0
git push origin v1.0.0

# Or manually from GitHub Actions UI
```

**Required Secrets:**

| Secret           | Description                    |
| ---------------- | ------------------------------ |
| `DEPLOY_HOST`    | Production server hostname/IP  |
| `DEPLOY_USER`    | SSH username                   |
| `DEPLOY_SSH_KEY` | SSH private key (full content) |

**First-time server setup:**

After first deployment, SSH into your server and configure:

```bash
ssh user@your-server
cd ~/mohak-tui
nano .env  # Add AI_GATEWAY_API_KEY and other secrets
```

### Production Configuration

**docker-compose.yml environment:**

```yaml
services:
  tui-server:
    ports:
      - "22:2222" # Use port 22 for production
    environment:
      - AI_GATEWAY_API_KEY=${AI_GATEWAY_API_KEY}
      - LOG_FORMAT=json
```

**Resource limits (recommended):**

| Service    | Memory | CPU |
| ---------- | ------ | --- |
| TUI Server | 512MB  | 1.0 |

### Monitoring

**Health endpoints:**

```bash
# TUI Server (via netcat)
nc -z localhost 2222
```

**Logs:**

```bash
# View logs
cd ~/mohak-tui
docker compose -f docker-compose.prod.yml logs -f

# JSON logs can be shipped to:
# - Datadog
# - Grafana Loki
# - ELK Stack
```

**PostHog Dashboard:**

Events are tracked automatically. Create dashboards for:

- Session duration
- Views visited
- Chat usage
- Error rates

SSH uses its own encryption, so no additional TLS termination is required for the public interface.

## Project Scripts

| Script                    | Description               |
| ------------------------- | ------------------------- |
| `scripts/deploy.sh`       | Build and deploy locally  |
| `scripts/setup-server.sh` | Setup fresh Ubuntu server |

```bash
# Deploy locally
./scripts/deploy.sh

# Deploy to staging
./scripts/deploy.sh staging
```

## License

MIT
