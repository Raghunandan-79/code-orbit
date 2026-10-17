# Code Orbit (Syncraft)

An anonymous, real-time collaborative coding workspace. Users can instantly create a room, open a link, and collaboratively write and edit code in real-time without signing up, logging in, or installing anything.

---

## Architecture Overview

```
                      +-----------------------------+
                      |         Code Orbit          |
                      |   Next.js 16 + React 19     |
                      |    Tailwind CSS + Monaco    |
                      +--------------+--------------+
                                     |
               +---------------------+---------------------+
               | HTTP REST                                 | WebSockets (WSS)
               | (POST /api/v1/rooms)                      | (Opaque CRDT Frames &
               |                                           |  JSON Control Envelopes)
               v                                           v
    +---------------------------------------------------------------+
    |                       Syncraft Backend                        |
    |                 Go 1.24 + Gin + coder/websocket               |
    |        In-Memory RWMutex Registry, Zero Database, Zero Disk   |
    +---------------------------------------------------------------+
```

---

## Directory Structure

- **[`frontend/`](file:///home/raghu/Documents/projects/code-orbit/frontend)**: Next.js 16 application with Monaco Editor, live cursor presence, language selector, and room share modal.
- **[`backend/`](file:///home/raghu/Documents/projects/code-orbit/backend)**: Go + Gin collaboration backend managing in-memory active rooms, monotonic update logs, and WebSockets.
- **[`docs/deployment.md`](file:///home/raghu/Documents/projects/code-orbit/docs/deployment.md)**: **100% Free Production Deployment Guide** (Vercel + Render / Fly.io).
- **[`docs/conflict-resolution.md`](file:///home/raghu/Documents/projects/code-orbit/backend/docs/conflict-resolution.md)**: CRDT conflict tie-breaking and offline reconnection architecture.

---

## Quickstart (Local Development)

### Option 1: Docker Compose (Frontend + Backend in One Command)
```bash
docker compose up --build
```
- Frontend: `http://localhost:3000`
- Backend API: `http://localhost:8080`

---

### Option 2: Running Separately

#### 1. Start the Backend
```bash
cd backend
make run
```
*Listens on `http://localhost:8080`.*

#### 2. Start the Frontend
```bash
cd frontend
bun run dev
# Or: npm run dev
```
*Listens on `http://localhost:3000`.*

---

## Deployment (100% Free)

See the full step-by-step guide in [docs/deployment.md](file:///home/raghu/Documents/projects/code-orbit/docs/deployment.md):

1. **Frontend**: Deploy `frontend/` on **Vercel** (Hobby Plan - Free).
2. **Backend**: Deploy `backend/` on **Render** (Free Web Service) or **Fly.io** (Free Allowance).
3. Connect them via `NEXT_PUBLIC_API_URL` and `ALLOWED_ORIGINS`.
