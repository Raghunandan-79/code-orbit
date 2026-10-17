# Syncraft Architecture Decision Records (ADR)

## ADR 1: Zero-Persistence, In-Memory Only Architecture

### Context
Syncraft is designed for transient, anonymous, real-time collaborative coding sessions without account creation or signups. The requirements explicitly forbid databases (PostgreSQL, MongoDB, SQLite, Redis) and server-side filesystem storage of user code.

### Decision
All state (room registries, participant sessions, transient document metadata, and CRDT update logs) is stored in server volatile memory (`sync.RWMutex`-protected data structures). 

### Consequences
- **Ultra-low latency**: Zero disk I/O or network hops for persistence.
- **Process restart volatility**: Server restarts or crashes intentionally purge all active rooms and update histories.
- **Client Responsibility**: Persistent document saving belongs to the client-side frontend using the browser File System Access API.

---

## ADR 2: WebSocket Transport Library Selection

### Context
We evaluated WebSocket implementations for Go. `gorilla/websocket` is largely in maintenance mode/archived, whereas `github.com/coder/websocket` (formerly `nhooyr.io/websocket`) is actively maintained, idiomatic, and features first-class `context.Context` support.

### Decision
Use `github.com/coder/websocket`.

### Consequences
- Native `context.Context` integration across all `Read`, `Write`, and `Ping` operations.
- Clean separation of concurrent reads and writes. Concurrent writes to a single connection are prohibited by design in `coder/websocket`, which we enforce using dedicated per-connection `writePump` goroutines.

---

## ADR 3: Dual-Mode Protocol (JSON Envelopes + Binary Frames)

### Context
Yjs CRDT updates are opaque binary blobs. Encoding them in JSON base64 strings introduces ~33% bandwidth overhead and CPU serialization costs. However, control messages (handshakes, room metadata, error codes, and presence) benefit from human-readable, self-describing JSON schemas.

### Decision
Use a hybrid protocol:
- **Text WebSocket Frames**: Deliver versioned JSON envelopes (`version: 1`, `type`, `payload`).
- **Binary WebSocket Frames**: Deliver binary frames prefixed with a compact 4-byte header:
  `[Version (1B)][Subtype (1B)][DocIDLen (1B)][Flags (1B)][DocID (NB)][Opaque Payload]`.
  Also supports automatic fallback for raw Yjs frames.

### Consequences
- Eliminates base64 overhead for document keystrokes and diffs.
- Clear separation between control flow and collaborative document data.

---

## ADR 4: Transport-Independent Relay Interface

### Context
The initial implementation must run in-process on a single server without introducing external messaging infrastructure (Redis/NATS). However, future multi-instance horizontal scaling should not require rewriting connection handlers or protocol layers.

### Decision
Define a `Relay` interface:
```go
type Relay interface {
    Publish(ctx context.Context, msg *BroadcastMessage) error
    Subscribe(ctx context.Context, roomID string, participantID string) (Subscription, error)
    Close() error
}
```
Implement `LocalRelay` for the current in-process deployment.

### Boundaries and Limitations
- `LocalRelay` routes broadcasts only within the local process.
- It is **not** a distributed relay. Messages do not cross process boundaries.
- When scaling horizontally in the future, a `NatsRelay` or `RedisRelay` can be dropped in without changing the WebSocket or Room code.

---

## ADR 5: Deterministic Resource Limits and Slow Consumer Backpressure

### Context
Malicious or slow clients could cause unbounded memory growth by either failing to read messages from the WebSocket socket, or by flooding a room with millions of updates.

### Decision
1. **Bounded Queues**: Each participant has a bounded outbound channel (`capacity: 256`). If the buffer overflows, messages are dropped, and exceeding 20 consecutive drops flags the participant as an unrecoverable slow consumer, terminating the connection safely.
2. **Bounded Update Log**: Each room holds at most `MAX_BUFFERED_UPDATES_PER_ROOM` updates. If reached, appends fail with `ErrUpdateLogCapacityExceeded`, and the server returns an `UPDATE_LIMIT_EXCEEDED` error to the client rather than silently discarding keystrokes.
3. **No Locks on Network I/O**: Broadcast dispatch and socket writes never hold room or manager locks.
