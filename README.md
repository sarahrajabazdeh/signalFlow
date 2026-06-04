# signalFlow

Real-time energy event processing system. Ingests meter readings from distributed
energy assets (solar farms, wind turbines, battery storage), processes them
asynchronously via NATS JetStream, persists to ClickHouse, and automatically
detects imbalance risk when actual production deviates more than 20% below expected.

## Architecture

```
┌──────────────────────────────────┐
│  API Process (:8080)             │
│  POST /assets  → PostgreSQL      │
│  POST /readings → NATS JetStream │
│  GET  /alerts  ← PostgreSQL      │
│  GET  /health                    │
│  GET  /metrics (Prometheus)      │
│  GET  /swagger/                  │
└──────────────┬───────────────────┘
               │ NATS JetStream
               ▼
┌──────────────────────────────────┐
│  Consumer Process                │
│  Pull subscribe (at-least-once)  │
│  → Write reading to ClickHouse   │
│  → If actual < 80% expected:     │
│      Write alert to PostgreSQL   │
└──────────────────────────────────┘
```

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25 |
| HTTP Router | chi |
| Message Broker | NATS JetStream |
| Operational DB | PostgreSQL |
| Analytics DB | ClickHouse |
| Logging | zerolog |
| Metrics | Prometheus |
| Infrastructure | Docker Compose |

## Prerequisites

- Go 1.25+
- Docker + Docker Compose
- `golangci-lint` — [install](https://golangci-lint.run/usage/install/)
- `swag` CLI — `go install github.com/swaggo/swag/cmd/swag@latest`

## Quick Start

```bash
# 1. Start infrastructure
make docker-up

# 2. Start API (terminal 1)
make run-api

# 3. Start consumer (terminal 2)
make run-consumer

# 4. Test the full pipeline
./scripts/test-flow.sh
```

## API Endpoints

| Method | Path | Description |
|---|---|---|
| POST | /assets | Create an energy asset |
| POST | /readings | Submit a meter reading (async) |
| GET | /alerts | List imbalance alerts |
| GET | /health | Health check (PostgreSQL + NATS) |
| GET | /metrics | Prometheus metrics |
| GET | /swagger/ | Interactive API documentation |

## Environment Variables

Copy `.env.example` to `.env` and adjust as needed.

| Variable | Default | Description |
|---|---|---|
| `API_PORT` | `8080` | API server port |
| `DATABASE_URL` | `postgres://...` | PostgreSQL DSN |
| `CLICKHOUSE_URL` | `clickhouse://...` | ClickHouse DSN |
| `NATS_URL` | `nats://localhost:4222` | NATS server |
| `NATS_STREAM` | `READINGS` | JetStream stream name |
| `NATS_SUBJECT` | `readings.ingest` | JetStream subject |
| `IMBALANCE_THRESHOLD` | `0.20` | Alert threshold (20% below expected) |

## Running Tests

```bash
make test                    # unit + integration tests (skip if Docker down)
make lint                    # golangci-lint
./scripts/test-flow.sh       # end-to-end pipeline test
```

## Regenerating Swagger Docs

```bash
make swagger
```
