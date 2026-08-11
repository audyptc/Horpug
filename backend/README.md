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

The service loads variables from `backend/.env` first and falls back to OS environment variables.

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
- `UPLOAD_DIR` (default: `./uploads`)
- `UPLOAD_BASE_URL` (default: `http://localhost:8080`)

Duration values must be valid Go duration strings such as `30s`, `15m`, `1h`, `24h`.

## Run Locally

From the `backend/` directory:

```bash
go mod download
go run ./cmd/api
```

On startup, the service will:

1. Load config
2. Connect to PostgreSQL
3. Run DB migrations
4. Start HTTP server

After changing or adding `@Router`/`@Summary` annotations on any handler, regenerate the Swagger
docs before committing (see [Swagger](#swagger) for the full command):

```powershell
./scripts/generate-swagger.ps1
```

## API and Health Endpoints

- `GET /health` - health check
- `GET /` - redirects to `/docs`
- `GET /docs` - Scalar API Reference UI
- `GET /swagger` - Swagger UI
- `GET /docs/swagger.json` - generated Swagger JSON
- `GET /uploads/*` - static file serving for uploaded documents (from `UPLOAD_DIR`)
- API base path: `/api/v1`

## Multi-Dormitory Request Scope

- Every dormitory-scoped endpoint now requires the `X-Dormitory-Id` header.
- The backend no longer auto-falls back to a user's first dormitory.
- Inactive dormitories are blocked for operational routes even if a user is assigned to them.
- Activity logs are now scoped by dormitory for operational endpoints.

Example:

```http
GET /api/v1/rooms
Authorization: Bearer <token>
X-Dormitory-Id: <dormitory-uuid>
```

## Permission Notes

- Users must have an active role to log in.
- Inactive roles cannot be assigned to users.
- Metadata endpoints such as `/api/v1/menus` and `/api/v1/permissions` now require role-management read permission.

## Swagger

Generate docs from `backend/` directory:

```powershell
./scripts/generate-swagger.ps1
```

The script keeps the `swag init` arguments in one place because this project stores Swagger
annotations and DTOs across many feature-level `domain` and `delivery` packages. A shorter direct
`swag init` command misses some types or duplicates routes with the current layout.


Then open:

- `http://localhost:8080/docs`

## Docker

Build backend image only:

```bash
docker build -t horpug-backend ./backend
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

The backend is organized by feature (vertical slices), not by technical layer. Each feature under
`internal/feature/` owns its own domain types, repository, usecase, and HTTP delivery code.

```text
backend/
├── cmd/
│   ├── api/                     # API entrypoint and docs routes
│   └── modelgen/                # generates SQL migrations from Go struct definitions
├── config/                      # configuration loading
├── docs/                        # generated swagger artifacts (docs.go, swagger.json, swagger.yaml)
├── internal/
│   ├── bootstrap/                # dependency wiring (constructs repos/usecases/handlers)
│   ├── database/                 # postgres connection + migrations
│   ├── delivery/http/            # router, error handler, shared response/pagination helpers
│   │   └── middleware/           # auth (JWT) and dormitory-scope middleware
│   ├── domain/                   # shared cross-feature types (tx, errors)
│   └── feature/                  # one folder per feature, e.g. room, tenant, bill, payment, ...
│       └── <feature>/
│           ├── domain/           # entities and request/response DTOs
│           ├── repository/       # persistence layer
│           ├── usecase/          # business logic
│           └── delivery/         # HTTP handlers + request validation (validate.go)
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```