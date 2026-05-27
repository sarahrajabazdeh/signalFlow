package models

import (
	"time"

	"github.com/google/uuid"
)

// Asset represents an energy generating unit (solar farm, wind turbine, battery).
// expected_output is what this asset should produce under normal conditions.
type Asset struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	ExpectedOutput float64   `json:"expected_output"` // in kW
	CreatedAt      time.Time `json:"created_at"`
}

// Reading is a meter reading sent by an asset.
// Stored in ClickHouse for time series analytics.
type Reading struct {
	ID           uuid.UUID `json:"id"`
	AssetID      uuid.UUID `json:"asset_id"`
	ActualOutput float64   `json:"actual_output"` // in kW
	RecordedAt   time.Time `json:"recorded_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// Alert is created when actual output drops more than 20% below expected.
// Stored in PostgreSQL for operational querying.
type Alert struct {
	ID             uuid.UUID `json:"id"`
	AssetID        uuid.UUID `json:"asset_id"`
	ActualOutput   float64   `json:"actual_output"`
	ExpectedOutput float64   `json:"expected_output"`
	DeviationPct   float64   `json:"deviation_pct"` // e.g. 0.35 means 35% below
	TriggeredAt    time.Time `json:"triggered_at"`
}

// CreateAssetRequest is the JSON body for POST /assets
type CreateAssetRequest struct {
	Name           string  `json:"name"`
	ExpectedOutput float64 `json:"expected_output"`
}

// CreateReadingRequest is the JSON body for POST /readings
type CreateReadingRequest struct {
	AssetID      uuid.UUID `json:"asset_id"`
	ActualOutput float64   `json:"actual_output"`
	RecordedAt   time.Time `json:"recorded_at"`
}
