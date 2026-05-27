-- Migration 003: Create readings table in ClickHouse
-- Readings are meter readings from assets — high volume, time series, analytics
-- Stored in ClickHouse — NOT PostgreSQL

-- MergeTree is ClickHouse's main table engine
-- ORDER BY (asset_id, recorded_at) means:
--   data is sorted by asset + time on disk → fast time range queries per asset
-- PARTITION BY toYYYYMM(recorded_at) means:
--   data is split into monthly chunks → fast to drop old data, fast range scans

CREATE TABLE IF NOT EXISTS readings (
    id              UUID        NOT NULL,
    asset_id        UUID        NOT NULL,
    actual_output   Float64     NOT NULL,           -- kW
    recorded_at     DateTime    NOT NULL,           -- when the asset measured this
    created_at      DateTime    NOT NULL DEFAULT now() -- when we received it
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(recorded_at)
ORDER BY (asset_id, recorded_at)
SETTINGS index_granularity = 8192;
