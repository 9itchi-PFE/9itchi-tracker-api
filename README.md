# 9itchi-tracker-api

Email open tracking: unique tracking IDs, invisible pixel endpoint, per-email stats. Publishes open events.

Port `8086`, health `/health`.

## Endpoints

| Method | Path | Access | Handler | Notes |
|---|---|---|---|---|
| GET | `/t/:trackingId` | public | Pixel | Always `200` with a 1x1 transparent GIF (even for unknown IDs). Counts the open unless it happens within 10 s of `sent_at`. Publishes `email.opened` on the first counted open. |
| POST | `/internal/tracking` | internal (`x-api-key`) | RegisterTracking | Body `{email_id, user_id}` (UUIDs). Returns `{tracking_id, pixel_url_path: "/t/<id>"}`. Idempotent per `email_id`. |
| GET | `/emails/:id/stats` | user | EmailStats | `:id` is the email ID, scoped to the caller (else `404`). Returns `sent_at`, `first_open_at`, `last_open_at`, `open_count`, last 20 `opens[].opened_at`. |

Opens are logged in `tracking_opens` with a truncated user agent and `sha256(ip + ITCHI_SECRET_KEY)`; raw IPs are never stored.

## Run

```
cp .env.example .env
go run main.go
```

## Test

```
go test ./...
go test -tags integration -count=1 ./repositories/...
```
