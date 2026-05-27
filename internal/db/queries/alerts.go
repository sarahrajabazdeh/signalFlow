package queries

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"signalflow/internal/models"
)

// SaveAlert inserts an imbalance alert into PostgreSQL.
// Called by the consumer when actual output < 80% of expected.
func SaveAlert(ctx context.Context, pool *pgxpool.Pool, alert models.Alert) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO alerts (id, asset_id, actual_output, expected_output, deviation_pct, triggered_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, alert.ID, alert.AssetID, alert.ActualOutput, alert.ExpectedOutput, alert.DeviationPct, alert.TriggeredAt)
	if err != nil {
		return fmt.Errorf("save alert: %w", err)
	}
	return nil
}

// ListAlerts returns all alerts, optionally filtered by asset_id.
// Used by GET /alerts and GET /alerts?asset_id=xxx
func ListAlerts(ctx context.Context, pool *pgxpool.Pool, assetID *uuid.UUID) ([]models.Alert, error) {
	query := `
		SELECT id, asset_id, actual_output, expected_output, deviation_pct, triggered_at
		FROM alerts
	`
	args := []any{}

	// if asset_id filter is provided, add WHERE clause
	if assetID != nil {
		query += " WHERE asset_id = $1"
		args = append(args, *assetID)
	}

	query += " ORDER BY triggered_at DESC"

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	defer rows.Close()

	var alerts []models.Alert
	for rows.Next() {
		var a models.Alert
		if err := rows.Scan(
			&a.ID, &a.AssetID, &a.ActualOutput,
			&a.ExpectedOutput, &a.DeviationPct, &a.TriggeredAt,
		); err != nil {
			return nil, fmt.Errorf("scan alert: %w", err)
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}
