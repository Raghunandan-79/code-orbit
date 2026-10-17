# Syncraft HTTP REST API Documentation

Base Path: `/api/v1`

Syncraft exposes a minimal, anonymous HTTP API for health checks, room provisioning, and non-sensitive room discovery.

---

## 1. Health & Operations

### `GET /healthz`
Returns system health, process uptime, memory consumption, and active room counts.

**Response `200 OK`**:
```json
{
  "status": "ok",
  "uptime_seconds": 3600,
  "active_rooms": 12,
  "participants": 28,
  "memory_alloc_mb": 14,
  "goroutines": 92
}
```

---

### `GET /readyz`
Readiness probe for load balancers and container orchestrators. Returns 200 during normal operation and 503 during graceful shutdown.

**Response `200 OK`**:
```json
{
  "status": "ready"
}
```

**Response `503 Service Unavailable`**:
```json
{
  "status": "shutting_down"
}
```

---

## 2. Rooms API

### `POST /api/v1/rooms`
Provisions a new anonymous collaborative workspace. No authentication or login is required.

**Request Headers**:
- `Content-Type: application/json` (optional)

**Request Body** (optional):
```json
{
  "language": "python"
}
```
*If omitted or empty, `language` defaults to `"javascript"`.*

Supported Languages:
- `c`
- `cpp` (aliases: `c++`)
- `java`
- `python` (aliases: `py`)
- `go` (aliases: `golang`)
- `javascript` (aliases: `js`)
- `typescript` (aliases: `ts`)
- `csharp` (aliases: `c#`)
- `rust` (aliases: `rs`)

**Response `201 Created`**:
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

**Error Responses**:
- `400 Bad Request`: If an unsupported language identifier is supplied.
  ```json
  {
    "error": "unsupported language: \"cobol\" (supported: c, cpp, java, python, go, javascript, typescript, csharp, rust)"
  }
  ```
- `503 Service Unavailable`: If `MAX_ROOMS` capacity is reached.
  ```json
  {
    "error": "maximum room capacity reached, please try again later"
  }
  ```

**Example `curl` Command**:
```bash
curl -X POST http://localhost:8080/api/v1/rooms \
     -H "Content-Type: application/json" \
     -d '{"language": "python"}'
```

---

### `GET /api/v1/rooms/:roomID`
Retrieves non-sensitive metadata for an active room. Never exposes code content or participant IP addresses.

**Parameters**:
- `roomID` (path, required): 8-64 character room identifier.

**Response `200 OK`**:
```json
{
  "room_id": "V-EPwtZJ5giTsRPl9n0zug",
  "language": "python",
  "document_title": "main.py",
  "participant_count": 2,
  "max_participants": 25,
  "created_at": "2026-10-17T15:00:00Z",
  "last_active_at": "2026-10-17T15:04:12Z",
  "update_count": 47
}
```

**Error Responses**:
- `400 Bad Request`: Malformed room ID.
- `404 Not Found`: Room does not exist or has expired.
  ```json
  {
    "error": "room not found or expired"
  }
  ```

**Example `curl` Command**:
```bash
curl http://localhost:8080/api/v1/rooms/V-EPwtZJ5giTsRPl9n0zug
```

---

### `GET /api/v1/rooms/:roomID/ws`
Upgrades the HTTP connection to a WebSocket session.

**Query Parameters**:
- `nickname` (optional): Ephemeral display name (e.g. `?nickname=Alice`).
- `since_seq` (optional, uint64): Last sequence number received by client before disconnect. Server replays only updates strictly greater than this sequence.

**Headers**:
- `Origin`: Validated against `ALLOWED_ORIGINS` configuration.

**Upgrade Responses**:
- `101 Switching Protocols`: Successful WebSocket upgrade.
- `400 Bad Request`: Invalid room ID format.
- `403 Forbidden`: Disallowed Origin header OR room participant capacity reached.
- `404 Not Found`: Room does not exist or has expired.
