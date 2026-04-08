# Tic-Tac-Toe Multiplayer (Go + React + Nakama)

A production-ready, server-authoritative multiplayer Tic-Tac-Toe game using Nakama for real-time networking and matchmaking.

## Stack
- **Backend:** Go (Nakama Server Runtime)
- **Database:** CockroachDB (via Docker)
- **Frontend:** React (Vite, TypeScript, Vanilla CSS)

## Prerequisites
- Docker & Docker Compose
- Node.js 18+ & npm
- Go 1.21+ (if developing backend locally)

## Running Locally

### 1. Start the Backend
The backend and frontend are containerized. A multi-stage Docker build compiles the Go plugin and mounts it into the Nakama server container automatically.

```bash
docker compose up --build
```
This will start:
- CockroachDB on port `26257`
- Nakama Server on port `7350` (API) and `7351` (Developer Console, default admin/password)
- Frontend web app on port `5173`

### 2. Optional: Start Frontend in Dev Mode
If you prefer local hot-reload instead of Docker-hosted frontend:

```bash
cd frontend
npm install
npm run dev
```
Open `http://localhost:5173` in your browser.

## Features implemented

- **Email Authentication:** Register or login using Nakama's native email identity.
- **Matchmaking Engine:** Players are added to a global Nakama matchmaker queue.
- **Server-Authoritative Gameplay:** A custom Go module spawns a dedicated Goroutine per match (`MatchLoop`), validating every move, preventing cheating, and detecting wins/draws.
- **Timeout Logic:** If a player doesn't move within 30 seconds, they forfeit.
- **Global Leaderboard:** Wins automatically increment a player's rank on the global leaderboard, displayed in the lobby.
- **Responsive UI:** Premium glassmorphism design leveraging CSS variables, built with pure React.

## Cloud Deployment Guide

### Deploying the Nakama Backend (AWS / GCP / DigitalOcean)
1. Provision a PostgreSQL / CockroachDB managed instance or run it on a VM.
2. Build the Docker image from `backend/Dockerfile` and push it to a private container registry (ECR, GCR, DockerHub).
3. Deploy the container on a managed service like AWS ECS, Google Cloud Run, or DigitalOcean App Platform.
4. Set the `DB_ADDRESS` environment variable to point to your managed database.
5. Expose ports `7350` for client traffic.

### Deploying the React Frontend
1. Change `SERVER_HOST` in `src/nakama.ts` to your production Nakama server IP or domain.
2. Build the production assets:
```bash
cd frontend
npm run build
```
3. Deploy the `dist/` folder to a static hosting provider like Vercel, Netlify, or AWS S3 + CloudFront.

## Local Teardown
To cleanly stop the local servers and wipe the database volumes:
```bash
docker-compose down -v
```

## Deployment Scripts (Windows PowerShell)

### Local Environment
Runs backend in Docker and starts the frontend dev server.

```powershell
.\deploy-local.ps1
```

Optional (skip `npm install`):

```powershell
.\deploy-local.ps1 -SkipFrontendInstall
```

### Production Environment
Builds backend Docker services and generates frontend production assets in `frontend/dist`.

```powershell
.\deploy-production.ps1
```

Optional host/port for frontend build-time Nakama target:

```powershell
.\deploy-production.ps1 -FrontendApiHost "your-api-host" -FrontendApiPort "7350"
```

## Deployment Scripts (Bash)

### Local Environment
Runs backend in Docker and starts the frontend dev server.

```bash
chmod +x deploy-local.sh
./deploy-local.sh
```

Optional (skip `npm install`):

```bash
./deploy-local.sh --skip-frontend-install
```

### Production Environment
Builds backend Docker services and generates frontend production assets in `frontend/dist`.

```bash
chmod +x deploy-production.sh
./deploy-production.sh
```

Optional host/port and skip install:

```bash
./deploy-production.sh --frontend-api-host=your-api-host --frontend-api-port=7350 --skip-frontend-install
```
