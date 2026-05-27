package queries

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"signalflow/internal/models"
)

// SaveReading inserts a meter reading into ClickHouse.
// Uses ON CONFLICT equivalent — ClickHouse deduplicates by (asset_id, recorded_at)
// via ReplacingMergeTree or we handle it at insert time.
func SaveReading(ctx context.Context, conn clickhouse.Conn, reading models.Reading) error {
	err := conn.Exec(ctx, `
		INSERT INTO readings (id, asset_id, actual_output, recorded_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, reading.ID, reading.AssetID, reading.ActualOutput, reading.RecordedAt, reading.CreatedAt)
	if err != nil {
		return fmt.Errorf("save reading to clickhouse: %w", err)
	}
	return nil
}
