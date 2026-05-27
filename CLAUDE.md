# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## Project

**signalFlow** — Real-Time Energy Event Processing System

A backend system that reliably ingests live meter readings from distributed energy assets (solar farms, wind turbines, battery storage), processes them asynchronously via a message broker, persists them to PostgreSQL, and automatically detects imbalance risk when actual production deviates more than 20% below expected output.

## Tech Stack

| Layer | Technology | Purpose |
|---|---|---|
| Language | Go | Primary language |
| HTTP Router | [chi](https://github.com/go-chi/chi) | REST API routing |
| Message Broker | NATS JetStream | Durable event queue, at-least-once delivery |
| Operational DB | PostgreSQL | Assets, alerts, transactional data |
| Analytics DB | ClickHouse OSS | Readings time series, aggregations, dashboards |
| Infrastructure | Docker Compose | Local dev environment |
| Config | Environment variables | All configuration |
| Logging | `log/slog` | Structured logging |

## Architecture

Two separate processes (decoupled by design):

```
┌─────────────────────────────────────────────────────────┐
│  API Process                                            │
│  - REST endpoints (chi router)                          │
│  - POST /assets         → create asset                  │
│  - POST /readings       → publish event to JetStream    │
│  - GET  /alerts         → query alerts (filter by asset)│
└───────────────────────┬─────────────────────────────────┘
                        │ publishes to NATS JetStream
                        ▼
┌─────────────────────────────────────────────────────────┐
│  Consumer Process                                       │
│  - Subscribes to meter-reading event stream             │
│  - Validates payload                                    │
│  - Writes reading to PostgreSQL                         │
│  - ACKs only after successful write (no silent drops)   │
│  - Triggers imbalance check:                            │
│      if actual < expected * 0.80 → create alert        │
└─────────────────────────────────────────────────────────┘
                        │
                        ├──────────────────────────────┐
                        ▼                              ▼
                  PostgreSQL                      ClickHouse
              (assets, alerts)              (readings time series)
              operational data                 analytical data
```

## Key Constraints

- API must return immediately — never block on downstream processing
- At-least-once delivery: message only ACKed after successful DB write
- Consumer restarts must not lose any readings
- Imbalance threshold: actual output < 80% of expected → create alert
- Graceful shutdown with proper context cancellation
- All config via environment variables

## Folder Structure (planned)

```
signalFlow/
├── CLAUDE.md
├── docker-compose.yml          # NATS + PostgreSQL + ClickHouse
├── go.mod
├── go.sum
├── cmd/
│   ├── api/
│   │   └── main.go             # API process entrypoint
│   └── consumer/
│       └── main.go             # Consumer process entrypoint
├── internal/
│   ├── api/
│   │   ├── handlers/           # HTTP handlers
│   │   └── router.go           # chi router setup
│   ├── consumer/
│   │   └── consumer.go         # NATS subscriber + imbalance logic
│   ├── db/
│   │   ├── migrations/         # SQL migration files
│   │   └── queries/            # DB query functions
│   └── models/
│       └── models.go           # Asset, Reading, Alert structs
├── config/
│   └── config.go               # Env-based config loading
└── .env.example                # Template for environment variables
```

## Database Schema (planned)

```sql
-- assets: energy generating units with expected output
CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    expected_output NUMERIC NOT NULL,  -- in kW or MW
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- readings: meter readings from assets
CREATE TABLE readings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID REFERENCES assets(id),
    actual_output NUMERIC NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- alerts: imbalance events (actual < 80% of expected)
CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id UUID REFERENCES assets(id),
    actual_output NUMERIC NOT NULL,
    expected_output NUMERIC NOT NULL,
    deviation_pct NUMERIC NOT NULL,
    triggered_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Environment Variables

```env
# PostgreSQL (operational)
DATABASE_URL=postgres://postgres:postgres@localhost:5432/signalflow?sslmode=disable

# ClickHouse (analytics)
CLICKHOUSE_URL=clickhouse://localhost:9000/signalflow

# NATS
NATS_URL=nats://localhost:4222
NATS_STREAM=READINGS
NATS_SUBJECT=readings.ingest

# API
API_PORT=8080

# Imbalance
IMBALANCE_THRESHOLD=0.20
```

## Build, Lint & Test

```bash
# Install dependencies
go mod tidy

# Run API process
go run ./cmd/api

# Run Consumer process
go run ./cmd/consumer

# Run all tests
go test ./...

# Run a single test
go test ./internal/consumer/... -run TestImbalanceDetection

# Lint
golangci-lint run

# Start infrastructure (NATS + PostgreSQL)
docker-compose up -d
```

## Success Criteria

The system is complete when an operator can:
1. ✅ Create an asset with an expected output
2. ✅ Submit a meter reading via API and receive an immediate response
3. ✅ Observe the consumer process the event and store the reading in PostgreSQL
4. ✅ See an alert automatically created when actual output is >20% below expected
5. ✅ Retrieve alerts through the API
6. ✅ Ingestion endpoint remains responsive under sustained load
7. ✅ No readings are lost across consumer restarts

## Implementation Notes

- Use `context.Context` everywhere for proper cancellation/timeout propagation
- NATS JetStream stream must be created before publishing (idempotent `AddStream`)
- Consumer should use `push` or `pull` subscribe with explicit ACK (`msg.Ack()`)
- Use `pgx` or `database/sql` with `lib/pq` for PostgreSQL
- Use `clickhouse-go/v2` driver for ClickHouse
- Consumer writes to BOTH PostgreSQL and ClickHouse before ACKing
- PostgreSQL stores: assets, alerts (operational — low volume, transactional)
- ClickHouse stores: readings (analytical — high volume, time series)
- Structured logging: attach `asset_id` and `reading_id` to every log line
- Do not use ORM — raw SQL queries preferred for clarity
- Readings insert must be idempotent: ON CONFLICT (asset_id, recorded_at) DO NOTHING
