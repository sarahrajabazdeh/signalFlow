-- Migration 002: Create alerts table
-- Alerts are created when actual output drops > 20% below expected
-- Stored in PostgreSQL — operational data, queried by operators, low volume

CREATE TABLE IF NOT EXISTS alerts (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id        UUID        NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    actual_output   NUMERIC     NOT NULL,           -- what the asset was producing (kW)
    expected_output NUMERIC     NOT NULL,           -- what it should have produced (kW)
    deviation_pct   NUMERIC     NOT NULL,           -- e.g. 0.35 = 35% below expected
    triggered_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- index for filtering alerts by asset (GET /alerts?asset_id=...)
CREATE INDEX IF NOT EXISTS idx_alerts_asset_id    ON alerts (asset_id);
CREATE INDEX IF NOT EXISTS idx_alerts_triggered_at ON alerts (triggered_at DESC);
