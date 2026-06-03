package queries

import (
	"context"
	"fmt"

	"signalflow/internal/models"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// SaveReading inserts a meter reading into ClickHouse.
// ReplacingMergeTree deduplicates on (asset_id, recorded_at):
// if the same reading is inserted twice (at-least-once delivery),
// the one with higher version is kept during merge.
// For simplicity, all inserts use version=1 since rows are identical anyway.
func SaveReading(ctx context.Context, conn clickhouse.Conn, reading models.Reading) error {
	err := conn.Exec(ctx, `
		INSERT INTO readings (id, asset_id, actual_output, recorded_at, created_at, version)
		VALUES (?, ?, ?, ?, ?, 1)
	`, reading.ID, reading.AssetID, reading.ActualOutput, reading.RecordedAt, reading.CreatedAt)
	if err != nil {
		return fmt.Errorf("save reading to clickhouse: %w", err)
	}
	return nil
}
