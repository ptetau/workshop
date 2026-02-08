# Deploying Workshop to Production

This guide walks through deploying Workshop Jiu Jitsu CRM to your OVH VPS.

## Architecture

```
Internet → Caddy (HTTPS :443) → Workshop app (:8080) → SQLite (workshop.db)
```

- **Caddy** handles HTTPS (auto Let's Encrypt) and reverse proxies to the app
- **Workshop** is a single Go binary, runs as a systemd service
- **SQLite** database file lives at `/opt/workshop/workshop.db`
- **GitHub Actions** builds and deploys on manual trigger

## One-Time VPS Setup

### Step 1: Generate an SSH deploy key

On your local machine, generate a key pair used **only** for deployments:

```bash
ssh-keygen -t ed25519 -f ~/.ssh/workshop_deploy -C "github-deploy" -N ""
```

This creates two files:
- `~/.ssh/workshop_deploy` (private key — goes to GitHub)
- `~/.ssh/workshop_deploy.pub` (public key — goes to VPS)

### Step 2: Set up VPS access

SSH into your VPS as root and add the deploy key:

```bash
ssh root@51.255.201.85
```

Create a deploy user with sudo access:

```bash
# Create deploy user
useradd -m -s /bin/bash deploy
echo "deploy ALL=(ALL) NOPASSWD: /usr/bin/mv,/usr/bin/chown,/usr/bin/chmod,/usr/bin/systemctl" > /etc/sudoers.d/deploy
chmod 440 /etc/sudoers.d/deploy

# Add deploy key
mkdir -p /home/deploy/.ssh
# Paste the contents of ~/.ssh/workshop_deploy.pub into this file:
nano /home/deploy/.ssh/authorized_keys
chmod 700 /home/deploy/.ssh
chmod 600 /home/deploy/.ssh/authorized_keys
chown -R deploy:deploy /home/deploy/.ssh
```

### Step 3: Run the setup script

From your local machine:

```bash
ssh root@51.255.201.85 'bash -s' < deploy/setup.sh
```

This installs Caddy, creates the workshop system user, sets up the app directory, and configures the firewall.

### Step 4: Configure Caddy

Edit `deploy/Caddyfile` and replace `YOUR_DOMAIN` with your actual domain.

If you **don't have a domain yet**, use this temporary Caddyfile:

```
:80 {
    reverse_proxy localhost:8080
}
```

Copy it to the VPS:

```bash
scp deploy/Caddyfile root@51.255.201.85:/etc/caddy/Caddyfile
ssh root@51.255.201.85 'systemctl restart caddy'
```

### Step 5: Install the systemd service

```bash
scp deploy/workshop.service root@51.255.201.85:/etc/systemd/system/workshop.service
ssh root@51.255.201.85 'systemctl daemon-reload && systemctl enable workshop'
```

### Step 6: Create the app directory structure

```bash
ssh root@51.255.201.85 'mkdir -p /opt/workshop/internal/adapters/http/templates /opt/workshop/static && chown -R workshop:workshop /opt/workshop'
```

### Step 7: Add GitHub Secrets

Go to https://github.com/ptetau/workshop/settings/secrets/actions and add:

| Secret | Value |
|--------|-------|
| `VPS_HOST` | `51.255.201.85` |
| `VPS_USER` | `deploy` |
| `VPS_SSH_KEY` | Contents of `~/.ssh/workshop_deploy` (the **private** key) |

To copy the private key:

```bash
cat ~/.ssh/workshop_deploy | clip    # Windows
cat ~/.ssh/workshop_deploy | pbcopy  # macOS
```

## Deploying

1. Go to https://github.com/ptetau/workshop/actions/workflows/deploy.yml
2. Click **"Run workflow"**
3. Type `deploy` in the confirmation field
4. Click the green **"Run workflow"** button

The workflow will:
1. Run all tests
2. Build a Linux binary
3. Upload the binary, static files, and templates to the VPS
4. Restart the service

## Monitoring

### Check if the app is running

```bash
ssh deploy@51.255.201.85 'sudo systemctl status workshop'
```

### View logs

```bash
ssh deploy@51.255.201.85 'sudo journalctl -u workshop -f'
```

### View Caddy logs

```bash
ssh deploy@51.255.201.85 'sudo journalctl -u caddy -f'
```

### Restart the app

```bash
ssh deploy@51.255.201.85 'sudo systemctl restart workshop'
```

## Database Backups

The SQLite database is at `/opt/workshop/workshop.db`. Back it up periodically:

```bash
ssh deploy@51.255.201.85 'sqlite3 /opt/workshop/workshop.db ".backup /opt/workshop/backup-$(date +%Y%m%d).db"'
```

## Troubleshooting

| Problem | Fix |
|---------|-----|
| App won't start | Check logs: `journalctl -u workshop -n 50` |
| 502 Bad Gateway | App isn't running — check systemd status |
| HTTPS not working | Ensure your domain's DNS A record points to `51.255.201.85` |
| Deploy fails at SSH | Check that `VPS_SSH_KEY` secret has the full private key including `-----BEGIN/END-----` lines |
| Permission denied | Ensure deploy user has correct sudoers config |
