# Security and deployment boundaries

This repository is a backend project, not a claim of audited production security.
The controls below are implemented and tested; deployment still has work to do.

## Implemented controls

- Salted Argon2id password hashing; no hashes in user JSON.
- HS256 JWT validation with issuer, expiration, and UUID subject validation.
- Random 32-byte refresh credentials, 60-day expiration, and persisted revocation.
- Authenticated ownership checks on publishing, account updates, and deletion.
- API-key-protected membership webhooks with constant-time key comparison.
- Dedicated public static root, 64 KiB request bounds, HTTP timeouts,
  response security headers, and bounded connection pooling.
- Development reset disabled unless `PLATFORM` is exactly `dev`.

## Required hosting decisions

Terminate TLS at a trusted proxy and use encrypted database connections in
shared environments. Local `sslmode=disable` is a development setting. Add
rate limits for login/signup and webhook delivery; Argon2id deliberately consumes
memory/CPU, so unbounded concurrent authentication is not appropriate publicly.
Do not log authorization headers, passwords, or database credentials.

Provision random secrets through your platform’s secret manager. `.env.example`
contains placeholders, not production values. The course Polka key must not be
used as a public deployment secret. Do not expose `/admin/reset`; the `dev`
environment check is intentionally not an authenticated admin system.

Refresh tokens currently exist in plaintext in PostgreSQL. Restrict database
roles and backups; hash refresh credentials before a production rollout.
Changing a password does not invalidate access or refresh tokens automatically.
JWT revocation is delayed until expiration. Refresh tokens are not rotated,
and expired rows are not automatically cleaned up. Webhooks authenticate by
shared key and tolerate repeats; there is no provider event-ID ledger, signed
payload, or replay window.

The app uses header credentials, not browser cookies. A client is responsible
for safe storage; do not place long-lived secrets in publicly accessible client
code. Cross-origin policy and CSRF protection need review if cookie auth is added.

## Verification

Run `make check`, integration tests with `TEST_DB_URL`, and a vulnerability scan:

```bash
go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...
```

Tests do not replace threat modeling, dependency maintenance, or infrastructure
monitoring. Report security findings privately to the repository owner through
GitHub contact channels; avoid publishing working credentials in an issue.
