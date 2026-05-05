# VPS Deployment

This document records the production VPS layout used for SYGY.

## Runtime

- OS: Ubuntu 24.04 LTS
- Public IPv4: `160.251.173.17`
- WebSocket endpoint: `ws://160.251.173.17:8080/ws`
- Health endpoint: `http://160.251.173.17:8080/health`
- SSH port: `1415`
- Service name: `sygy-server.service`
- Binary path: `/opt/sygy-server/sygy-server`
- Working directory: `/opt/sygy-server`
- Runtime user: `sygy-server`

## Build

```bash
go test ./...
go build -trimpath -ldflags='-s -w' -o sygy-server .
```

## systemd

The service should run as a dedicated unprivileged user, not as root.

```ini
[Unit]
Description=sygy-server
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=sygy-server
Group=sygy-server
WorkingDirectory=/opt/sygy-server
Environment=HOST=0.0.0.0
Environment=PORT=8080
ExecStart=/opt/sygy-server/sygy-server
Restart=always
RestartSec=5
NoNewPrivileges=true
PrivateTmp=true
PrivateDevices=true
ProtectSystem=strict
ProtectHome=true
ProtectClock=true
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectKernelLogs=true
ProtectControlGroups=true
RestrictSUIDSGID=true
RestrictRealtime=true
LockPersonality=true
MemoryDenyWriteExecute=true
SystemCallArchitectures=native
RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX
CapabilityBoundingSet=
AmbientCapabilities=
UMask=0077

[Install]
WantedBy=multi-user.target
```

Apply changes:

```bash
systemctl daemon-reload
systemctl restart sygy-server.service
systemctl status sygy-server.service --no-pager
```

## Firewall

UFW should allow only SSH and the SYGY WebSocket service.

```bash
ufw allow 1415/tcp
ufw allow 8080/tcp comment 'sygy-server websocket'
ufw status numbered
```

Expected exposed listeners:

- `0.0.0.0:1415` / `[::]:1415` for SSH socket activation
- `*:8080` for `sygy-server`

UDP 123 should not be exposed. Use `systemd-timesyncd` for time sync.

## SSH

Expected effective SSH settings:

```text
port 1415
permitrootlogin no
passwordauthentication no
kbdinteractiveauthentication no
maxauthtries 3
x11forwarding no
```

Fail2ban should monitor SSH on port `1415`.

## Security Notes

- The WebSocket server accepts clients without an `Origin` header so the native SYGY `ClientWebSocket` client works.
- Browser-originated WebSocket connections are rejected unless the exact origin is listed in `ALLOWED_ORIGINS`.
- If a browser UI is added later, set `ALLOWED_ORIGINS` as a comma-separated allowlist, for example:

```ini
Environment=ALLOWED_ORIGINS=https://sygy.example.com
```

## Verification

```bash
go test ./...
curl -i http://127.0.0.1:8080/health
curl -i http://160.251.173.17:8080/health
systemctl --failed --no-pager
systemd-analyze security sygy-server.service --no-pager
ss -tulpn
```

The deployed hardened service should report an overall systemd exposure level around `3.7 OK`.
