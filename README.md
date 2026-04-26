<div align="center">

# Shortly

**Open-source URL shortener with real-time analytics**

[![Live Demo](https://img.shields.io/badge/Live%20Demo-shortly.app-0ea5e9?style=for-the-badge)](https://shortly.app)
[![License: MIT](https://img.shields.io/badge/License-MIT-22c55e?style=for-the-badge)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=for-the-badge&logo=react&logoColor=black)](https://react.dev)

</div>

---

Shortly turns any long URL into a clean, shareable link and gives you a live dashboard to understand who's clicking it — by device, country, browser, and day. No tracking pixels, no cookies, no account required to use a shortened link.

## Screenshots

> _Dashboard and analytics screenshots coming soon. Run it locally to see it in action._

| Shorten | Dashboard | Analytics |
|---------|-----------|-----------|
| ![shorten placeholder](https://placehold.co/380x220/0f172a/0ea5e9?text=Shorten+URL) | ![dashboard placeholder](https://placehold.co/380x220/0f172a/0ea5e9?text=Dashboard) | ![analytics placeholder](https://placehold.co/380x220/0f172a/0ea5e9?text=Analytics) |

## Features

- **Instant shortening** — generate a 6-character nanoid or bring your own custom slug
- **One-click redirect** — `301` permanent redirect with sub-millisecond response from Redis cache
- **Click analytics** — total clicks, trend chart (last 30 days), breakdown by device, browser, country
- **Async click recording** — redirects are never blocked waiting for analytics writes
- **Sliding-window rate limiting** — 10 shorten requests / minute per IP, enforced with a Redis Lua script
- **Expiring links** — optional `expires_at` timestamp; expired links return `410 Gone`
- **Soft delete** — deleted URLs are tombstoned, not erased, preserving click history
- **Geo-attribution** — country and city via [ip-api.com](http://ip-api.com) (no API key required)
- **Request tracing** — every response carries an `X-Request-ID` header for correlation
- **Graceful shutdown** — in-flight requests finish before the process exits

## Tech Stack

| Layer | Technology |
|---|---|
| **API** | ![Go](https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white) ![Gin](https://img.shields.io/badge/Gin-00ADD8?style=flat-square) |
| **Database** | ![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat-square&logo=postgresql&logoColor=white) ![GORM](https://img.shields.io/badge/GORM-ORM-4169E1?style=flat-square) |
| **Cache** | ![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat-square&logo=redis&logoColor=white) |
| **Frontend** | ![React](https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react&logoColor=black) ![Vite](https://img.shields.io/badge/Vite-646CFF?style=flat-square&logo=vite&logoColor=white) ![Tailwind](https://img.shields.io/badge/Tailwind-3-06B6D4?style=flat-square&logo=tailwindcss&logoColor=white) |
| **Charts** | ![Recharts](https://img.shields.io/badge/Recharts-2-22c55e?style=flat-square) |
| **Infra** | ![Docker](https://img.shields.io/badge/Docker-2496ED?style=flat-square&logo=docker&logoColor=white) |

## Getting Started

### Prerequisites

| Tool | Version | Install |
|---|---|---|
| Go | 1.22+ | [go.dev/dl](https://go.dev/dl) |
| Node.js | 20+ | [nodejs.org](https://nodejs.org) |
| Docker | any | [docs.docker.com](https://docs.docker.com/get-docker/) |
| Air _(hot reload)_ | latest | `go install github.com/air-verse/air@latest` |

### 1 — Clone and configure

```bash
git clone https://github.com/your-username/shortly.git
cd shortly
cp .env.example .env
```

Open `.env` and set at minimum:

```dotenv
JWT_SECRET=your-randomly-generated-secret   # required in production
```

Everything else works with the defaults for local development.

### 2 — Start backing services

```bash
make up        # starts Postgres 16 + Redis 7 via Docker Compose
```

Verify both are healthy:

```bash
docker compose ps
```

### 3 — Install dependencies

```bash
make setup     # go mod tidy && npm install
```

### 4 — Run the servers

Open two terminals:

```bash
# Terminal 1 — backend (live reload)
make backend   # http://localhost:8080

# Terminal 2 — frontend (HMR)
make frontend  # http://localhost:5173
```

The frontend proxies `/api` requests to the backend automatically — no extra config.

### All available commands

```text
make up                Start Postgres + Redis
make down              Stop Docker services
make logs              Tail Docker logs
make setup             Install all dependencies (run once)
make backend           Run backend with live reload
make frontend          Run frontend dev server
make build-backend     Compile binary → backend/bin/server
make build-frontend    Bundle → frontend/dist/
make test-backend      go test with race detector + coverage report
make lint-backend      golangci-lint
```

## API Reference

All API responses include an `X-Request-ID` header. Error responses follow a consistent shape:

```json
{
  "error": "human-readable message",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

---

### `POST /api/v1/shorten`

Shorten a URL. Rate-limited to **10 requests / minute per IP**.

**Request**

```json
{
  "url": "https://example.com/very/long/path?with=query&params=true",
  "custom_slug": "my-link",       // optional — omit for auto-generated code
  "expires_at": "2025-12-31T23:59:59Z"  // optional ISO 8601
}
```

**Response `201 Created`**

```json
{
  "id": "018f4e3a-7c2b-7000-8def-abcdef123456",
  "short_code": "my-link",
  "original_url": "https://example.com/very/long/path?with=query&params=true",
  "custom_slug": "my-link",
  "click_count": 0,
  "expires_at": "2025-12-31T23:59:59Z",
  "created_at": "2024-06-01T12:00:00Z",
  "updated_at": "2024-06-01T12:00:00Z"
}
```

| Status | Reason |
|--------|--------|
| `201` | Created |
| `400` | Missing or malformed body |
| `409` | Custom slug already taken |
| `422` | URL failed validation (no scheme, private IP, etc.) |
| `429` | Rate limit exceeded — check `Retry-After` header |

---

### `GET /:shortCode`

Redirect to the original URL. This is the hot path — served from Redis cache.

**Response `301 Moved Permanently`**

```
Location: https://example.com/very/long/path?with=query&params=true
```

Click analytics are recorded asynchronously and never block the redirect.

| Status | Reason |
|--------|--------|
| `301` | Redirect |
| `404` | Short code not found |
| `410` | Link has expired |

---

### `GET /api/v1/urls`

List all shortened URLs, newest first.

**Response `200 OK`**

```json
{
  "count": 2,
  "urls": [
    {
      "id": "018f4e3a-7c2b-7000-8def-abcdef123456",
      "short_code": "my-link",
      "original_url": "https://example.com/...",
      "click_count": 142,
      "created_at": "2024-06-01T12:00:00Z",
      "updated_at": "2024-06-01T12:00:00Z"
    }
  ]
}
```

---

### `DELETE /api/v1/urls/:id`

Soft-delete a URL. The short code stops resolving immediately (cache is invalidated). Historical click data is preserved.

**Response `204 No Content`**

| Status | Reason |
|--------|--------|
| `204` | Deleted |
| `400` | Invalid UUID |
| `404` | URL not found |

---

### `GET /api/v1/urls/:id/analytics`

Return aggregated click statistics for a URL.

**Response `200 OK`**

```json
{
  "total_clicks": 1842,
  "clicks_by_day": [
    { "date": "2024-06-01", "count": 214 },
    { "date": "2024-06-02", "count": 189 }
  ],
  "clicks_by_device": [
    { "label": "desktop", "count": 1102 },
    { "label": "mobile",  "count": 698  },
    { "label": "tablet",  "count": 42   }
  ],
  "clicks_by_country": [
    { "label": "US", "count": 910 },
    { "label": "GB", "count": 321 }
  ],
  "clicks_by_browser": [
    { "label": "Chrome",  "count": 987 },
    { "label": "Safari",  "count": 512 },
    { "label": "Firefox", "count": 343 }
  ]
}
```

`clicks_by_day` covers the **last 30 days**. `total_clicks` is the all-time count.

---

### `GET /health`

Readiness probe used by load balancers and container orchestrators.

```json
{ "status": "ok", "env": "production" }
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `APP_ENV` | `development` | `development` or `production` |
| `APP_PORT` | `8080` | HTTP listen port |
| `APP_BASE_URL` | `http://localhost:8080` | Public base URL (used in responses) |
| `DB_HOST` | `localhost` | Postgres host |
| `DB_PORT` | `5432` | Postgres port |
| `DB_USER` | `shortly` | Postgres user |
| `DB_PASSWORD` | — | Postgres password |
| `DB_NAME` | `shortly` | Postgres database name |
| `DB_SSLMODE` | `disable` | `disable` locally, `require` in production |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | — | Redis password |
| `JWT_SECRET` | — | **Required in production** |
| `SHORT_URL_LENGTH` | `7` | Auto-generated code length |
| `RATE_LIMIT_REQUESTS` | `100` | Requests per window |
| `RATE_LIMIT_WINDOW_SECONDS` | `60` | Rate limit window |
| `VITE_API_URL` | `http://localhost:8080/api/v1` | Frontend → backend URL |

## Deployment

### Backend — Railway

1. Push your repo to GitHub.
2. Create a new Railway project → **Deploy from GitHub repo** → select `backend/`.
3. Add a **Postgres** and **Redis** plugin from the Railway dashboard.
4. Set environment variables (copy from `.env.example`, fill real values).
5. Set the start command:
   ```
   ./server
   ```
   Railway automatically builds with the `backend/Dockerfile`.

### Frontend — Vercel

1. Import the repo on [vercel.com](https://vercel.com).
2. Set **Root Directory** to `frontend`.
3. Vercel auto-detects Vite — no build config needed.
4. Add one environment variable:
   ```
   VITE_API_URL=https://your-railway-backend.up.railway.app/api/v1
   ```
5. Deploy.

> **Tip:** Set `APP_BASE_URL` on Railway to your Railway backend URL and `CORS_ALLOWED_ORIGIN` (or `APP_ENV=production`) so the backend restricts CORS to your Vercel domain.

## Project Structure

```
shortly/
├── backend/
│   ├── cmd/server/          # Entrypoint — wires the dependency graph
│   ├── internal/
│   │   ├── config/          # Typed config loaded from env
│   │   ├── cache/           # Redis client
│   │   ├── database/        # GORM + Postgres (connection pool)
│   │   ├── models/          # GORM models + analytics response types
│   │   ├── repository/      # Data access layer (interfaces + Postgres impl)
│   │   ├── services/        # Business logic (URL shortening, analytics)
│   │   ├── handlers/        # Gin HTTP handlers
│   │   ├── middleware/       # Request ID, CORS, rate limiter
│   │   └── routes/          # Route registration
│   └── pkg/utils/           # nanoid, URL validator, UA parser
└── frontend/
    └── src/
        ├── components/      # Reusable UI components
        ├── pages/           # Route-level page components
        ├── hooks/           # Custom React hooks
        ├── services/        # API client (axios)
        ├── store/           # State management
        └── utils/           # Helpers
```

## Contributing

Contributions are welcome. Please open an issue before submitting a large pull request so we can discuss the approach first.

```bash
# Fork the repo, then:
git checkout -b feat/your-feature
# make your changes
make test-backend
git commit -m "feat: add your feature"
git push origin feat/your-feature
# open a pull request
```

**Bug reports** — please include your Go version, OS, and the output of `docker compose ps`.

**Feature requests** — open a GitHub issue with the label `enhancement` and describe the use case, not just the solution.

## License

[MIT](LICENSE) — do whatever you want, just keep the copyright notice.

---

<div align="center">
<sub>Built with Go, React, and too much coffee.</sub>
</div>
