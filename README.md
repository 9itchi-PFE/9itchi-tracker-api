# 9itchi-tracker-api

Email open tracking: unique tracking IDs, invisible pixel endpoint, per-email stats. Publishes open events.

Port `8086`, health `/health`. Skeleton only: every endpoint returns `501 not implemented` until built.

## Endpoints

| Method | Path | Access | Handler |
|---|---|---|---|
| GET | `/t/:trackingId` | public | Pixel |
| POST | `/internal/tracking` | internal | RegisterTracking |
| GET | `/emails/:id/stats` | user | EmailStats |

## Run

```
cp .env.example .env
go run main.go
```
