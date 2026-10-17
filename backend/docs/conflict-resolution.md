# Conflict Resolution & Network Disconnection Guide

This document explains precisely how **Syncraft** resolves editing conflicts (such as two participants concurrently modifying the same line of code) and how **network disconnections and offline editing** are handled.

---

## 1. Conflict Resolution: "Who Wins When Editing the Same Line?"

In traditional central architectures, simultaneous writes often result in **Last-Write-Wins (LWW)** or painful **merge conflict dialogs** (like Git conflict markers). 

Syncraft utilizes a **Conflict-free Replicated Data Type (CRDT)** architecture (compatible with Yjs). Under this model:

> **Neither user loses their work. There is no destructive overwrite.**

### 1.1 The Mathematical Model (YATA CRDT Algorithm)

Every character or token inserted into the editor is assigned a globally unique, immutable identifier:
$$\text{ID} = (\text{ClientID}, \text{Clock})$$

- `ClientID`: A randomly generated unique integer assigned to the client editor instance.
- `Clock`: A monotonically incrementing counter tracking edits made by that client.

Each piece of content also remembers its immediate left neighbor (`originLeft`) and right neighbor (`originRight`) at the time of creation.

```
       Original State: [let counter = 0;]
                         ^
                         | (Both users edit at column 12)
        +----------------+----------------+
        |                                 |
        v                                 v
User A (Alice, ClientID=10):      User B (Bob, ClientID=20):
Insert "/* A */"                  Insert "/* B */"
Left: "counter = "                Left: "counter = "
Right: "0;"                       Right: "0;"
```

### 1.2 Deterministic Tie-Breaking

When Alice and Bob type at the exact same cursor position simultaneously:
1. Both operations reference the exact same `originLeft` and `originRight`.
2. The CRDT algorithm breaks the positional tie deterministically using the client identifiers (`ClientID`):
   - If $\text{ClientID}_A < \text{ClientID}_B$, Alice's insertion is placed first, and Bob's insertion is placed immediately adjacent.
   - If $\text{ClientID}_B < \text{ClientID}_A$, Bob's insertion is placed first.
3. **Result**: Both Alice and Bob's editors display:
   ```typescript
   let counter = /* A *//* B */0;
   ```
   (or `/* B *//* A */0;` depending on client ID order).

Because the tie-breaking rule is **deterministic, commutative, and idempotent**, every peer calculates the exact same visual text regardless of the order in which network packets arrived.

### 1.3 Insert vs. Delete Conflicts

What if Alice deletes line 1 while Bob edits line 1?
- In Yjs CRDT, deletions do not physically delete nodes immediately; they mark nodes as **tombstones** (`deleted = true`).
- If Bob inserted text next to a character that Alice concurrently deleted, Bob's inserted text remains visible, while Alice's deleted text disappears.
- No null pointer exceptions, index shifting bugs, or data corruption occur.

---

## 2. The Backend's Role in Conflict Management

The Go backend acts as the **authoritative monotonic sequencer and message relay**:

1. **Atomic Sequence Assignment**:
   When client updates arrive, the backend serializes them into the room's in-memory `UpdateLog` under a write mutex:
   ```go
   entry, err := room.UpdateLog().Append(docID, senderID, payload, isBinary)
   ```
   Each entry is assigned a strictly incrementing sequence number (`Seq: 1, 2, 3...`).
2. **Total Ordering Guarantee**:
   Even if two clients send updates within the same microsecond, the Go runtime orders them deterministically into the log.
3. **No Update Dropping**:
   The backend never silently discards valid updates. Both edits are committed and relayed to all other room participants.

---

## 3. Handling Network Disconnections & Offline Edits

```
+-----------+                      +------------------+                      +-----------+
|  Alice    |                      | Syncraft Backend |                      |   Bob     |
+-----+-----+                      +--------+---------+                      +-----+-----+
      |                                     |                                      |
      | 1. Connected (Editing normally)     |                                      |
      +------------------------------------>|                                      |
      |                                     |                                      |
      |                                     |           2. INTERNET DROPS          |
      |                                     |           (Bob's WiFi disconnects)   |
      |                                     |                  X                   |
      |                                     |                  |                   |
      |                                     | 3. Heartbeat Timeout (15s + 5s)      |
      |                                     |    Detects dead peer                 |
      | 4. Broadcast participant_left       |<-----------------+                   |
      |<------------------------------------+                                      |
      |                                     |                                      |
      | 5. Alice continues editing:         |                                      |
      |    - Seq 2: "x = 10"                |                                      |
      |    - Seq 3: "y = 20"                |                                      |
      +------------------------------------>|           6. Bob continues offline:  |
      |                                     |              (Edits stored in local  |
      |                                     |               Yjs update queue)      |
      |                                     |                                      |
      |                                     |           7. INTERNET RECOVERS       |
      |                                     |              Bob reconnects with     |
      |                                     |              ?since_seq=1            |
      |                                     |<-------------------------------------+
      |                                     |                                      |
      |                                     | 8. Server replays Seq 2 & 3          |
      |                                     +------------------------------------->|
      |                                     |    (Bob's editor applies missed edits|
      |                                     |                                      |
      |                                     | 9. Bob streams offline edits         |
      | 10. Server relays Bob's offline edit|<-------------------------------------+
      |     as Seq 4                        |                                      |
      |<------------------------------------+                                      |
      |                                     |                                      |
      v                                     v                                      v
  [Document States Converge: Alice and Bob have identical code with 0 lost edits]
```

### 3.1 Dead Connection Detection (TCP Half-Open Sockets)

When a client loses internet connectivity (e.g. WiFi drops, laptop closed, or mobile network handoff), the client may not send a clean TCP `FIN` packet.
- The Syncraft backend executes proactive ping probes:
  - `WS_PING_INTERVAL` (default `15s`)
  - `WS_PING_TIMEOUT` (default `5s`)
- If a connection fails to respond within the ping deadline, the server terminates the socket, cleans up the participant queue, and broadcasts `participant_left` to remaining room members.
- The room state and update log remain safely preserved in server memory.

### 3.2 Offline Editing

While disconnected, the client's local editor does not freeze. The frontend's local Yjs instance continues accepting keystrokes and buffering the generated binary update packets locally in memory.

### 3.3 Reconnection with `since_seq`

When connectivity returns, the client reconnects to the WebSocket endpoint:
```
GET /api/v1/rooms/<roomID>/ws?nickname=Bob&since_seq=1
```

1. **Delta Synchronization**:
   The server inspects `since_seq=1` and replays **only** the updates created while Bob was offline (`Seq > 1`).
2. **Local Integration**:
   Bob's local Yjs engine applies the missed server updates.
3. **Offline Edit Flushing**:
   Bob's client flushes its buffered offline updates to the server as standard binary frames.
4. **Broadcast & Convergence**:
   The server commits Bob's offline updates as new sequential entries (`Seq 4`) and broadcasts them to Alice.
5. Both Alice and Bob's editors run the CRDT merge rules and arrive at the exact same document state.
