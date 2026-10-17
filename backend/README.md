# Syncraft Backend

Syncraft is an anonymous, real-time collaborative coding workspace backend built with **Go**, **Gin**, and **WebSockets** (`github.com/coder/websocket`).

Users create ephemeral rooms, share a link, and collaboratively edit code in real-time without signing up, logging in, or creating accounts.

---

## Key Features & Constraints

- **Zero Database / Pure Server Memory**: No PostgreSQL, Redis, MongoDB, or SQLite. All room state, participant memberships, and CRDT update logs exist strictly in volatile server RAM.
- **Concurrency Safe**: Powered by fine-grained `sync.RWMutex` locks, non-blocking channels, and independent per-subscriber mutexes. **Zero global locks are held during network I/O**.
- **Yjs-Compatible Dual Protocol**:
  - Versioned **JSON envelopes** for room control, handshakes, awareness metadata, and error reporting.
  - High-performance **binary WebSocket frames** for opaque CRDT document updates.
- **Catch-Up & Synchronization Replay**: Newly joined participants automatically receive replayed update history before receiving live keystrokes.
- **Resource Exhaustion Protection**:
  - Bounded participant outbound queues with slow consumer detection and auto-disconnect.
  - Bounded in-memory update logs per room with deterministic backpressure (`UPDATE_LIMIT_EXCEEDED`).
  - IP-based token bucket rate limiting for HTTP endpoints.
  - Per-connection message rate limiting for WebSocket sessions.
  - Message size ceiling via WebSocket `ReadLimit`.
- **Anonymous Session Security**:
  - Unguessable room IDs generated with 128 bits of entropy via `crypto/rand`.
  - Server-assigned ephemeral participant IDs (`p_<hex>`).
  - Origin verification matching against configurable whitelist.
  - Never logs user code or raw document payloads.
- **Transport-Independent Relay**: In-process `LocalRelay` with an interface ready for future distributed message brokers (NATS, Redis Pub/Sub).

---

## Supported Editor Languages

The backend validates and normalizes the following editor languages as document metadata:
1. `c`
2. `cpp` (aliases: `c++`)
3. `java`
4. `python` (aliases: `py`)
5. `go` (aliases: `golang`)
6. `javascript` (aliases: `js`)
7. `typescript` (aliases: `ts`)
8. `csharp` (aliases: `c#`)
9. `rust` (aliases: `rs`)

---

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go                  # Service entrypoint & graceful shutdown
├── internal/
│   ├── collaboration/
│   │   ├── language.go              # Validated editor languages & normalizer
│   │   ├── document.go              # Document metadata representation
│   │   └── update_log.go            # Monotonic in-memory bounded CRDT update log
│   ├── config/
│   │   └── config.go                # Env var loader with validation & defaults
│   ├── http/
│   │   ├── handlers/
│   │   │   ├── health.go            # /healthz and /readyz probes
│   │   │   └── room.go              # Room creation & metadata lookup
│   │   ├── middleware/
│   │   │   ├── cors.go              # CORS configuration
│   │   │   ├── logger.go            # Structured latency logger (no code logging)
│   │   │   ├── ratelimit.go         # Per-IP token bucket rate limiting
│   │   │   └── recovery.go          # Sanitized panic recovery
│   │   └── routes/
│   │       └── routes.go            # Gin engine & route mapping
│   ├── lifecycle/
│   │   └── cleaner.go               # Background idle room & stale limiter pruning
│   ├── relay/
│   │   ├── relay.go                 # Transport-independent pub/sub interface
│   │   └── local.go                 # In-process race-safe channel relay
│   ├── room/
│   │   ├── id.go                    # Cryptographic room and participant ID generators
│   │   ├── participant.go           # Participant model with bounded outbound queue
│   │   ├── room.go                  # Room state, participant tracking, and snapshots
│   │   └── manager.go               # Concurrency-safe registry of active rooms
│   ├── security/
│   │   ├── origin.go                # Origin header validation
│   │   └── ratelimit.go             # Token bucket rate limiter implementation
│   └── websocket/
│       ├── protocol.go              # JSON envelopes & binary framing encoder/decoder
│       ├── connection.go            # Dual pump (read/write/relay) connection lifecycle
│       └── handler.go               # WebSocket upgrade handler for Gin
├── tests/
│   ├── integration_test.go          # Multi-client sync, replay, and race integration tests
│   └── conflict_and_disconnect_test.go # Same-line conflict & internet disconnect tests
├── docs/
│   ├── architecture.md              # System design and component diagrams
│   ├── conflict-resolution.md       # Same-line conflict & disconnection recovery guide
│   ├── websocket-protocol.md        # Protocol envelopes and binary frame spec
│   ├── api.md                       # HTTP REST API reference & examples
│   └── decisions.md                 # Architecture Decision Records (ADR)
├── .env.example                     # Sample configuration
├── .gitignore
├── Dockerfile                       # Multi-stage production container build
├── Makefile                         # Build and test commands
└── README.md
```

---

## Getting Started

### Prerequisites
- Go 1.22+ (tested on Go 1.27)
- `make` (optional)
- `docker` (optional)

### 1. Run Locally
```bash
# Clone and enter backend directory
cd backend

