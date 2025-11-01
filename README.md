# GradeFlow

Управление учебным процессом: группы, предметы, расписание, оценки и роли (админ, декан, преподаватель, студент). Бэкенд — Go + Gin, БД — PostgreSQL, кэш — Redis, хранилище файлов (аватары/отчёты) — MinIO. Готовая Swagger-документация. 

## Архитектура

```
cmd/api/main.go
internal/
  app/               # сборка приложения, DI, HTTP сервер, CORS, Swagger
  config/            # загрузка конфигурации/ENV
  controllers/       # Gin-роуты (auth, admin, dean, teacher, student, profile)
  domain/            # модели, DTO запросов/ответов
  middleware/        # JWT
  migrations/        # gormigrate миграции
  repository/        # интерфейсы репозиториев + реализация GORM
  service/           # бизнес-логика
pkg/
  cache/             # Redis-кэш
  logger/            # логгер
  swagger/           # сгенерированные Swagger-файлы
```

## Возможности

* Роли: admin / dean / teacher / student (JWT Bearer). 
* Группы, предметы, занятия, оценки; пагинация и фильтры. 
* Кэширование с инвалидацией по префиксу (Redis). 
* Хранение файлов в MinIO (аватары, отчёты). Указаны бакеты. 
* Swagger UI: `/swagger/index.html`. 

## Требования

* Go 1.22+
* PostgreSQL с расширением `uuid-ossp` (используется `uuid_generate_v4()`). 
* Redis (необязательно, но рекомендуется)
* MinIO (для аватаров/отчётов)

## Переменные окружения

Ты прав — в README я не расписал все ENV’ы. Ниже дал обновлённый блок «Environment Variables» для README.md, где перечислены **все** переменные из `internal/config/config.go` (включая детальные `LOG_*`) и то, что используется в `compose.yml` и `.env.example`. Можешь заменить старый раздел на этот.

```markdown
## Environment Variables

Бэкенд читает конфигурацию из ENV (см. `internal/config/config.go`). Полный перечень:

| Variable | Description | Defaults |
|---|---|---|
| `PG_URL` | PostgreSQL connection string | — :contentReference[oaicite:0]{index=0} |
| `APP_URL` | Public origin used for links | — :contentReference[oaicite:1]{index=1} |
| `HTTP_ADDR` | Listen address | `:8080` :contentReference[oaicite:2]{index=2} |
| `API_BASE_PATH` | API prefix | `/api` :contentReference[oaicite:3]{index=3} |
| `JWT_SECRET` | HS256 secret for signing tokens | — :contentReference[oaicite:4]{index=4} |
| `ACCESS_TTL` | Access token TTL | `15m` :contentReference[oaicite:5]{index=5} |
| `REFRESH_TTL` | Refresh token TTL | `720h` (30d) :contentReference[oaicite:6]{index=6} |
| `CORS` | Allowed origins (comma-separated) or `*` | `*` :contentReference[oaicite:7]{index=7} |

**MinIO**

| Variable | Description | Defaults |
|---|---|---|
| `MINIO_ENDPOINT` | MinIO host:port | `minio:9000` :contentReference[oaicite:8]{index=8} |
| `MINIO_ACCESS_KEY` | MinIO access key | — :contentReference[oaicite:9]{index=9} |
| `MINIO_SECRET_KEY` | MinIO secret key | — :contentReference[oaicite:10]{index=10} |
| `MINIO_USE_SSL` | Use TLS for MinIO | `false` :contentReference[oaicite:11]{index=11} |
| `MINIO_BUCKET_AVATARS` | Bucket for avatars | `avatars` :contentReference[oaicite:12]{index=12} |
| `MINIO_BUCKET_REPORTS` | Bucket for reports | `reports` :contentReference[oaicite:13]{index=13} |

> При старте приложение создаёт указанные бакеты, если их нет. :contentReference[oaicite:14]{index=14}

**Redis (cache)**

| Variable | Description | Defaults |
|---|---|---|
| `REDIS_ADDR` | Redis address | `redis:6379` :contentReference[oaicite:15]{index=15} |
| `REDIS_PASSWORD` | Redis password | empty :contentReference[oaicite:16]{index=16} |
| `REDIS_DB` | Redis DB number | `0` :contentReference[oaicite:17]{index=17} |

**Logging (`LOG_*`)**

| Variable | Description | Defaults |
|---|---|---|
| `LOG_LEVEL` | Level: `debug`/`info`/`warn`/`error` | `info` :contentReference[oaicite:18]{index=18} |
| `LOG_PATH` | File path to write logs (optional) | empty :contentReference[oaicite:19]{index=19} |
| `LOG_MAX_SIZE_MB` | Max file size before rotate (MB) | `20` :contentReference[oaicite:20]{index=20} |
| `LOG_MAX_BACKUPS` | Max rotated files to keep | `10` :contentReference[oaicite:21]{index=21} |
| `LOG_MAX_AGE_DAYS` | Days to keep old logs | `30` :contentReference[oaicite:22]{index=22} |
| `LOG_STDOUT` | Also log to stdout | `true` :contentReference[oaicite:23]{index=23} |

**Frontend**

| Variable | Description |
|---|---|
| `VITE_API_BASE_URL` | Полный базовый URL к backend API (например, `http://localhost:8080/api`). Устанавливается в `gradeflow-front/.env`. См. секцию Frontend. :contentReference[oaicite:24]{index=24}

### Примеры из Docker Compose и `.env.example`

* В `compose.yml` показано, как эти переменные применяются по-умолчанию при локальном запуске (включая `LOG_STDOUT` и `REDIS_ADDR`). :contentReference[oaicite:25]{index=25}
* В `.env.example` приведены удобные значения для локальной разработки. :contentReference[oaicite:26]{index=26}


## Быстрый старт (Docker Compose)

Запусти весь стек (Postgres, MinIO, Redis, API и фронтенд dev-сервер):

```bash
docker compose up --build
```

* API: `http://localhost:8080`
* Swagger UI: `http://localhost:8080/swagger/index.html`
* Frontend (Vite): `http://localhost:5173`

Compose-описание уже в репозитории; при первом запуске миграции выполняются автоматически. 

> При первом запуске API создаёт администратора и логирует INS и пароль в старте (используется вход по INS). Смотри логи приложения. 

## Локальный запуск (без Docker)

```bash
# 1) Подними Postgres, Redis, MinIO (локально/в Docker), создай .env
# 2) Запусти сервер
go run ./cmd/api
```

База и Swagger-база пути конфигурируются в `cmd/api/main.go` через конфиг (`APIBasePath`), Swagger UI остаётся на `/swagger/index.html`. 

## Swagger

* UI: `http://localhost:8080/swagger/index.html`
* Если менялся код контроллеров/DTO, перегенерируй спецификацию:

```bash
swag init -g cmd/api/main.go -o pkg/swagger
```

Файлы доков лежат в `pkg/swagger`. 

## Миграции

Миграции описаны через gormigrate в `internal/migrations/migrations.go` и применяются автоматически при старте API. 

## Тесты

Запуск юнит-тестов:

```bash
go test ./...
```

В репозитории есть тесты сервисов и контроллеров. 

## Полезные детали домена

* Модели используют UUID и каскадные связи (Group/Subject/Teacher/ClassSession/Grade). 
* Пагинация (`limit`, `offset`, `search`) унифицирована через DTO. 
