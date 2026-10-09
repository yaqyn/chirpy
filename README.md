![Threadline — Social Publishing API](docs/readme/cover.png)

# Threadline

**The backend behind the conversation.**

A Go and PostgreSQL API for short-form social publishing. Threadline connects
accounts, authenticated posts, renewable sessions, and paid membership events
through a small, explicit HTTP contract.

[![Verify](https://github.com/yaqyn/threadline-social-api/actions/workflows/ci.yml/badge.svg)](https://github.com/yaqyn/threadline-social-api/actions/workflows/ci.yml)
![Go](https://img.shields.io/badge/Go-1.27.2+-39d6c0?labelColor=101827)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-39d6c0?labelColor=101827)
![API](https://img.shields.io/badge/API-OpenAPI_3.0-f4f0e7?labelColor=101827)

[Quick start](docs/OPERATIONS.md) · [API reference](docs/API.md) ·
[OpenAPI specification](docs/openapi.json) · [Architecture](docs/ARCHITECTURE.md) ·
[Security & operations](docs/SECURITY.md)

## <img src="docs/readme/icons/overview.svg" width="26" height="26" alt=""> Publish with identity

Create an account, sign in, and publish a message of up to **140 Unicode code
points**. The access token determines the author; a caller cannot choose another
user’s identity. Public feeds support author filtering, chronological sorting,
and bounded pagination. Only an author can delete their own post.

| Capability | Behavior |
|---|---|
| Accounts | Create an account; update your own email and password |
| Authentication | Argon2id password hashes; signed one-hour access tokens |
| Sessions | Random 256-bit refresh tokens, valid for 60 days and revocable |
| Publishing | Validate and clean posts before persisting them |
| Discovery | Filter by author; sort oldest or newest first; paginate |
| Membership | API-key-protected Polka webhooks upgrade a user |
| Operations | Process readiness, static-request metrics, development-only reset |

Built from the Boot.dev **Chirpy** backend project and extended with runtime
hardening, regression tests, pagination, database indexes, CI, and documentation.
The course’s `/api/chirps`, `is_chirpy_red`, and `chirpy-access` names remain in the
API for compatibility. Threadline is the project brand; this is a backend,
not a complete social-network frontend.

## <img src="docs/readme/icons/flow.svg" width="26" height="26" alt=""> One account. A continuous session.

```mermaid
sequenceDiagram
    participant C as Client
    participant A as Threadline API
    participant D as PostgreSQL
    C->>A: Email + password
    A->>D: Look up password hash
    A->>A: Compare with Argon2id
    A->>D: Store 60-day refresh token
    A-->>C: User + access token + refresh token
    C->>A: Publish with access token
    A->>A: Verify JWT and derive author ID
    A->>D: Save validated chirp
    A-->>C: 201 Created
    C->>A: Refresh with refresh token
    A->>D: Check expiration and revocation
    A-->>C: New one-hour access token
```

A refresh token cannot publish a post. An access token cannot refresh a session.
Revoking a refresh token stops future renewals; an already issued access token
remains usable until its expiration.

## <img src="docs/readme/icons/setup.svg" width="26" height="26" alt=""> Run locally

**Requirements:** Go 1.27.2 or newer, PostgreSQL 15 or newer, Bash, and the
`psql` client. Docker Compose is an optional way to run PostgreSQL and the API.

```bash
git clone https://github.com/yaqyn/threadline-social-api.git
cd threadline-social-api
cp .env.example .env
chmod 600 .env
openssl rand -base64 64
```

Put the generated value in `JWT_SECRET` in `.env`. Set `POLKA_KEY` to your
provider key, and update `DB_URL` to a database you own. The sample values are
local placeholders. Do not commit `.env`.

```bash
# Create the database using your PostgreSQL credentials.
createdb -h localhost -U postgres threadline
make tools
make migrate-up
make run
```

```bash
curl -i http://localhost:8080/api/healthz
# HTTP 200, text/plain; body: OK
```

The server reads environment variables directly; `.env` is optional when the
variables are already supplied. Run the binary from the repository root so it
can find `public/`. See [container setup and configuration](docs/OPERATIONS.md)
for Compose, migration, and deployment instructions.

| Variable | Purpose | Default |
|---|---|---|
| `DB_URL` | PostgreSQL connection URL | Required |
| `JWT_SECRET` | HS256 signing secret, at least 32 characters | Required |
| `POLKA_KEY` | Incoming webhook credential | Required |
| `ADDR` | HTTP listening address | `:8080` |
| `PLATFORM` | Enables reset only when exactly `dev` | Reset disabled |
| `TEST_DB_URL` | Disposable, migrated integration-test database | Tests skip without it |

## <img src="docs/readme/icons/api.svg" width="26" height="26" alt=""> An explicit HTTP contract

| Method | Endpoint | Authorization | Success |
|---|---|---|---|
| `POST` | `/api/users` | None | `201` User |
| `PUT` | `/api/users` | Access token | `200` User |
| `POST` | `/api/login` | Email + password | `200` User + both tokens |
| `POST` | `/api/refresh` | Refresh token | `200` Access token |
| `POST` | `/api/revoke` | Refresh token | `204` Empty |
| `POST` | `/api/chirps` | Access token | `201` Chirp |
| `GET` | `/api/chirps` | None | `200` Chirp array |
| `GET` | `/api/chirps/{chirpID}` | None | `200` Chirp |
| `DELETE` | `/api/chirps/{chirpID}` | Access token, author only | `204` Empty |
| `POST` | `/api/polka/webhooks` | Polka API key | `204` Empty |
| `GET` | `/api/healthz` | None | `200` Plain text |
| `GET` | `/admin/metrics` | None | `200` HTML |
| `POST` | `/admin/reset` | Development environment only | `200` Empty |
| `GET` | `/app/` | None | Static public page |

```bash
curl -sS -X POST http://localhost:8080/api/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"your-local-password"}'

curl -sS -X POST http://localhost:8080/api/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"your-local-password"}'

# Use the token field from login, not refresh_token.
curl -sS -X POST http://localhost:8080/api/chirps \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"body":"A new thread starts here."}'

curl -sS 'http://localhost:8080/api/chirps?sort=desc&limit=20&offset=0'
```

The feed accepts `author_id=<UUID>`, `sort=asc|desc`, `limit=1..100`, and
`offset=0..2147483647`. Defaults are ascending order, 100 records, and offset 0.
Filtering, sorting, and pagination happen in PostgreSQL. Equal timestamps use
ID ascending as a stable tie-breaker. No results returns `[]`.

[Full request/response examples and error semantics →](docs/API.md)

## <img src="docs/readme/icons/security.svg" width="26" height="26" alt=""> Keep credentials out of the conversation

- Passwords are hashed with salted Argon2id; user resources never contain hashes.
- Access tokens require HS256, the expected issuer, and an expiration claim.
- Authorization comes from verified token subjects, not submitted user IDs.
- Webhook keys are checked before an event is processed.
- Only `public/` is served; source files and `.env` are outside the static root.
- JSON requests are capped at 64 KiB; oversized payloads receive `413`.
- HTTP timeouts and a bounded database pool limit resource usage.
- Shutdown drains requests for up to 10 seconds before exiting.

**Deployment boundary:** terminate HTTPS at a trusted reverse proxy, apply
request-rate limits, and keep `PLATFORM` out of `dev`. Refresh tokens are stored
as opaque plaintext credentials in this implementation; protect database access
and backups. Metrics are public and reset is intentionally a course development
tool. [Read the operating constraints before hosting →](docs/SECURITY.md)

## <img src="docs/readme/icons/data.svg" width="26" height="26" alt=""> Small schema, clear ownership

```mermaid
erDiagram
    USERS ||--o{ CHIRPS : authors
    USERS ||--o{ REFRESH_TOKENS : owns
    USERS {
        uuid id PK
        timestamp created_at
        timestamp updated_at
        text email UK
        text hashed_password
        boolean is_chirpy_red
    }
    CHIRPS {
        uuid id PK
        timestamp created_at
        timestamp updated_at
        text body
        uuid user_id FK
    }
    REFRESH_TOKENS {
        text token PK
        timestamp created_at
        timestamp updated_at
        uuid user_id FK
        timestamp expires_at
        timestamp revoked_at
    }
```

Deleting a user cascades to their chirps and refresh tokens. Goose owns schema
migrations; SQLC generates typed Go queries from SQL. Composite feed indexes and
an index on refresh-token ownership support the main access patterns.

## <img src="docs/readme/icons/quality.svg" width="26" height="26" alt=""> Verify the behavior

```bash
make check       # go vet, race-enabled tests, formatting
make build       # reproducible binary build
make generate    # regenerate typed database queries

# Point only at a migrated disposable database.
TEST_DB_URL='postgres://USER:PASSWORD@localhost:5432/threadline_test?sslmode=disable' \
  make integration
```

The test suite covers password hashing, JWT expiry and signatures, header
extraction, JSON validation, query parameters, profanity replacement, static-file
isolation, authorization guards, and real PostgreSQL account/session workflows.
CI also checks SQLC drift, migration rollback/reapply, and the container build.
Database integration tests skip locally unless `TEST_DB_URL` is supplied.

| Location | Responsibility |
|---|---|
| `main.go` | Configuration, connection pool, server lifecycle |
| `router.go`, `middleware.go` | Routes, limits, response headers |
| `handlers_*.go` | Accounts, chirps, administration, payment events |
| `validation.go`, `http_helpers.go` | Input rules and JSON responses |
| `internal/auth/` | Passwords, JWTs, refresh randomness, headers |
| `internal/database/` | SQLC-generated code; do not edit by hand |
| `sql/schema/`, `sql/queries/` | Schema history and database operations |
| `docs/` | API, architecture, operations, security, and brand assets |

## <img src="docs/readme/icons/roadmap.svg" width="26" height="26" alt=""> Where it goes next

The implemented scope is deliberately focused. Post editing, social graphs,
media uploads, search, refresh-token rotation, and a full client are future work.
Membership status is persisted; it does not yet unlock an editing endpoint.

Before broader deployment: add gateway rate limiting, hash stored refresh
credentials, define retention/cleanup policies, and replace development admin
operations with authenticated administration.

## <img src="docs/readme/icons/owner.svg" width="26" height="26" alt=""> Built by Abdulrahman M. Yaqyn

Designed, developed, and maintained by [Abdulrahman M. Yaqyn](https://github.com/yaqyn).
Origin: the Boot.dev Chirpy project, extended into Threadline.
