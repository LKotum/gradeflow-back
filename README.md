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
| `REDIS_ADDR` | Redis connection address (`redis:6379`) |
| `REDIS_PASSWORD` | Redis password (optional) |
| `REDIS_DB` | Redis database number (default `0`) |

Logging options live under `LOG_*` (see `internal/config/config.go` for defaults).

## Running Locally

```bash
# Backend (from gradeflow/)
go run ./cmd/api

# Frontend (from gradeflow-front/)
cd ../gradleflow-front
npm install
npm run dev
```

Swagger UI is exposed at `http://localhost:8080/swagger/index.html` with API routes under `/api`.

## Docker Compose

A convenience stack is provided:

```bash
docker compose up --build
```

This brings up PostgreSQL, MinIO, the Go API (`http://localhost:8080`), the Vite dev server (`http://localhost:5173`), and Adminer for DB inspection (`http://localhost:8081`). The compose file mounts `../gradleflow-front` into the frontend container.

When the API runs migrations for the first time it bootstraps an administrator account, logging the generated INS and password at startup. Admin authentication now uses the INS alongside the password.

## Database Migrations

Migrations are managed via gormigrate (`internal/migrations/migrations.go`). They run automatically on boot. To execute manually:

```bash
go test ./internal/migrations -run TestPlaceholder # triggers gormigrate wiring
```

or import the package and call `migrations.Run(ctx, db)` from your tooling.

## Swagger Docs

To regenerate docs after editing controllers:

```bash
swag init -g cmd/api/main.go -o pkg/swagger
```

## Testing

Unit tests cover authentication and controller plumbing. Run them with `go test ./...` when a Go toolchain is available.

## Frontend

The React application lives in `../gradleflow-front/` and ships Chakra UI based workspaces for each role:

- Admin dashboard to create dean staff
- Dean panel for managing subjects, groups, teachers, students and scheduling
- Teacher dashboard + gradebook editor consuming `/teacher/*` APIs
- Student dashboard and per-subject grade tracking

Update `../gradleflow-front/.env` with `VITE_API_BASE_URL` when pointing at a non-default API origin.
