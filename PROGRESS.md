# signalFlow — Build Progress

> Update this file every time you finish a phase.
> When you come back, read this first — then tell Claude "continue from where we left off".

---

## Current Status

**→ System complete.**

Run `./scripts/test-flow.sh` to verify the full pipeline end-to-end.

---

## Phases

| Phase | What | Status |
|---|---|---|
| 1 | Scaffold (docker-compose, go.mod, config, models) | ✅ Done |
| 2 | Database migrations (PostgreSQL + ClickHouse) | ✅ Done |
| 3 | API process (chi router, handlers, NATS publish) | ✅ Done |
| 4 | Consumer process (NATS subscribe, save to DBs, ACK) | ✅ Done |
| 5 | Imbalance detection (actual < 80% expected → alert) | ✅ Done |
| 6 | Verify end-to-end (/test-flow) | ✅ Done |

---

## What's Running

```bash
docker compose up -d   # already done — 3 containers running
```

| Container | Status |
|---|---|
| signalflow-postgres | ✅ healthy — port 5432 |
| signalflow-clickhouse | ✅ healthy — port 9000 |
| signalflow-nats | ✅ healthy — port 4222 |

---

## What Was Built in Phase 1

```
signalFlow/
├── CLAUDE.md                    ✅ full project context
├── PROGRESS.md                  ✅ this file
├── docker-compose.yml           ✅ NATS + PostgreSQL + ClickHouse
├── go.mod                       ✅ Go module
├── .env.example                 ✅ env var template
├── config/config.go             ✅ env-based config
├── internal/models/models.go    ✅ Asset, Reading, Alert structs
└── .claude/
    ├── settings.json            ✅ project permissions
    ├── skills/                  ✅ /stack-up /stack-down /db-check /test-flow /nats-check
    └── agents/                  ✅ code-reviewer debugger db-reader system-design-expert
```

---

## Key Decisions Made

- **PostgreSQL** stores: assets, alerts (operational, transactional)
- **ClickHouse** stores: readings (time series, analytics)
- **Consumer ACKs** only after both DB writes succeed
- **API returns 202** (accepted, not processed) — async by design
- **Imbalance threshold**: actual < expected × 0.80

---

## How to Resume

1. Open VSCode in this folder
2. Read this file
3. Run `docker compose up -d` if containers are stopped
4. Tell Claude: *"continue from where we left off"*
