# signalFlow

A backend system that ingests live meter readings from energy assets (solar farms, wind turbines, battery storage), processes them asynchronously, and automatically raises an alert when actual output drops more than 20% below expected.

## Architecture

```
HTTP Client
    │
    ▼
API Process (:8080)          ← accepts requests, publishes to NATS
    │
    │ NATS JetStream
    ▼
Consumer Process             ← reads from NATS, writes to databases
    │
    ├── PostgreSQL            assets, alerts
    └── ClickHouse            readings (time series)
```

## Tech Stack

Go · chi · NATS JetStream · PostgreSQL · ClickHouse · JWT · Prometheus · Docker Compose

## Prerequisites

- Go 1.25+
- Docker + Docker Compose
- `atlas` CLI — [atlasgo.io](https://atlasgo.io/getting-started)

## Startup

```bash
# 1. Copy env config
cp .env.example .env

# 2. Start infrastructure
make docker-up

# 3. Run migrations
export $(cat .env | xargs)
make migrate

# 4. Start API (terminal 1)
make run-api

# 5. Start consumer (terminal 2)
make run-consumer
```

## Try it

```bash
# Get a token
curl -s -X POST localhost:8080/token \
  -H "Content-Type: application/json" \
  -d '{"api_key":"dev-api-key"}'

# Use the token on all other endpoints
curl localhost:8080/alerts \
  -H "Authorization: Bearer <token>"
```

Full end-to-end test (requires running stack):

```bash
./scripts/test-flow.sh
```

## Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/token` | — | Get a JWT token |
| GET | `/health` | — | Health check |
| POST | `/assets` | ✓ | Create an energy asset |
| POST | `/readings` | ✓ | Submit a meter reading |
| GET | `/alerts` | ✓ | List imbalance alerts |
| GET | `/metrics` | — | Prometheus metrics |
| GET | `/swagger/` | — | API docs |
