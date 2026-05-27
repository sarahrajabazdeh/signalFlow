package queries

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"signalflow/internal/models"
)

// SaveAsset inserts a new asset into PostgreSQL.
func SaveAsset(ctx context.Context, pool *pgxpool.Pool, asset models.Asset) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO assets (id, name, expected_output, created_at)
		VALUES ($1, $2, $3, $4)
	`, asset.ID, asset.Name, asset.ExpectedOutput, asset.CreatedAt)
	if err != nil {
		return fmt.Errorf("save asset: %w", err)
	}
	return nil
}

// GetAssetByID fetches a single asset by its ID.
// Used by the consumer to check expected_output before imbalance detection.
func GetAssetByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (*models.Asset, error) {
	var a models.Asset
	err := pool.QueryRow(ctx, `
		SELECT id, name, expected_output, created_at
		FROM assets
		WHERE id = $1
	`, id).Scan(&a.ID, &a.Name, &a.ExpectedOutput, &a.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get asset by id: %w", err)
	}
	return &a, nil
}

// ListAssets returns all assets ordered by creation time.
func ListAssets(ctx context.Context, pool *pgxpool.Pool) ([]models.Asset, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, name, expected_output, created_at
		FROM assets
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list assets: %w", err)
	}
	defer rows.Close()

	var assets []models.Asset
	for rows.Next() {
		var a models.Asset
		if err := rows.Scan(&a.ID, &a.Name, &a.ExpectedOutput, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan asset: %w", err)
		}
		assets = append(assets, a)
	}
	return assets, rows.Err()
}
