-- Migration 003: Create readings table in ClickHouse
-- Readings are meter readings from assets — high volume, time series, analytics
-- Stored in ClickHouse — NOT PostgreSQL
-- Uses ReplacingMergeTree to handle at-least-once delivery:
--   if the same (asset_id, recorded_at) is inserted twice, the one with higher version wins

-- ReplacingMergeTree deduplicates based on ORDER BY key during merges
-- ORDER BY (asset_id, recorded_at, version) ensures:
--   data is sorted by asset + time + version on disk
--   duplicate (asset_id, recorded_at) pairs get deduplicated by keeping max(version)
-- PARTITION BY toYYYYMM(recorded_at) means:
--   data is split into monthly chunks → fast to drop old data, fast range scans

CREATE TABLE IF NOT EXISTS readings (
    id              UUID        NOT NULL,
    asset_id        UUID        NOT NULL,
    actual_output   Float64     NOT NULL,           -- kW
    recorded_at     DateTime    NOT NULL,           -- when the asset measured this
    created_at      DateTime    NOT NULL DEFAULT now(), -- when we received it
    version         UInt32      NOT NULL DEFAULT 1  -- for dedup: max(version) wins
)
ENGINE = ReplacingMergeTree(version)
PARTITION BY toYYYYMM(recorded_at)
ORDER BY (asset_id, recorded_at, version)
SETTINGS index_granularity = 8192;
