# Server

Backend service for Horpug, built with Go and Fiber.

## Architecture

```text
[Client] -> Fiber Router / Handlers -> Usecase -> Repository -> PostgreSQL
```

## Requirements

- Go 1.26+
- PostgreSQL 14+
- (Optional) `swag` CLI for generating Swagger files

Install `swag` (optional):

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

## Environment Variables

The service loads variables from `server/.env` first and falls back to OS environment variables.

Required:

- `APP_SECRETKEY`

Common variables:

- `APP_PORT` (default: `8080`)
- `DB_HOST` (default: `localhost`)
- `DB_PORT` (default: `5432`)
- `DB_USERNAME` (default: `postgres`)
- `DB_PASSWORD` (default: `password`)
- `DB_NAME` (default: `goapi_db`)
- `DB_SSLMODE` (default: `disable`)
- `CORS_ORIGINS` comma-separated list (default: `http://localhost,http://localhost:5173,http://localhost:3000`)
- `ACCESS_TOKEN_TTL` (default: `15m`)
- `REFRESH_TOKEN_TTL` (default: `168h`)

Duration values must be valid Go duration strings such as `30s`, `15m`, `1h`, `24h`.

## Run Locally

From the `server/` directory:

```bash
go mod download
go run ./cmd/api
```

On startup, the service will:

1. Load config
2. Connect to PostgreSQL
3. Run DB migrations
4. Start HTTP server

## API and Health Endpoints

- `GET /health` - health check
- `GET /` - redirects to `/docs`
- `GET /docs` - Scalar API Reference UI
- `GET /docs/swagger.json` - generated Swagger JSON
- API base path: `/api/v1`

## Swagger

Generate docs from `server/` directory:

```bash
swag init -g main.go -d cmd/api,internal/delivery/http,internal/domain -o docs
```

Then open:

- `http://localhost:8080/docs`

## Docker

Build backend image only:

```bash
docker build -t horpug-backend ./server
```

Run full stack from repository root:

```bash
docker compose up -d --build
```

Main entry points when running with compose:

- App gateway: `http://localhost`
- Backend health (through nginx): `http://localhost/health`
- Docs (through nginx): `http://localhost/docs`

## Project Structure

```text
server/
├── cmd/
│   ├── api/                # API entrypoint and docs routes
│   └── modelgen/           # helper/generator entrypoint
├── config/                 # configuration loading
├── docs/                   # generated swagger artifacts
├── internal/
│   ├── bootstrap/          # dependency wiring
│   ├── database/           # postgres + migrations
│   ├── delivery/http/      # handlers, middleware, routes
│   ├── domain/             # entities and interfaces
│   ├── repository/         # persistence layer
│   ├── usecase/            # business logic
│   └── validator/          # validation utilities
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```