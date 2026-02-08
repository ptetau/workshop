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

echo "=== 2. Install security tooling ==="
apt-get install -y \
    debian-keyring debian-archive-keyring apt-transport-https curl \
    fail2ban \
    unattended-upgrades \
    apt-listchanges

echo "=== 3. Enable unattended security upgrades ==="
cat > /etc/apt/apt.conf.d/20auto-upgrades << 'EOF'
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
APT::Periodic::AutocleanInterval "7";
EOF

echo "=== 4. Configure fail2ban ==="
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5

[sshd]
enabled = true
port = ssh
logpath = %(sshd_log)s
maxretry = 3
bantime = 86400
EOF
systemctl enable fail2ban
systemctl restart fail2ban

echo "=== 5. Harden SSH ==="
# Backup original config
cp /etc/ssh/sshd_config /etc/ssh/sshd_config.bak

# Apply hardened SSH config
cat > /etc/ssh/sshd_config.d/99-hardened.conf << 'EOF'
# Disable root login
PermitRootLogin no

# Disable password authentication (key-only)
PasswordAuthentication no
ChallengeResponseAuthentication no

# Disable empty passwords
PermitEmptyPasswords no

# Use only protocol 2
Protocol 2

# Limit authentication attempts
MaxAuthTries 3
LoginGraceTime 20

# Idle timeout (5 min)
ClientAliveInterval 300
ClientAliveCountMax 0

# Disable X11 and agent forwarding
X11Forwarding no
AllowAgentForwarding no

# Restrict to deploy user only
AllowUsers deploy
EOF
# Validate config before restarting
sshd -t && systemctl restart sshd
echo "SSH hardened: root login disabled, password auth disabled, key-only access for 'deploy' user"

echo "=== 6. Install Caddy ==="
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
apt-get update
apt-get install -y caddy

echo "=== 7. Create deploy user ==="
if ! id -u deploy &>/dev/null; then
    useradd -m -s /bin/bash deploy
fi
# Restricted sudo: only the specific commands needed for deploy
cat > /etc/sudoers.d/deploy << 'EOF'
deploy ALL=(ALL) NOPASSWD: /usr/bin/mv /opt/workshop/workshop.new /opt/workshop/workshop
deploy ALL=(ALL) NOPASSWD: /usr/bin/chown workshop\:workshop /opt/workshop/workshop
deploy ALL=(ALL) NOPASSWD: /usr/bin/chmod +x /opt/workshop/workshop
deploy ALL=(ALL) NOPASSWD: /usr/bin/systemctl restart workshop
deploy ALL=(ALL) NOPASSWD: /usr/bin/systemctl is-active workshop
deploy ALL=(ALL) NOPASSWD: /usr/bin/journalctl -u workshop *
EOF
chmod 440 /etc/sudoers.d/deploy
visudo -c  # validate sudoers syntax

echo "=== 8. Create workshop system user ==="
if ! id -u workshop &>/dev/null; then
    useradd --system --home /opt/workshop --shell /usr/sbin/nologin workshop
fi

echo "=== 9. Create application directory ==="
mkdir -p /opt/workshop/internal/adapters/http/templates
mkdir -p /opt/workshop/static
mkdir -p /var/log/caddy
chown -R workshop:workshop /opt/workshop
chmod 750 /opt/workshop
chown -R caddy:caddy /var/log/caddy

echo "=== 10. Configure firewall ==="
ufw default deny incoming
ufw default allow outgoing
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable

echo "=== 11. Kernel hardening (sysctl) ==="
cat > /etc/sysctl.d/99-security.conf << 'EOF'
# Prevent IP spoofing
net.ipv4.conf.all.rp_filter = 1
net.ipv4.conf.default.rp_filter = 1

# Ignore ICMP redirects
net.ipv4.conf.all.accept_redirects = 0
net.ipv6.conf.all.accept_redirects = 0

# Don't send ICMP redirects
net.ipv4.conf.all.send_redirects = 0

# Ignore broadcast ping
net.ipv4.icmp_echo_ignore_broadcasts = 1

# Enable TCP SYN cookies (SYN flood protection)
net.ipv4.tcp_syncookies = 1

# Log martian packets
net.ipv4.conf.all.log_martians = 1
EOF
sysctl -p /etc/sysctl.d/99-security.conf

echo ""
echo "=== Setup complete ==="
echo ""
echo "IMPORTANT: Before logging out, verify you can SSH as deploy user from another terminal!"
echo "  ssh deploy@$(hostname -I | awk '{print $1}')"
echo ""
echo "Next steps:"
echo "  1. Add deploy user's SSH public key: /home/deploy/.ssh/authorized_keys"
echo "  2. Copy deploy/Caddyfile to /etc/caddy/Caddyfile (edit YOUR_DOMAIN first)"
echo "  3. Run: systemctl restart caddy"
echo "  4. Copy deploy/workshop.service to /etc/systemd/system/workshop.service"
echo "  5. Run: systemctl daemon-reload && systemctl enable workshop"
echo "  6. Set up GitHub secrets (see deploy/DEPLOY.md)"
echo "  7. Generate CSRF key: openssl rand -hex 32"
echo "  8. Set WORKSHOP_CSRF_KEY in workshop.service environment"
