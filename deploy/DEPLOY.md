# Deploying Workshop to Production

This guide walks through deploying Workshop Jiu Jitsu CRM to your OVH VPS.

## Architecture

```
Internet → Caddy (HTTPS :443) → Workshop app (127.0.0.1:8080) → SQLite (workshop.db)
```

- **Caddy** handles HTTPS (auto Let's Encrypt), security headers, and reverse proxies to the app
- **Workshop** is a single Go binary, runs as a sandboxed systemd service, binds to **localhost only**
- **SQLite** database file lives at `/opt/workshop/workshop.db` (owner-only permissions)
- **GitHub Actions** builds and deploys on manual trigger

## Security Overview

| Layer | Protection |
|-------|-----------|
| **Network** | UFW firewall (SSH/80/443 only), Caddy TLS, HSTS, app bound to localhost |
| **SSH** | Key-only auth, root login disabled, fail2ban (24h ban after 3 failures), deploy user only |
| **OS** | Unattended security upgrades, kernel hardening (sysctl), SYN flood protection |
| **Process** | Systemd sandbox (PrivateTmp, PrivateDevices, ProtectKernel*, no capabilities, syscall filter) |
| **Application** | CSRF key from env var, rate limiting, no hardcoded secrets, synthetic data disabled in prod |
| **CI/CD** | Actions pinned by SHA, SSH host key verified, health check after deploy, restricted sudoers |

## One-Time VPS Setup

### Step 1: Generate an SSH deploy key

On your local machine, generate a key pair used **only** for deployments:

```bash
ssh-keygen -t ed25519 -f ~/.ssh/workshop_deploy -C "github-deploy" -N ""
```

This creates two files:
- `~/.ssh/workshop_deploy` — private key (goes to GitHub Secrets)
- `~/.ssh/workshop_deploy.pub` — public key (goes to VPS)

### Step 2: Add deploy key to VPS

SSH into your VPS as root (last time you'll use root):

```bash
ssh root@51.255.201.85

# Create deploy user (the setup script also does this, but we need it for the key)
useradd -m -s /bin/bash deploy

# Add deploy key
mkdir -p /home/deploy/.ssh
# Paste the contents of ~/.ssh/workshop_deploy.pub:
nano /home/deploy/.ssh/authorized_keys
chmod 700 /home/deploy/.ssh
chmod 600 /home/deploy/.ssh/authorized_keys
chown -R deploy:deploy /home/deploy/.ssh
```

### Step 3: Run the setup script

**⚠️ IMPORTANT:** Before running this, verify you can SSH as `deploy` from another terminal. The setup script disables root login and password auth.

```bash
ssh root@51.255.201.85 'bash -s' < deploy/setup.sh
```

The setup script:
1. Updates system packages
2. Installs **fail2ban** (SSH brute-force protection — 24h ban after 3 failures)
3. Installs **unattended-upgrades** (automatic security patches)
4. **Hardens SSH** — disables root login, disables password auth, key-only for `deploy` user
5. Installs **Caddy** reverse proxy
6. Creates **deploy** user with restricted sudo (only deploy-specific commands)
7. Creates **workshop** system user (no login shell, runs the app)
8. Sets up app directory with correct permissions
9. Configures **UFW firewall** (SSH, 80, 443 only, deny all else)
10. Applies **kernel hardening** (anti-spoofing, SYN cookies, ICMP restrictions)

### Step 4: Generate production secrets

On your local machine, generate the required secrets:

```bash
# CSRF key (64 hex chars = 32 bytes)
openssl rand -hex 32
```

Save this value — you'll need it in Step 6.

### Step 5: Configure the environment file

SSH as deploy and create the env file:

```bash
ssh deploy@51.255.201.85

sudo tee /opt/workshop/.env << 'EOF'
WORKSHOP_ENV=production
WORKSHOP_ADDR=127.0.0.1:8080
WORKSHOP_CSRF_KEY=<paste-your-64-hex-char-key-here>
WORKSHOP_ADMIN_EMAIL=admin@workshop.co.nz
WORKSHOP_ADMIN_PASSWORD=<choose-a-strong-password>
EOF

sudo chown workshop:workshop /opt/workshop/.env
sudo chmod 600 /opt/workshop/.env
```

Then update the systemd service to use it:

```bash
# Edit the service file to uncomment EnvironmentFile and remove inline Environment lines:
sudo sed -i 's|^#.*EnvironmentFile=|EnvironmentFile=|' /etc/systemd/system/workshop.service
sudo systemctl daemon-reload
```

**⚠️ Change the admin password immediately after first login!**

### Step 6: Configure Caddy

Edit `deploy/Caddyfile` and replace `YOUR_DOMAIN` with your actual domain.

If you **don't have a domain yet**, use this temporary Caddyfile:

```
:80 {
    reverse_proxy localhost:8080
}
```

Copy it to the VPS:

```bash
scp deploy/Caddyfile deploy@51.255.201.85:/tmp/Caddyfile
ssh deploy@51.255.201.85 'sudo mv /tmp/Caddyfile /etc/caddy/Caddyfile && sudo systemctl restart caddy'
```

### Step 7: Install the systemd service

```bash
scp deploy/workshop.service deploy@51.255.201.85:/tmp/workshop.service
ssh deploy@51.255.201.85 'sudo mv /tmp/workshop.service /etc/systemd/system/workshop.service && sudo systemctl daemon-reload && sudo systemctl enable workshop'
```

### Step 8: Add GitHub Secrets

Go to https://github.com/ptetau/workshop/settings/secrets/actions and add:

| Secret | Value |
|--------|-------|
| `VPS_HOST` | `51.255.201.85` |
| `VPS_USER` | `deploy` |
| `VPS_SSH_KEY` | Contents of `~/.ssh/workshop_deploy` (the **private** key, including `-----BEGIN/END-----` lines) |

To copy the private key:

```bash
cat ~/.ssh/workshop_deploy | clip    # Windows (PowerShell)
cat ~/.ssh/workshop_deploy | pbcopy  # macOS
```

## Deploying

1. Go to https://github.com/ptetau/workshop/actions/workflows/deploy.yml
2. Click **"Run workflow"**
3. Type `deploy` in the confirmation field
4. Click the green **"Run workflow"** button

The workflow will:
1. Run all tests
2. Build a stripped Linux binary (`-ldflags="-s -w" -trimpath`)
3. Upload the binary, static files, and templates via rsync
4. Atomically swap the binary and restart the service
5. Verify the service is active AND responding to HTTP requests

## Monitoring

```bash
# Service status
ssh deploy@51.255.201.85 'sudo systemctl status workshop'

# Live application logs
ssh deploy@51.255.201.85 'sudo journalctl -u workshop -f'

# Caddy (reverse proxy) logs
ssh deploy@51.255.201.85 'sudo journalctl -u caddy -f'

# fail2ban status (SSH brute-force bans)
ssh deploy@51.255.201.85 'sudo fail2ban-client status sshd'

# Restart the app
ssh deploy@51.255.201.85 'sudo systemctl restart workshop'
```

## Database Backups

The SQLite database is at `/opt/workshop/workshop.db`. Back it up periodically:

```bash
ssh deploy@51.255.201.85 'sqlite3 /opt/workshop/workshop.db ".backup /opt/workshop/backup-$(date +%Y%m%d).db"'
```

**Consider:** setting up a cron job on the VPS to automate daily backups.

## Security Checklist

Run through this after initial setup:

- [ ] Can SSH as `deploy` user (key-only)
- [ ] Cannot SSH as `root`
- [ ] Cannot SSH with password
- [ ] fail2ban is running: `sudo fail2ban-client status sshd`
- [ ] UFW is active: `sudo ufw status`
- [ ] Only ports 22, 80, 443 are open
- [ ] App binds to `127.0.0.1:8080` (not reachable from internet directly)
- [ ] `WORKSHOP_ENV=production` is set
- [ ] `WORKSHOP_CSRF_KEY` is set (not the default)
- [ ] Admin password has been changed from default
- [ ] HTTPS is working (if domain is configured)
- [ ] HSTS header present: `curl -I https://YOUR_DOMAIN`
- [ ] Unattended-upgrades active: `sudo systemctl status unattended-upgrades`

## Troubleshooting

| Problem | Fix |
|---------|-----|
| App won't start | Check logs: `journalctl -u workshop -n 50` |
| `WORKSHOP_CSRF_KEY is required` | Set the key in `/opt/workshop/.env` — see Step 5 |
| 502 Bad Gateway | App isn't running — check systemd status |
| HTTPS not working | Ensure your domain's DNS A record points to `51.255.201.85` |
| Deploy fails at SSH | Check that `VPS_SSH_KEY` secret has the full private key including `-----BEGIN/END-----` lines |
| Permission denied | Ensure deploy user has correct sudoers config |
| Locked out of SSH | Access VPS via OVH console (KVM), fix `/etc/ssh/sshd_config.d/99-hardened.conf` |
| Synthetic data in prod | Ensure `WORKSHOP_ENV=production` is set in `.env` |
