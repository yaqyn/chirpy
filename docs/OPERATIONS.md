# Operations and development

## Native setup

Use Go 1.27.2+, PostgreSQL 15+, and Bash. Run `make tools` to install pinned Goose
and SQLC binaries; ensure `$(go env GOPATH)/bin` is on PATH. Create a database,
copy `.env.example` to `.env`, and replace credentials and keys. Never reuse the
example values for a hosted service.

`make migrate-up` loads `.env` and runs all pending migrations. `make run` builds
and starts the binary. Running from the project root is required for the public
asset directory. An existing Chirpy database works with the preserved schema;
back it up before applying new migrations. Migration 006 adds only indexes.

## Docker Compose

Docker daemon access is required. This is a local setup with loopback-bound ports.

```bash
cp .env.example .env
# Edit JWT_SECRET, POLKA_KEY, and POSTGRES_PASSWORD first.
docker compose up -d db

# In .env, set DB_URL to the Compose database's host-facing URL:
# postgres://postgres:YOUR_PASSWORD@localhost:5433/threadline?sslmode=disable
make tools
make migrate-up
docker compose up -d --build api
curl -i http://localhost:8080/api/healthz
```

Compose passes the internal database URL to the API automatically. A named volume
persists PostgreSQL data. `docker compose down` keeps the volume; `down -v`
permanently deletes it. The runtime image contains the binary and public files,
uses a non-root user, and receives secrets through environment variables.
Migrations are an explicit step, not executed by each application replica.

## Day-to-day commands

| Command | Purpose |
|---|---|
| `make check` | Vet, race-enabled tests, formatting check |
| `make build` | Build `out` with trimmed source paths |
| `make run` | Build and run |
| `make generate` | SQLC regeneration after SQL changes |
| `make tools` | Install pinned Goose and SQLC |
| `make migrate-up` | Apply pending migrations |
| `make migrate-down` | Roll back one migration; review destructive SQL first |
| `make integration` | PostgreSQL tests; requires `TEST_DB_URL` |
| `python scripts/generate_openapi.py` | Regenerate API reference after contract edits |

The integration test creates uniquely named users and cleans up its own records;
it does not call reset. Still use a separate migrated test database. CI provides
one automatically and validates generated query consistency and image builds.

## Troubleshooting

- **Connection refused:** start PostgreSQL, verify the host/port and `DB_URL`.
- **Missing JWT_SECRET/POLKA_KEY:** supply environment variables or edit `.env`.
- **Address already in use:** stop the previous binary or set `ADDR=:8081`.
- **Missing relation:** apply migrations to the same database the app uses.
- **401 on refresh:** use `refresh_token`, not the login response’s `token`.
- **401 on publishing:** use the access token, not the refresh token.
- **403 on deletion:** the signed-in user is not the author.
- **403 on reset:** reset is disabled outside `PLATFORM=dev`.
- **Unknown sqlc/goose command:** add your Go binary directory to PATH.
- **Docker socket permission denied:** configure Docker access with your system’s
  supported instructions, or use native PostgreSQL and Go.

## Release checks

Back up persistent data, apply migrations once, build with the patched toolchain,
run checks and smoke requests, set a non-dev platform, and configure HTTPS and
proxy rate limits. Healthz indicates HTTP process availability, not ongoing
schema or database health. Rotate signing secrets intentionally: old JWTs signed
with a previous secret become invalid immediately.
