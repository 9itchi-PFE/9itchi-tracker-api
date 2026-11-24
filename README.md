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

## Events (RabbitMQ)

Optional. With `RABBITMQ_URL` set, `email.opened` (first open of an email: `email_id`, `user_id`, `tracking_id`, `opened_at`) is published to the topic exchange `9itchi.events` (routing key = event name) as persistent JSON messages, each confirmed by the broker and carrying a message id consumers deduplicate on. Without the variable the event name is only logged.

A publish failure is logged and never fails the request. A broker that cannot be reached is left alone for 15 s, so it costs one short wait and not one per request; events raised meanwhile are lost (there is no outbox). Broker tests: `RABBITMQ_TEST_URL=amqp://guest:guest@localhost:5672/ go test -tags integration ./core/` (skipped without the variable).

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
