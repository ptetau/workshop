#!/bin/bash
# VPS Configuration Diagnostic Script for Bug #314
# Run this on the VPS as deploy user to check configuration

set -e

echo "=== VPS Configuration Diagnostic ==="
echo "Date: $(date)"
echo ""

# Check 1: SSH Key Authentication
echo "1. Checking SSH configuration..."
if [ -f ~/.ssh/authorized_keys ]; then
    KEY_COUNT=$(grep -c "ssh-" ~/.ssh/authorized_keys 2>/dev/null || echo "0")
    echo "   ✓ authorized_keys exists with $KEY_COUNT key(s)"
else
    echo "   ✗ authorized_keys missing!"
fi

# Check 2: Deploy user sudo access
echo ""
echo "2. Checking sudo access..."
if sudo -l 2>/dev/null | grep -q "NOPASSWD"; then
    echo "   ✓ sudo access configured"
else
    echo "   ✗ sudo access may be limited"
fi

# Check 3: Workshop service
echo ""
echo "3. Checking workshop service..."
if systemctl is-active workshop >/dev/null 2>&1; then
    echo "   ✓ workshop service is running"
else
    echo "   ✗ workshop service is not running!"
    echo "   Status: $(systemctl is-active workshop 2>&1)"
fi

if [ -f /etc/systemd/system/workshop.service ]; then
    echo "   ✓ workshop.service file exists"
else
    echo "   ✗ workshop.service file missing!"
fi

# Check 4: Application directory
echo ""
echo "4. Checking application directory..."
if [ -d /opt/workshop ]; then
    echo "   ✓ /opt/workshop exists"
    ls -la /opt/workshop/ | head -5
else
    echo "   ✗ /opt/workshop directory missing!"
fi

# Check 5: Environment file
echo ""
echo "5. Checking environment configuration..."
if [ -f /opt/workshop/.env ]; then
    echo "   ✓ .env file exists"
    # Check for required vars (don't show values)
    for var in WORKSHOP_ENV WORKSHOP_CSRF_KEY WORKSHOP_RESEND_KEY; do
        if grep -q "$var=" /opt/workshop/.env; then
            echo "   ✓ $var is set"
        else
            echo "   ✗ $var is MISSING"
        fi
    done
else
    echo "   ✗ .env file missing!"
fi

# Check 6: Database
echo ""
echo "6. Checking database..."
if [ -f /opt/workshop/workshop.db ]; then
    SIZE=$(du -h /opt/workshop/workshop.db | cut -f1)
    echo "   ✓ Database exists ($SIZE)"
else
    echo "   ! No database file (first deploy?)"
fi

# Check 7: Caddy
echo ""
echo "7. Checking Caddy..."
if systemctl is-active caddy >/dev/null 2>&1; then
    echo "   ✓ Caddy is running"
else
    echo "   ✗ Caddy is not running!"
fi

if [ -f /etc/caddy/Caddyfile ]; then
    echo "   ✓ Caddyfile exists"
else
    echo "   ✗ Caddyfile missing!"
fi

# Check 8: Firewall
echo ""
echo "8. Checking firewall..."
if sudo ufw status | grep -q "Status: active"; then
    echo "   ✓ UFW is active"
    echo "   Open ports:"
    sudo ufw status | grep "ALLOW" | head -5
else
    echo "   ! UFW is not active"
fi

# Check 9: fail2ban
echo ""
echo "9. Checking fail2ban..."
if systemctl is-active fail2ban >/dev/null 2>&1; then
    echo "   ✓ fail2ban is running"
    BANNED=$(sudo fail2ban-client status sshd 2>/dev/null | grep "Currently banned" | awk '{print $NF}' || echo "?")
    echo "   Currently banned IPs: $BANNED"
else
    echo "   ✗ fail2ban is not running!"
fi

# Check 10: HTTP health check
echo ""
echo "10. Testing local HTTP endpoint..."
if curl -sf http://127.0.0.1:8080/login >/dev/null 2>&1; then
    echo "   ✓ App responds to HTTP requests"
else
    echo "   ✗ App not responding on localhost:8080"
fi

echo ""
echo "=== Diagnostic Complete ==="
