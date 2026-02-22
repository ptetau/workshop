#!/bin/bash
# VPS Setup Fix Script for Bug #314
# Run this on the VPS as root or deploy user with sudo

set -e

echo "=== VPS Setup Fix Script ==="
echo "This script fixes common VPS configuration issues"
echo ""

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ] && ! sudo -n true 2>/dev/null; then
    echo "Error: This script must be run as root or with sudo access"
    exit 1
fi

# Function to run command with sudo if not root
run_cmd() {
    if [ "$EUID" -eq 0 ]; then
        "$@"
    else
        sudo "$@"
    fi
}

echo "Step 1: Creating deploy user (if missing)..."
if ! id "deploy" &>/dev/null; then
    run_cmd useradd -m -s /bin/bash deploy
    echo "✓ Created deploy user"
else
    echo "✓ Deploy user exists"
fi

echo ""
echo "Step 2: Setting up SSH directory for deploy user..."
run_cmd mkdir -p /home/deploy/.ssh
run_cmd chmod 700 /home/deploy/.ssh

echo ""
echo "Step 3: Creating workshop system user (if missing)..."
if ! id "workshop" &>/dev/null; then
    run_cmd useradd -r -s /bin/false workshop
    echo "✓ Created workshop system user"
else
    echo "✓ Workshop user exists"
fi

echo ""
echo "Step 4: Creating application directory..."
run_cmd mkdir -p /opt/workshop/backups
run_cmd chown workshop:workshop /opt/workshop
run_cmd chown workshop:workshop /opt/workshop/backups

echo ""
echo "Step 5: Setting up sudoers for deploy user..."
SUDOERS_FILE="/etc/sudoers.d/workshop-deploy"
if [ ! -f "$SUDOERS_FILE" ]; then
    cat > /tmp/workshop-sudoers << 'EOF'
# Workshop deployment sudoers config
deploy ALL=(root) NOPASSWD: /bin/systemctl restart workshop
deploy ALL=(root) NOPASSWD: /bin/systemctl stop workshop
deploy ALL=(root) NOPASSWD: /bin/systemctl start workshop
deploy ALL=(root) NOPASSWD: /bin/systemctl status workshop
deploy ALL=(root) NOPASSWD: /bin/systemctl is-active workshop
deploy ALL=(root) NOPASSWD: /bin/systemctl daemon-reload
deploy ALL=(root) NOPASSWD: /bin/mv /opt/workshop/workshop.new /opt/workshop/workshop
deploy ALL=(root) NOPASSWD: /bin/chown workshop\:workshop /opt/workshop/workshop
deploy ALL=(root) NOPASSWD: /bin/chmod +x /opt/workshop/workshop
EOF
    run_cmd mv /tmp/workshop-sudoers "$SUDOERS_FILE"
    run_cmd chmod 440 "$SUDOERS_FILE"
    echo "✓ Created sudoers file for deploy user"
else
    echo "✓ Sudoers file exists"
fi

echo ""
echo "Step 6: Checking/installing sqlite3..."
if ! command -v sqlite3 &>/dev/null; then
    run_cmd apt-get update -qq
    run_cmd apt-get install -y -qq sqlite3
    echo "✓ Installed sqlite3"
else
    echo "✓ sqlite3 already installed"
fi

echo ""
echo "Step 7: Installing Caddy (if missing)..."
if ! command -v caddy &>/dev/null; then
    run_cmd apt-get install -y -qq debian-keyring debian-archive-keyring apt-transport-https
    run_cmd curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | run_cmd apt-key add -
    run_cmd curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | run_cmd tee /etc/apt/sources.list.d/caddy-stable.list
    run_cmd apt-get update -qq
    run_cmd apt-get install -y -qq caddy
    run_cmd systemctl enable caddy
    echo "✓ Installed Caddy"
else
    echo "✓ Caddy already installed"
fi

echo ""
echo "Step 8: Installing fail2ban (if missing)..."
if ! command -v fail2ban-client &>/dev/null; then
    run_cmd apt-get install -y -qq fail2ban
    run_cmd systemctl enable fail2ban
    run_cmd systemctl start fail2ban
    echo "✓ Installed fail2ban"
else
    echo "✓ fail2ban already installed"
fi

echo ""
echo "Step 9: Creating systemd service file..."
SERVICE_FILE="/etc/systemd/system/workshop.service"
if [ ! -f "$SERVICE_FILE" ]; then
    cat > /tmp/workshop.service << 'EOF'
[Unit]
Description=Workshop Jiu Jitsu CRM
After=network.target

[Service]
Type=simple
User=workshop
Group=workshop
WorkingDirectory=/opt/workshop
Environment=WORKSHOP_ENV=production
Environment=WORKSHOP_ADDR=127.0.0.1:8080
# Load additional env vars from file if it exists
EnvironmentFile=-/opt/workshop/.env
ExecStart=/opt/workshop/workshop
Restart=always
RestartSec=5

# Security hardening
PrivateTmp=yes
PrivateDevices=yes
ProtectKernelTunables=yes
ProtectKernelModules=yes
ProtectControlGroups=yes
NoNewPrivileges=yes
RestrictSUIDSGID=yes
RestrictRealtime=yes
RestrictNamespaces=yes
LockPersonality=yes
MemoryDenyWriteExecute=yes
SystemCallFilter=@system-service
SystemCallErrorNumber=EPERM

[Install]
WantedBy=multi-user.target
EOF
    run_cmd mv /tmp/workshop.service "$SERVICE_FILE"
    run_cmd systemctl daemon-reload
    echo "✓ Created workshop.service"
else
    echo "✓ workshop.service exists"
fi

echo ""
echo "Step 10: Setting up UFW firewall..."
if command -v ufw &>/dev/null; then
    run_cmd ufw default deny incoming
    run_cmd ufw default allow outgoing
    run_cmd ufw allow ssh
    run_cmd ufw allow http
    run_cmd ufw allow https
    if ! run_cmd ufw status | grep -q "Status: active"; then
        echo "y" | run_cmd ufw enable
    fi
    echo "✓ UFW configured"
else
    echo "! UFW not installed, skipping"
fi

echo ""
echo "Step 11: Creating placeholder binary..."
if [ ! -f /opt/workshop/workshop ]; then
    cat > /tmp/placeholder.sh << 'EOF'
#!/bin/bash
echo "Workshop placeholder - replace with actual binary"
sleep 30
EOF
    run_cmd mv /tmp/placeholder.sh /opt/workshop/workshop
    run_cmd chown workshop:workshop /opt/workshop/workshop
    run_cmd chmod +x /opt/workshop/workshop
    echo "✓ Created placeholder binary"
else
    echo "✓ Binary exists"
fi

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Next steps:"
echo "1. Add your SSH public key to /home/deploy/.ssh/authorized_keys"
echo "2. Set up /opt/workshop/.env with required environment variables"
echo "3. Upload the actual workshop binary"
echo "4. Start the service: sudo systemctl start workshop"
echo ""
echo "To check status, run: sudo systemctl status workshop"
