# Syncraft Architecture Specification

## 1. System Overview

Syncraft is a low-latency, anonymous real-time collaborative workspace backend designed for synchronized code editing. It adheres to strict zero-persistence constraints:
- **No Database**: No PostgreSQL, MongoDB, Redis, SQLite, or disk-backed storage.
- **In-Memory Concurrency Safe**: All room states, participant registries, CRDT update buffers, and presence indicators reside strictly in volatile server RAM.
- **Anonymous Sessions**: No sign-up, user accounts, passwords, or OAuth. Access is controlled by unguessable room identifiers generated via cryptographically secure pseudo-random generators (`crypto/rand`).
- **CRDT-Agnostic Transport**: Document edits are treated as opaque binary byte streams compatible with Yjs / Monaco editor frontends.

---

## 2. Logical Components

```
                   +----------------------------------+
                   |          Reverse Proxy           |
                   |   (TLS Termination, Load Balancer)|
                   +-----------------+----------------+
                                     |
                                     v
                       +-------------+------------+
                       |       Gin HTTP Engine    |
                       |  - CORS / Safe Recovery  |
                       |  - IP Rate Limiting      |
                       +-------------+------------+
                                     |
               +---------------------+---------------------+
               |                                           |
               v                                           v
    +--------------------+                     +--------------------+
    |  REST Handlers     |                     | WebSocket Handler  |
    | - GET /healthz     |                     | - Origin Check     |
    | - GET /readyz      |                     | - Capacity Check   |
    | - POST /api/v1/rooms                     | - WS Upgrade       |
    | - GET  /api/v1/rooms/:id                 +---------+----------+
    +----------+---------+                               |
               |                                         v
               |                             +-----------+-----------+
               |                             |   Connection Session  |
               |                             | - Read / Write Pumps  |
               |                             | - Rate Limiter (Token)|
               |                             | - Bounded Send Queue  |
               |                             +-----------+-----------+
               |                                         |
               +--------------------+--------------------+
                                    |
                                    v
                     +--------------+--------------+
                     |         Room Manager        |
                     |  - RWMutex Registry         |
                     |  - Active Rooms Map         |
                     |  - Participant Management   |
                     |  - Idle Cleaner Hook        |
                     +--------------+--------------+
                                    |
               +--------------------+--------------------+
               |                                         |
               v                                         v
    +--------------------+                    +--------------------+
    |        Room        |                    |    Relay Layer     |
    | - UpdateLog Buffer |                    | (LocalRelay / NATS)|
    | - Document Meta    |                    | - Pub/Sub Per Room |
    | - Participant Refs |                    | - Non-blocking Send|
    +--------------------+                    +--------------------+
```

### Component Details

1. **HTTP API Server (Gin)**:
   - Exposes operational probes (`/healthz`, `/readyz`).
   - Serves room creation and metadata lookup endpoints under `/api/v1/rooms`.
   - Protects endpoints using global recovery middleware, structured latency logging (without logging code payloads), and per-IP token bucket rate limiting.

2. **WebSocket Server (`github.com/coder/websocket`)**:
   - Manages connection upgrades with origin verification.
   - Assigns ephemeral server-generated participant IDs (`p_<hex>`).
   - Isolates network operations with dual-pump goroutines (`readPump` and `writePump`) to eliminate concurrent socket writes.
   - Enforces per-connection message rate limits and payload size ceilings.

3. **Room Manager (`internal/room`)**:
   - Concurrency-safe registry of active rooms using fine-grained reader/writer mutexes (`sync.RWMutex`).
   - Enforces global limits on maximum concurrent rooms (`MAX_ROOMS`) and participants per room (`MAX_PARTICIPANTS_PER_ROOM`).
   - Never exposes internal mutable pointers outside the package; all queries return immutable snapshots.

4. **Collaboration & In-Memory Update Log (`internal/collaboration`)**:
   - Maintains a bounded, sequentially indexed update log (`UpdateLog`) for each room.
   - Assigns monotonic sequence numbers and timestamps to updates.
   - Supports catch-up / sync replay for newly joined clients (`GetSince(seq)`).
   - Rejects updates deterministically with `ErrUpdateLogCapacityExceeded` when capacity is hit, preventing silent data drops.

5. **Relay Layer (`internal/relay`)**:
   - Decoupled `Relay` interface for message forwarding.
   - Initial version provides `LocalRelay` operating entirely in-memory using non-blocking Go channels and per-subscriber mutexes.
   - Clean boundary enables zero-downtime transition to distributed pub/sub (Redis, NATS) in multi-node deployments without modifying WebSocket protocols.

6. **Lifecycle Cleaner (`internal/lifecycle`)**:
   - Periodic background worker executing at `ROOM_CLEANUP_INTERVAL`.
   - Detects rooms with 0 participants exceeding `ROOM_IDLE_EXPIRATION` and purges them from memory.
   - Prunes stale IP rate limiter buckets.

---

## 3. Concurrency and Synchronization Strategy

- **No Global Locks During Network I/O**: Network writes never occur while holding room manager or room state locks. Messages are placed on bounded participant channels or dispatched via relay outside locks.
- **Race-Safe Participant Queues**: Outbound buffers utilize `sync.RWMutex` to ensure `Close()` cannot race against non-blocking `Enqueue()` operations.
- **Slow Consumer Isolation**: If an individual client cannot consume outbound messages fast enough, their bounded queue saturates without blocking other participants in the room or the relay. Consecutive drops trigger graceful termination of the offending connection.
