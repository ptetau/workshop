#!/bin/bash
#
# One-time VPS setup script for Workshop Jiu Jitsu CRM.
# Run this as root on a fresh Ubuntu 25.04 VPS.
#
# Usage: ssh root@51.255.201.85 'bash -s' < deploy/setup.sh
#
set -euo pipefail

echo "=== 1. System update ==="
apt-get update && apt-get upgrade -y

echo "=== 2. Install Caddy ==="
apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
apt-get update
apt-get install -y caddy

echo "=== 3. Create workshop user ==="
if ! id -u workshop &>/dev/null; then
    useradd --system --home /opt/workshop --shell /usr/sbin/nologin workshop
fi

echo "=== 4. Create application directory ==="
mkdir -p /opt/workshop
mkdir -p /var/log/caddy
chown -R workshop:workshop /opt/workshop
chown -R caddy:caddy /var/log/caddy

echo "=== 5. Configure firewall ==="
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

echo "=== 6. Done ==="
echo ""
echo "Next steps:"
echo "  1. Copy deploy/Caddyfile to /etc/caddy/Caddyfile (edit YOUR_DOMAIN first)"
echo "  2. Run: systemctl restart caddy"
echo "  3. Copy deploy/workshop.service to /etc/systemd/system/workshop.service"
echo "  4. Run: systemctl daemon-reload && systemctl enable workshop"
echo "  5. Set up GitHub secrets (see deploy/DEPLOY.md)"
echo "  6. Trigger a deploy from GitHub Actions"
