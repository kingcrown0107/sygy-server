# sygy-server

GoGo signal server for SYGY — WebSocket-based real-time lamp sync for GBF co-op.

## Build

```bash
go build -o sygy-server .

# Cross-compile for Linux (ConoHa VPS)
GOOS=linux GOARCH=amd64 go build -o sygy-server .
```

## Run

```bash
HOST=0.0.0.0 PORT=8080 ./sygy-server
```

Native SYGY clients connect without a browser `Origin` header and are accepted by default.
Browser-originated WebSocket connections are rejected unless the exact origin is listed in
`ALLOWED_ORIGINS`.

```bash
ALLOWED_ORIGINS=https://sygy.example.com HOST=0.0.0.0 PORT=8080 ./sygy-server
```

## Protocol

All messages are JSON over WebSocket at `ws://<host>:<port>/ws`.

### Client → Server

| type | fields | description |
|------|--------|-------------|
| `join` | `room`, `index`, `participants` | Join a room (index is 1-based) |
| `signal_done` | `room`, `index`, `participants` | Mark self as done. `index` and `participants` may be omitted after `join`. |
| `post_raid_id` | `room`, `index`, `participants`, `battle_id` or `raid_id` | Share battle ID. `index` and `participants` may be omitted after `join`. |
| `reset_lamps` | `room` | Reset all lamps |

### Server → Client

| type | description |
|------|-------------|
| `room_state` | Current lamp/raid state broadcast to all in room |
| `all_done` | Fired once when all participants are done |

### State payload

```json
{
  "type": "room_state",
  "room": 1,
  "participant_count": 4,
  "lamps": [true, false, true, false],
  "raid_ids": ["abc123", "", "def456", ""],
  "all_done_at": null
}
```

## Deploy (ConoHa VPS)

See [docs/vps-deployment.md](docs/vps-deployment.md) for the current hardened VPS layout.

```bash
go test ./...
go build -trimpath -ldflags='-s -w' -o sygy-server .
install -o root -g root -m 0755 sygy-server /opt/sygy-server/sygy-server
systemctl restart sygy-server.service
```
