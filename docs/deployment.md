# Complete Free Deployment Guide: Code Orbit & Syncraft

This document analyzes the architecture of both the **Frontend** (Next.js 16 + React 19 + Monaco) and **Backend** (Go + Gin + WebSockets) and provides step-by-step instructions to deploy the entire system **100% free of charge**.

---

## 1. Architectural Analysis & Deployment Constraints

| Component | Stack | Characteristics | Hosting Constraint |
|---|---|---|---|
| **Frontend** | Next.js 16 (App Router), React 19, Tailwind CSS 4, Monaco Editor | SSR / Static web application | **Serverless Edge Platforms** (Vercel, Cloudflare Pages, Netlify) — 100% Free |
| **Backend** | Go 1.24, Gin, `coder/websocket`, In-Memory State | Long-lived WebSocket sessions, zero database, stateful RAM | **Requires Continuous Container Compute** (Cannot use traditional serverless functions like AWS Lambda because they cannot sustain stateful WebSocket rooms in RAM) |

---

## 2. Free Tier Hosting Comparison Matrix

| Provider | Target | Free Allowance | WebSocket Support | Sleep Behavior | Best For |
|---|---|---|---|---|---|
| **Vercel** | Frontend | Unlimited bandwidth (Hobby Tier), global Edge CDN | N/A (Frontend only) | Never sleeps | **Recommended Frontend** |
| **Render** | Backend | 750 free instance hours/month, 512MB RAM, free SSL | Full native WebSocket support | Sleeps after 15 min idle; auto-wakes on HTTP request | **Recommended Backend (Easiest)** |
| **Fly.io** | Backend | Free allowance: up to 3 shared-cpu-1x VMs (256MB RAM), global Anycast | Native low-latency WebSockets | Can run continuously without sleeping | **Recommended Backend (Fastest)** |
| **Koyeb** | Backend | Free Nano instance (512MB RAM), free SSL, Docker | Native WebSockets | Always-on continuous running | **Alternative Free Backend** |
| **Oracle Cloud** | Both | **Always Free Tier**: 4 ARM OCPUs, 24GB RAM, 200GB disk | Full native Docker Compose | Always-on, never sleeps | **Best Self-Hosted (Single VM)** |

---

## 3. Recommended Free Deployment: Vercel (Frontend) + Render (Backend)

This is the most reliable, zero-maintenance setup: Vercel provides world-class Next.js CDN hosting, while Render runs the Go backend container.

```
       [User Browser]
         |         |
   HTTPS |         | WSS (WebSockets) + REST
         v         v
    +---------+  +-------------------------------+
    | Vercel  |  | Render (Free Web Service)     |
    | (Next)  |  | (Go Container: syncraft-api) |
    +---------+  +-------------------------------+
```

### Step 1: Push Project to GitHub

Ensure your repository is pushed to GitHub:
```bash
git add .
git commit -m "feat: complete backend and frontend"
git push origin main
```

---

### Step 2: Deploy Backend to Render (Free)

