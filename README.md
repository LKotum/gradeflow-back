# GradeFlow API

Minimal setup and endpoints for auth, verification, and password reset.

## Environment

Required:
- `PG_URL` – Postgres connection URL
- `APP_URL` – Public base URL (e.g., https://example.com)
- `JWT_SECRET` – HS256 secret for access tokens

Optional SMTP (for emails):
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`
- `SMTP_FROM_EMAIL`, `SMTP_FROM_NAME`

## Run

Build and run with your usual Go tooling. Swagger UI is served at `/swagger/index.html` (BasePath `/api`).

### Local email testing (MailHog)

`compose.yml` includes MailHog. When running via Docker Compose, emails will be visible at http://localhost:8025 and SMTP is available on port 1025. The API service is pre-configured to use it.

## Auth endpoints

- POST `/api/auth/register` — { email, fullName, role, password }
- GET `/api/auth/verify?token=...` — activates account
- POST `/api/auth/login` — { email, password, totp? } → { access, refresh }
- POST `/api/auth/refresh` — { refresh } → new { access, refresh }
- POST `/api/auth/request-reset` — { email } — always returns ok=true
- POST `/api/auth/reset` — { token, password }

Notes:
- Login requires `status = active`.
- Verification and reset emails are sent only if SMTP config provided.

### Using protected routes

Send the access token in the Authorization header:

Authorization: Bearer <access>

Example:

1) Login and capture access token:

```bash
ACCESS=$(curl -s -X POST http://localhost:8080/api/auth/login \
	-H 'Content-Type: application/json' \
	-d '{"email":"user@example.local","password":"pass12345"}' | jq -r .access)
```

2) Call /auth/me with access token:

```bash
curl -s http://localhost:8080/api/auth/me -H "Authorization: Bearer $ACCESS" | jq
```
