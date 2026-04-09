# Tic-Tac-Toe Multiplayer (Go + React + Nakama)

## Deployment Links
- **Accessible Game URL:** `[Insert deployed game URL or IP here, e.g., http://your-server-ip]`
- **Deployed Nakama Server Endpoint:** `[Insert Nakama API URL or IP here, e.g., http://your-server-ip:7350]`

---

## Architecture and Design Decisions
- **Architecture Overview**: The application uses a Client-Server architecture. The backend acts as an authoritative source of truth to prevent cheating, validating all game moves securely on the server.
- **Backend (Nakama + Go)**: Nakama handles the heavy lifting of user authentication, session management, and matchmaking. A Go plugin overrides Nakama's default match handler to implement custom Tic-Tac-Toe validation, maintaining state (board, turns, and timers) completely in-memory.
- **Database (CockroachDB)**: CockroachDB is the native database choice for Nakama, offering a highly resilient distributed SQL store for user profiles, leaderboard scores, and authentication records.
- **Frontend (React + TypeScript)**: The UI prioritizes a sleek responsive glassmorphism design. To keep it light and fast, it uses Vite as the build tool, native React state management, and direct integration with `@heroiclabs/nakama-js`.

## API & Server Configuration Details
- **Nakama Backend**: 
  - API Port: `7350` (used by the web/mobile client to communicate)
  - Admin Console Port: `7351`
  - Admin Credentials (Default): `admin` / `password`
- **CockroachDB**: Exposes port `26257` for internal DB traffic. 
- **Environment Variables**:
  - `VITE_NAKAMA_HOST`: Informs the React frontend where the Nakama API is located. If left empty, the frontend automatically falls back to `window.location.hostname`. This design provides seamless functionality when accessed remotely without requiring you to constantly update IP addresses.
  - `DB_ADDRESS`: Connection string used by Nakama to talk to CockroachDB. Default: `root@cockroachdb:26257`

## Setup and Installation Instructions

### Prerequisites
- Docker & Docker Compose
- Node.js 18+ & npm (Optional: only if running frontend dev server locally apart from Docker)

### Running Locally with Docker
The entire project (Database, API, and React Web App via Nginx) is containerized and orchestrated via Docker Compose.

1. Clone the repository to your machine.
2. Open a terminal in the project root containing `docker-compose.yml`.
3. Run the following command:
   ```bash
   docker compose up --build
   ```
   *(Add `-d` to run in detached mode).*
4. Access the frontend app locally at `http://localhost` (or `http://127.0.0.1`).

## How to Test the Multiplayer Functionality
To accurately test the matchmaking and real-time multiplayer flow, follow these steps:
1. Open two different browsers (e.g., Chrome and Firefox) or two incognito/private windows.
2. Navigate to the game URL (`http://localhost` or your server IP) in both windows.
3. In Window A, register a new account (e.g., `player1@example.com` / `password`).
4. In Window B, register a second account (e.g., `player2@example.com` / `password`).
5. Click **"Find Match"** in both windows simultaneously.
6. The global matchmaking engine will pair both accounts instantly. You will be routed into a live game session.
7. Verify that clicking a cell on one screen instantly reflects on the other. 
8. Test timeout parameters by waiting 30 seconds without playing; the server will naturally forfeit the idle player. Complete a full game to observe board validation and leaderboard reporting.

## Deployment Process Documentation
The simplest way to deploy the entire production stack is on a cloud Linux Virtual Machine (AWS EC2, DigitalOcean Droplet, GCP Compute Engine). 

### Single-Node VM Deployment
1. **Provision a VM** with Docker and `docker-compose` installed. 
2. **Configure Firewalls**: Ensure HTTP (Port 80) and the Nakama API (Port 7350) are exposed to the public internet. If using Azure/AWS, update the underlying Security Groups/Inbound rules.
3. Clone the repository onto your server.
4. Run the production docker-compose directly:
   ```bash
   docker compose up -d --build
   ```
5. **Auto-Resolution**: Since `docker-compose.yml` binds to port 80 and the application `nakama.ts` uses `window.location.hostname`, players visiting `http://<YOUR_VM_PUBLIC_IP>` will seamlessly have their WebSocket/API requests routed back to `<YOUR_VM_PUBLIC_IP>:7350`.

*Note: For horizontal scaling in an Enterprise environment, the Nakama image should be deployed using Kubernetes (EKS/GKE), and CockroachDB would be shifted to a managed provider (CockroachDB Serverless).*

### Local Teardown
To gracefully stop the local servers and completely wipe the database volumes:
```bash
docker compose down -v
```
