# 9itchi Tracker API
Email open tracking: invisible pixel, per-email open stats, `email.opened` event. Port 8086, DB `tracking`.

## Conventions (shared by every 9itchi Go service)
- `handlers/` thin (validate, authorize, call service, JSON) -> `services/` all logic, returns `(T, *models.ApiError)` -> `repositories/` only Postgres, plain errors (`ErrNotFound`, `ErrConflict`). Error body `{"error": "..."}`.
- DB calls use `repositories.QueryTimeout`; multi-table writes `Store.WithTx`.
- Schema: add `migrations/NNNN_name.sql` (goose, embedded, runs at startup). Never edit a merged migration.
- Auth: `Authorization: Bearer` AES-GCM-wrapped HS256 JWT (claims `id`, `email`, `role`, `typ`, `exp`; access only). `/internal/*` = `x-api-key` (`INTERNAL_API_KEY`), blocked by nginx.
- Events: exchange `9itchi.events` (topic, routing key = event name), names in `core/events.go`; empty `RABBITMQ_URL` = `LogPublisher`.
- Env only, validated in `utils/config.go`. Common: `PORT`, `ENVIRONMENT`, `DATABASE_URL`, `ITCHI_SECRET_KEY`, `ITCHI_ENCRYPTION_KEY`, `INTERNAL_API_KEY`, `INTERNAL_API_KEY_REQUIRED`. Never log tokens, passwords, PII.
- Tests: `{resource}_test.go`, testify, func-field fakes (no testify/mock), success + error paths. Repo tests: `//go:build integration`, embedded Postgres.
- Check: `gofmt -l .; go build ./... && go vet ./... && go test ./...`; `go test -tags integration -count=1 ./repositories/...`; `golangci-lint run`.

## Routes
- Public: `GET /t/:trackingId` (nginx `/t/`): 1x1 GIF, always 200, `no-store`, never reveals if the id exists.
- User: `GET /emails/:id/stats` (sent, first open, count, recent opens; 404 if not owner).
- Internal: `POST /internal/tracking` (`email_id`, `user_id`; returns tracking id + path; idempotent per email), `DELETE /internal/users/:id`.

## Data, events
- Tables: `tracking_records` (unique `tracking_id`, unique `email_id`), `tracking_opens` (`ip_hash char(64)`). One migration `0001_init.sql`.
- Publishes `email.opened` on first open only (`open_count == 1`); publish failure logged. Env: `RABBITMQ_URL`, `EMAIL_API_URL`.

## Domain rules
- Opens within 10 s of `sent_at` (`core.OpenIgnoreWindow`) are provider prefetch: ignored.
- Tracking id = 32 random bytes base64url (43 chars); longer ids rejected.
- IP stored only as `sha256(ip + ITCHI_SECRET_KEY)`; user agent truncated.

## Gotchas
- `pkg/email` client is an unused stub.