# Copy sample configuration
cp .env.example .env

# Run service
go run ./cmd/api
# Or via Makefile:
make run
```
The server will start on `http://localhost:8080`.

---

### 2. Run Tests with Race Detector
```bash
# Run all unit and integration tests with Go race detection
go test -race -v ./...
# Or via Makefile:
make test-race
```

---

### 3. Build Static Binary
```bash
go build -o bin/syncraft-api ./cmd/api
# Or via Makefile:
make build
```

---

### 4. Build and Run with Docker
```bash
# Build production Docker image
docker build -t syncraft-backend:latest .

# Run container
docker run -p 8080:8080 syncraft-backend:latest
```

---

## HTTP API Quick Reference

### 1. Health Probe
```bash
curl http://localhost:8080/healthz
```
Response:
```json
{
  "active_rooms": 1,
  "goroutines": 14,
  "memory_alloc_mb": 4,
  "participants": 0,
  "status": "ok",
  "uptime_seconds": 12
}
```

### 2. Create Anonymous Room
```bash
curl -X POST http://localhost:8080/api/v1/rooms \
     -H "Content-Type: application/json" \
     -d '{"language": "python"}'
```
Response `201 Created`:
```json
{
  "room_id": "V-EPwtZJ5giTsRPl9n0zug",
  "ws_url": "/api/v1/rooms/V-EPwtZJ5giTsRPl9n0zug/ws",
  "language": "python",
  "document_title": "main.py",
  "created_at": "2026-10-17T15:00:00Z",
  "expires_at": "2026-10-18T15:00:00Z",
  "max_participants": 25
}
```

### 3. Get Room Metadata
```bash
curl http://localhost:8080/api/v1/rooms/V-EPwtZJ5giTsRPl9n0zug
```
Response `200 OK`:
```json
{
  "room_id": "V-EPwtZJ5giTsRPl9n0zug",
  "language": "python",
  "document_title": "main.py",
  "participant_count": 0,
  "max_participants": 25,
  "created_at": "2026-10-17T15:00:00Z",
  "last_active_at": "2026-10-17T15:00:00Z",
  "update_count": 0
}
```

### 4. Connect via WebSocket
```
ws://localhost:8080/api/v1/rooms/V-EPwtZJ5giTsRPl9n0zug/ws?nickname=Alice
```

---

## Configuration Reference

| Environment Variable | Default | Description |
|---|---|---|
| `HTTP_HOST` | `0.0.0.0` | Host interface to bind |
| `HTTP_PORT` | `8080` | Port to listen on |
| `READ_TIMEOUT` | `15s` | HTTP server read timeout |
| `WRITE_TIMEOUT` | `15s` | HTTP server write timeout |
| `GRACEFUL_SHUTDOWN_TIMEOUT` | `10s` | Maximum duration to wait during shutdown |
| `ALLOWED_ORIGINS` | `*` | Comma-separated allowed CORS & WS origins |
| `MAX_ROOMS` | `1000` | Global limit of concurrent active rooms |
| `MAX_PARTICIPANTS_PER_ROOM` | `25` | Maximum connected clients per room |
| `MAX_WEBSOCKET_MESSAGE_SIZE`| `524288` (512KB) | Maximum single message size in bytes |
| `MAX_BUFFERED_UPDATES_PER_ROOM`| `2000` | Maximum updates in room replay log |
| `ROOM_IDLE_EXPIRATION` | `24h` | Inactivity duration before empty room is purged |
| `ROOM_CLEANUP_INTERVAL` | `10m` | Interval between background cleanup runs |
| `RATE_LIMIT_REQUESTS_PER_MINUTE`| `120` | Max HTTP requests per minute per IP |
| `RATE_LIMIT_BURST` | `30` | Maximum burst capacity for IP rate limiter |
| `WS_MESSAGE_RATE_LIMIT` | `100` | Max messages per second per WebSocket client |

---

## Known Limitations & Frontend Integration Notes

1. **In-Memory Volatility**: Because no database is used, all rooms and update histories are lost if the server process restarts. Persistent document storage is the responsibility of the client (using the browser File System Access API).
2. **Single-Node In-Process Relay**: `LocalRelay` operates in-process. In a future horizontally scaled deployment, swap `LocalRelay` with a Redis Pub/Sub or NATS implementation conforming to `relay.Relay`.
3. **CRDT Internal Convergence**: The backend treats CRDT updates as opaque binary byte streams and ensures ordered replay. Final conflict resolution is performed client-side by Yjs CRDT logic.