1. Sign up / Log in to [Render.com](https://render.com) (free account, no credit card required).
2. Click **New +** > **Web Service**.
3. Connect your GitHub repository.
4. Configure the service settings:
   - **Name**: `syncraft-backend` (or your choice)
   - **Region**: Choose the region closest to you (e.g., `Oregon (US West)` or `Frankfurt (EU)`)
   - **Root Directory**: `backend`
   - **Runtime**: **Docker**
   - **Instance Type**: **Free** ($0/month)
5. Under **Advanced** > **Add Environment Variables**:
   - `ALLOWED_ORIGINS` = `*` *(We will restrict this to your Vercel URL in Step 4)*
   - `PORT` = `10000` *(Render provides this automatically, but you can specify 10000 or 8080)*
   - `MAX_ROOMS` = `500`
   - `MAX_PARTICIPANTS_PER_ROOM` = `25`
   - `ROOM_IDLE_EXPIRATION` = `24h`
6. Click **Create Web Service**.
7. Wait 1–2 minutes for the Docker container to build.
8. Once deployed, Render will display your backend URL (e.g.):
   ```
   https://syncraft-backend.onrender.com
   ```
9. Verify the backend health:
   ```bash
   curl https://syncraft-backend.onrender.com/healthz
   # Expected response: {"status":"ok","uptime_seconds":...}
   ```

> **Note on Render Free Tier Sleep Behavior**:
> Render free instances sleep after 15 minutes of inactivity. When a user creates a new room via the landing page (`POST /api/v1/rooms`), the HTTP request automatically wakes up the backend in ~20–30 seconds.

---

### Step 3: Deploy Frontend to Vercel (Free)

1. Sign up / Log in to [Vercel.com](https://vercel.com) with GitHub.
2. Click **Add New...** > **Project**.
3. Import your GitHub repository.
4. In the configuration screen:
   - **Framework Preset**: `Next.js`
   - **Root Directory**: Click **Edit** and select **`frontend`**
5. Expand **Environment Variables** and add:
   - `NEXT_PUBLIC_API_URL` = `https://syncraft-backend.onrender.com` *(Replace with your Render URL from Step 2, without trailing slash)*
   - `NEXT_PUBLIC_APP_URL` = `https://your-app-name.vercel.app` *(Your Vercel domain)*
6. Click **Deploy**.
7. In ~60 seconds, Vercel will deploy your Next.js application and give you a live production URL (e.g. `https://code-orbit.vercel.app`).

---

### Step 4: Lock Down CORS & Origins (Security Best Practice)

Once you have your live Vercel URL:
1. Go to your Render Dashboard > `syncraft-backend` > **Environment**.
2. Update `ALLOWED_ORIGINS`:
   ```env
   ALLOWED_ORIGINS=https://code-orbit.vercel.app
   ```
3. Render will redeploy automatically in ~30 seconds with origin locking enabled.

---

## 4. Alternative Option: Fly.io (Zero Sleep / Lowest Latency)

If you want a free backend that **does not sleep**, Fly.io provides a free allowance of up to 3 shared micro VMs.

1. Install Fly CLI:
   ```bash
   curl -L https://fly.io/install.sh | sh
   ```
2. Log in:
   ```bash
   fly auth login
   ```
3. Navigate to the backend directory:
   ```bash
   cd backend
   ```
4. Deploy using the included `fly.toml`:
   ```bash
   fly launch --no-deploy
   fly deploy
   ```
5. Fly will provision an Anycast IP with automatic SSL:
   ```
   https://syncraft-backend.fly.dev
   ```
6. Use `https://syncraft-backend.fly.dev` as `NEXT_PUBLIC_API_URL` in Vercel.

---

## 5. Alternative Option: Self-Hosting on Free Cloud VPS (Oracle Cloud Always Free)

Oracle Cloud provides an **Always Free Tier** with 4 ARM Ampere OCPUs and 24GB of RAM (significantly more powerful than any shared serverless tier).

1. Create a free Ubuntu compute instance on Oracle Cloud.
2. SSH into your server:
   ```bash
   ssh ubuntu@<YOUR_SERVER_IP>
   ```
3. Install Docker & Docker Compose:
   ```bash
   sudo apt update && sudo apt install -y docker.io docker-compose-v2
   sudo usermod -aG docker $USER
   ```
4. Clone your repository:
   ```bash
   git clone https://github.com/<your-username>/code-orbit.git
   cd code-orbit
   ```
5. Run both frontend and backend with a single command:
   ```bash
   docker compose up -d --build
   ```
6. Open ports `80`, `443`, `3000`, `8080` in the Oracle Cloud VCN Security List.
7. Both services will run continuously on your free virtual machine.

---

## 6. Pre-Flight Verification Checklist

Before sharing your public URL, verify these steps:

1. **Test REST API**:
   ```bash
   curl -X POST https://<YOUR_BACKEND_URL>/api/v1/rooms \
        -H "Content-Type: application/json" \
        -d '{"language": "python"}'
   ```
   Should return `201 Created` with a new `room_id`.

2. **Test Room Page**:
   Open `https://<YOUR_FRONTEND_URL>/room/<room_id>` in two separate browser windows (or incognito).
   - Verify connection indicator turns green (`Connected`).
   - Type in window 1; check that code and cursor positions update in window 2 instantly.

3. **Test Network Drop**:
   - Turn off Wi-Fi in window 1.
   - Type a few lines offline.
   - Re-enable Wi-Fi.
   - Verify the client reconnects and all offline edits merge cleanly without data loss.
