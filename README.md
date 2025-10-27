# GradeFlow Platform

GradeFlow is structured as a modular Go backend (Gin + GORM + PostgreSQL + MinIO) and a React + Chakra UI frontend. The backend is organised into domain models, DTOs, repositories, services, controllers, and middleware to keep business logic testable and extendable.

## Environment Variables

The API expects the following configuration (see `.env.example` for local defaults):

| Variable | Description |
| --- | --- |
| `PG_URL` | PostgreSQL connection string |
| `APP_URL` | Public origin used for links |
| `HTTP_ADDR` | Listen address (default `:8080`) |
| `API_BASE_PATH` | API prefix (default `/api`) |
| `JWT_SECRET` | HS256 secret for signing tokens |
| `ACCESS_TTL` / `REFRESH_TTL` | Token expiry durations |
| `CORS` | Comma separated list of allowed origins or `*` |
| `MINIO_ENDPOINT` | MinIO host:port |
| `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` | MinIO credentials |
| `MINIO_USE_SSL` | Enable TLS when contacting MinIO (`false` by default) |
| `MINIO_BUCKET_AVATARS` / `MINIO_BUCKET_REPORTS` | Buckets created on boot |

Logging options live under `LOG_*` (see `internal/config/config.go` for defaults).

## Running Locally

```bash
# Backend
go run ./cmd/api

# Frontend (in another terminal)
cd frontend
npm install
npm run dev
```

Swagger UI will be available at `http://localhost:8080/swagger/index.html` with API routes mounted under `/api`.

## Docker Compose

A convenience stack is provided:

```bash
docker compose up --build
```

This starts PostgreSQL, MinIO, the Go API (`http://localhost:8080`), the Vite dev server (`http://localhost:5173`), and Adminer for DB inspection (`http://localhost:8081`).

## Database Migrations

Initial DDL lives in `internal/migrations/0001_init.up.sql`. For development you can rely on the Postgres init hook from Docker Compose. For manual runs:

```bash
migrate -path internal/migrations -database "$PG_URL" up
```

Rollback support is available through the paired `.down.sql` file.

## Swagger Docs

To regenerate docs after editing controllers:

```bash
swag init -g cmd/api/main.go -o pkg/swagger
```

## Testing

Unit tests cover service authentication flows and controller integration. Run them via:

```bash
go test ./...
```

## Frontend

The React scaffold under `frontend/` provides Chakra UI layouts for:

- Login and token handling
- Dashboard with role-aware placeholder widgets
- Student grades view
- Teacher grade editor skeleton
- Dean panel for user/subject/group creation

Update `frontend/.env` with `VITE_API_BASE_URL` if you consume a different API origin.
