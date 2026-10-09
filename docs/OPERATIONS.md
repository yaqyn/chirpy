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

## One-command container setup

Requires Docker Engine/Desktop with Compose v2 and OpenSSL. From a checkout:

```bash
./scripts/threadline start
```

For a fresh installation:

```bash
git clone https://github.com/yaqyn/threadline-social-api.git && cd threadline-social-api && ./scripts/threadline start
```

The launcher checks Docker access before changing files. It generates random
JWT, webhook, and database credentials once into `.threadline.env` with mode
600. Existing settings and your native `.env` are preserved. Database startup,
Goose migrations, and API startup are ordered through Compose dependencies;
startup returns only when service health checks pass. Subsequent starts apply
pending migrations without resetting the data. First builds need internet
access and can take several minutes.

| Command | Behavior |
|---|---|
| `./scripts/threadline start` | Build, migrate, start, and wait for health |
| `./scripts/threadline stop` | Stop containers; preserve volume and credentials |
| `./scripts/threadline status` | Show service state, including migration job |
| `./scripts/threadline logs` | Follow API logs; Ctrl+C leaves the API running |
| `./scripts/threadline doctor` | Check Docker, Compose, and configuration |
| `./scripts/threadline help` | Show command help without requiring Docker |

The API is exposed only on `127.0.0.1:8080`; PostgreSQL on `127.0.0.1:5433`.
Edit `API_PORT` and `DATABASE_PORT` in `.threadline.env` to avoid occupied ports.
The launcher defaults to a non-dev platform, so destructive reset is disabled.
The generated Polka key is for local webhook experiments; replace it with your
provider's real key if integrating a real provider.

Do not change `POSTGRES_PASSWORD` after database initialization without also
changing the database role password. Container environment changes do not
rewrite credentials inside an existing volume. Stop/start preserves data.
Do not use `docker compose down -v` unless intentionally discarding the database.

Advanced Compose users can supply their own env file:

```bash
docker compose --project-name threadline --env-file .threadline.env up -d --build --wait
```

The runtime and migration images use non-root users. Only public files enter the
runtime image; settings and secrets are excluded from both Git and build context.
The migration service runs Goose before the API starts. For hosted deployments,
run migrations as one release job rather than concurrently across replicas.

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
