#!/bin/bash
set -euo pipefail

# Server Setup Script for mohak.sh
# Run this on a fresh Ubuntu/Debian server

REPO_URL="https://github.com/mohakbajaj/mohak-tui.git"
INSTALL_DIR="$HOME/mohak-tui"

echo "═══════════════════════════════════════════════"
echo "  mohak.sh Server Setup"
echo "═══════════════════════════════════════════════"

# Update system
echo ""
echo "▶ Updating system..."
sudo apt-get update
sudo apt-get upgrade -y

# Install Docker
echo ""
echo "▶ Installing Docker..."
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com | sh
    sudo usermod -aG docker $USER
    echo "Docker installed. You may need to log out and back in."
fi

# Install Docker Compose
echo ""
echo "▶ Installing Docker Compose..."
if ! command -v docker-compose &> /dev/null; then
    sudo apt-get install -y docker-compose-plugin
fi

# Install git if needed
if ! command -v git &> /dev/null; then
    sudo apt-get install -y git
fi

# Clone or update repository
echo ""
echo "▶ Setting up application..."
if [[ -d "$INSTALL_DIR" ]]; then
    echo "Updating existing installation..."
    cd "$INSTALL_DIR"
    git pull origin main
else
    echo "Cloning repository..."
    git clone "$REPO_URL" "$INSTALL_DIR"
    cd "$INSTALL_DIR"
fi

# Copy .env from example if it doesn't exist
echo ""
echo "▶ Setting up environment..."
if [[ ! -f "$INSTALL_DIR/.env" ]]; then
    cp "$INSTALL_DIR/.env.example" "$INSTALL_DIR/.env"
    echo "Created .env from .env.example"
    echo "⚠  Please edit $INSTALL_DIR/.env with your API keys"
else
    echo ".env already exists, skipping..."
fi

# Setup firewall
echo ""
echo "▶ Configuring firewall..."
if command -v ufw &> /dev/null; then
    sudo ufw allow 22/tcp comment "SSH/TUI"
    sudo ufw allow 80/tcp comment "HTTP"
    sudo ufw allow 443/tcp comment "HTTPS"
    sudo ufw --force enable
fi

# Create systemd service for auto-start
echo ""
echo "▶ Creating systemd service..."
sudo tee /etc/systemd/system/mohak-tui.service > /dev/null << EOF
[Unit]
Description=mohak.sh TUI Portfolio
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
User=$USER
WorkingDirectory=$INSTALL_DIR
ExecStart=/usr/bin/docker compose -f docker/docker-compose.prod.yml up -d
ExecStop=/usr/bin/docker compose -f docker/docker-compose.prod.yml down
TimeoutStartSec=300

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable mohak-tui

echo ""
echo "═══════════════════════════════════════════════"
echo "  Server Setup Complete!"
echo ""
echo "  Next steps:"
echo "  1. Edit $INSTALL_DIR/.env with your API keys:"
echo "     nano $INSTALL_DIR/.env"
echo ""
echo "  2. Start the service:"
echo "     sudo systemctl start mohak-tui"
echo ""
echo "  3. Test connection:"
echo "     ssh -p 22 localhost"
echo "═══════════════════════════════════════════════"
