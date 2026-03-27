# VPS Setup Guide

Full procedure to bring up a new tavrn.sh server from a blank Ubuntu 22.04 VPS.
Run every command as **root** unless the step says otherwise.

---

## 1. System prerequisites

```bash
apt update && apt upgrade -y
apt install -y git curl wget ufw fail2ban libcap2-bin
```

---

## 2. Install Go

```bash
GO_VERSION=1.24.1
curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf /tmp/go.tar.gz
rm /tmp/go.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile.d/go.sh
source /etc/profile.d/go.sh
go version
```

---

## 3. Install Caddy

```bash
apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' \
  | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' \
  | tee /etc/apt/sources.list.d/caddy-stable.list
apt update && apt install -y caddy
```

---

## 4. Create the tavrn user and clone the repo

```bash
useradd -m -s /bin/bash tavrn
su - tavrn -c "
  git clone https://github.com/neur0map/tavrn.git ~/tavrn
"
```

---

## 5. Create the environment file (secrets)

As the `tavrn` user (or root writing to `/home/tavrn/tavrn/.env`):

```bash
cat > /home/tavrn/tavrn/.env <<'EOF'
JAMENDO_CLIENT_ID=your_jamendo_id_here
EOF
chmod 600 /home/tavrn/tavrn/.env
chown tavrn:tavrn /home/tavrn/tavrn/.env
```

---

## 6. Build the server binary

```bash
su - tavrn -c "
  export PATH=\$PATH:/usr/local/go/bin
  cd ~/tavrn
  go build -o tavrn      ./cmd/tavrn-admin
  go build -o tavrn-client ./cmd/tavrn
"
```

---

## 7. Install deploy assets

Copy each file from this `deploy/` directory into its system location.

### Systemd service

```bash
cp /home/tavrn/tavrn/deploy/tavrn.service /etc/systemd/system/tavrn.service
mkdir -p /etc/systemd/system/tavrn.service.d
cp /home/tavrn/tavrn/deploy/tavrn.service.d/hardening.conf \
   /etc/systemd/system/tavrn.service.d/hardening.conf
systemctl daemon-reload
```

### Root helper scripts

```bash
cp /home/tavrn/tavrn/deploy/sbin/tavrn-finalize-update /usr/local/sbin/tavrn-finalize-update
cp /home/tavrn/tavrn/deploy/sbin/tavrn-refresh-client  /usr/local/sbin/tavrn-refresh-client
chmod 755 /usr/local/sbin/tavrn-finalize-update
chmod 755 /usr/local/sbin/tavrn-refresh-client
chown root:root /usr/local/sbin/tavrn-finalize-update /usr/local/sbin/tavrn-refresh-client
```

### Sudoers rule

```bash
cp /home/tavrn/tavrn/deploy/sudoers.d/tavrn /etc/sudoers.d/tavrn
chmod 440 /etc/sudoers.d/tavrn
chown root:root /etc/sudoers.d/tavrn
visudo -c   # verify no syntax errors
```

### Caddy config and web root

```bash
mkdir -p /var/www/tavrn
cp /home/tavrn/tavrn/deploy/Caddyfile /etc/caddy/Caddyfile
cp /home/tavrn/tavrn/deploy/www/goget.html /var/www/tavrn/goget.html
# Copy website landing page if desired
cp /home/tavrn/tavrn/website/* /var/www/tavrn/
chown -R caddy:caddy /var/www/tavrn
```

---

## 8. Run the finalize script to set capabilities and symlinks

```bash
/usr/local/sbin/tavrn-finalize-update
```

This sets `cap_net_bind_service` on the binary, creates `/usr/local/bin/tavrn` and
`/usr/local/bin/tavrn-client` symlinks, and starts the service.

---

## 9. SSH hardening — move admin shell to port 2222

Edit `/etc/ssh/sshd_config`:

```bash
# Change or add:
Port 2222
PasswordAuthentication no
PermitRootLogin no
MaxAuthTries 3
LoginGraceTime 30
```

Then restart sshd:

```bash
systemctl restart ssh
```

> **Before closing your current session**, open a second terminal and confirm you can
> reach port 2222 with your key. Port 22 is now owned by the tavrn Go process.

### Lock admin SSH to specific IPs (recommended)

Add to `/etc/ufw/before.rules` or use UFW's limit:

```bash
# Allow admin SSH only from known IPs:
ufw allow from <YOUR_IP>     to any port 2222
ufw allow from <OTHER_IP>    to any port 2222
```

Then deny it by default (done in step 10 below).

---

## 10. Firewall (UFW)

```bash
ufw default deny incoming
ufw default allow outgoing
ufw allow 22/tcp      # tavrn SSH server (users)
ufw allow 80/tcp      # Caddy HTTP (redirects to HTTPS)
ufw allow 443/tcp     # Caddy HTTPS
# Admin SSH (add per-IP rules above before this deny):
ufw allow 2222/tcp    # or restrict to specific IPs as shown above
ufw enable
ufw status verbose
```

---

## 11. fail2ban

The default install protects SSH. Enable it:

```bash
systemctl enable --now fail2ban
# Verify it's watching the right port
fail2ban-client status sshd
```

For port 2222, create `/etc/fail2ban/jail.d/tavrn-admin.conf`:

```ini
[sshd]
enabled  = true
port     = 2222
maxretry = 3
bantime  = 3600
```

```bash
systemctl restart fail2ban
```

---

## 12. Enable and start services

```bash
systemctl enable --now tavrn
systemctl enable --now caddy
systemctl status tavrn
systemctl status caddy
```

---

## 13. Verify the deployment

Run the verification script from your local machine (see `deploy/verify.sh`):

```bash
bash deploy/verify.sh tavrn.sh
```

Or check manually:

```bash
# Vanity import
curl -s "https://tavrn.sh/?go-get=1" | grep go-import

# SSH port 22 (tavrn)
ssh -o ConnectTimeout=5 -o BatchMode=yes tavrn.sh 2>&1 | head -1

# Admin port 2222
ssh -p 2222 -o ConnectTimeout=5 tavrn@<VPS_IP> "systemctl status tavrn --no-pager"
```

---

## Ongoing maintenance

### Update server + restart

As the `tavrn` user on the VPS (via admin SSH):

```bash
~/tavrn/tavrn --update
```

### Update client binary only (no downtime)

```bash
~/tavrn/tavrn --update-client
```

### Send a message to all connected users

```bash
~/tavrn/tavrn --message "Maintenance in 5 minutes"
```

### Purge all data

```bash
~/tavrn/tavrn purge
```
