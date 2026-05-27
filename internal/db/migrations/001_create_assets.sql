-- Migration 001: Create assets table
-- Assets are energy generating units (solar farms, wind turbines, batteries)
-- Stored in PostgreSQL — operational data, relational, low volume

CREATE EXTENSION IF NOT EXISTS "pgcrypto"; -- needed for gen_random_uuid()

CREATE TABLE IF NOT EXISTS assets (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT        NOT NULL,
    expected_output NUMERIC     NOT NULL CHECK (expected_output > 0), -- kW, must be positive
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- index for name lookups
CREATE INDEX IF NOT EXISTS idx_assets_name ON assets (name);
