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
PORT=8080 ./sygy-server
```

## Protocol

All messages are JSON over WebSocket at `ws://<host>:<port>/ws`.

### Client → Server

| type | fields | description |
|------|--------|-------------|
| `join` | `room`, `index`, `participants` | Join a room (index is 1-based) |
| `signal_done` | `room`, `index`, `participants` | Mark self as done |
| `post_raid_id` | `room`, `index`, `participants`, `battle_id` | Share battle ID |
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

```bash
scp sygy-server user@<vps-ip>:/usr/local/bin/
ssh user@<vps-ip> 'systemctl enable --now sygy-server'
```
