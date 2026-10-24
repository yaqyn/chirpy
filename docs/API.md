# Threadline API reference

Base URL: `http://localhost:8080`. Import [openapi.json](openapi.json) into an
OpenAPI-aware client. API responses are compact JSON unless an endpoint says
otherwise. Wrong HTTP methods receive `405`; Go also accepts `HEAD` for GET routes.

## Resource shapes

User:

```json
{"id":"3311741c-680c-4546-99f3-fc9efac2036c","created_at":"2026-10-09T10:00:00Z","updated_at":"2026-10-09T10:00:00Z","email":"you@example.com","is_chirpy_red":false}
```

Chirp:

```json
{"id":"94b7e44c-3604-42e3-bef7-ebfcc3efff8f","created_at":"2026-10-09T10:00:00Z","updated_at":"2026-10-09T10:00:00Z","body":"Hello!","user_id":"3311741c-680c-4546-99f3-fc9efac2036c"}
```

JSON errors use `{"error":"Human-readable explanation"}`. Router-generated
errors, such as a wrong HTTP method, can be plain text. Passwords and hashes
never appear in User resources. Timestamps use RFC 3339 JSON serialization;
PostgreSQL timestamps are stored using the existing course schema.

## Accounts

`POST /api/users` takes `{"email":"you@example.com","password":"local-password"}`.
Returns `201` and User. Email must be a plain email address, not a display-name
form. Password must be nonempty and at most 1,024 bytes. Duplicate email is `409`.
Email matching and uniqueness are case-sensitive in this schema.

`PUT /api/users` takes the same body and requires `Authorization: Bearer ACCESS_TOKEN`.
Returns `200` and the updated User. The JWT subject selects the account; submitted
IDs do not select another user. `created_at` is preserved and `updated_at` changes.
Missing/invalid/expired access token is `401`; duplicate email is `409`.
Changing a password does not revoke already-issued sessions in this implementation.

`POST /api/login` takes email/password and returns `200` with all User fields,
plus `token` and `refresh_token`. Incorrect email or password is `401` with
`Incorrect email or password`. Access tokens last one hour; refresh tokens last
60 days. The old `expires_in_seconds` input no longer controls expiration.

## Session renewal and revocation

```bash
curl -i -X POST http://localhost:8080/api/refresh \
  -H "Authorization: Bearer $REFRESH_TOKEN"

curl -i -X POST http://localhost:8080/api/revoke \
  -H "Authorization: Bearer $REFRESH_TOKEN"
```

Neither endpoint needs a body. Refresh returns `200` and `{"token":"NEW_JWT"}`.
Missing, unknown, expired, or revoked refresh tokens return `401`. The token is
not rotated by refresh. A JWT is not a valid refresh credential.

Revoke sets `revoked_at` and `updated_at` and returns `204`, with no body. Unknown
well-formed token strings also return `204` (there is nothing left to revoke).
Missing or malformed authorization headers return `401`. Access tokens already
issued remain valid until expiration.

## Publish and delete

`POST /api/chirps` requires `Authorization: Bearer ACCESS_TOKEN` and a body:

```json
{"body":"An idea worth sharing."}
```

Returns `201` and Chirp. Identity comes only from the JWT. Empty/whitespace-only
posts and posts above 140 Unicode code points are rejected with `400`. Count is
code points, not bytes or visual grapheme clusters. Standalone space-delimited
`kerfuffle`, `sharbert`, and `fornax` are replaced with `****`, ignoring case.
Punctuation changes the word: `Sharbert!` is left untouched. Validation happens
before cleaning, and original spacing is preserved.

`GET /api/chirps/{chirpID}` returns `200` and Chirp, or `404` when missing.
Malformed UUIDs receive `400`.

`DELETE /api/chirps/{chirpID}` requires an access token. Returns `204` when deleted,
`403` if another user owns it, `404` if missing, and `401` for invalid credentials.
Authorization is checked in Go and included in the SQL deletion predicate.

## Feed queries

```text
GET /api/chirps
GET /api/chirps?sort=desc
GET /api/chirps?author_id=3311741c-680c-4546-99f3-fc9efac2036c&sort=asc&limit=20&offset=40
GET /api/chirps?sort=newest&limit=20&page=3
```

| Query | Valid values | Default |
|---|---|---|
| `author_id` | UUID | All authors |
| `sort` | `asc`, `desc`, `oldest`, `newest` | `asc` |
| `limit` | Integer 1–100 | 100 |
| `offset` | Integer 0–2147483647 | 0 |
| `page` | Positive integer; cannot be combined with `offset` | Omitted |

Returns `200` and an array; an empty result is `[]`. Invalid values return `400`.
Empty query values behave as omitted parameters. Ordering is by `created_at`
with ID ascending as a tie-breaker. Querying and pagination happen in PostgreSQL.
Offset pagination is simple; insertions/deletions between requests can shift
page boundaries. There is no total-count field or cursor in the current API.

## Payment webhook

```bash
curl -i -X POST http://localhost:8080/api/polka/webhooks \
  -H "Authorization: ApiKey $POLKA_KEY" \
  -H 'Content-Type: application/json' \
  -d '{"event":"user.upgraded","data":{"user_id":"3311741c-680c-4546-99f3-fc9efac2036c"}}'
```

The key is checked first. Wrong/missing key returns `401`. `user.upgraded` sets
`is_chirpy_red=true` and updates the user timestamp. Success is `204`; unknown
user is `404`. Other events are immediately acknowledged with `204` and do not
require valid user data. Repeated upgrade events are safe: membership remains
true. Polka is a course payment-provider integration, not a live billing SDK.

## Operations

- `GET /api/healthz`: `200`, `Content-Type: text/plain; charset=utf-8`, body `OK`.
  This checks process availability; it is not a continuous database probe.
- `GET /admin/metrics`: HTML count of requests handled by the public fileserver.
  This counts requests, including failed static requests, not unique visitors.
- `POST /admin/reset`: allowed only with `PLATFORM=dev`; deletes all users and
  cascading records, and zeros the counter. Otherwise `403`. No authentication
  is added to this course development tool; keep it off shared environments.
- `/app/`: serves only `public/`. No application secrets or source are served.

## Input and error handling

Requests are limited to 64 KiB. The shared decoder rejects malformed JSON,
`null`, incompatible types, and multiple JSON documents. It permits extra JSON
fields for compatibility. Wrong `Content-Type` is not independently rejected.
Most endpoints return `400` for invalid input and `413` for oversized JSON.
Unexpected database or encoding failures return `500` without exposing internals.
