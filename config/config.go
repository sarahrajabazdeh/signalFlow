package config

import (
	"os"
	"strconv"
)

// Config holds all environment-based configuration for signalFlow.
// Every value comes from environment variables — no hardcoded secrets.
type Config struct {
	// API
	APIPort string

	// PostgreSQL — operational data (assets, alerts)
	DatabaseURL string

	// ClickHouse — analytics data (readings time series)
	ClickHouseURL string

	// NATS JetStream
	NATSURL     string
	NATSStream  string
	NATSSubject string

	// Imbalance detection threshold (default 0.20 = 20%)
	ImbalanceThreshold float64

	// Auth
	JWTSecret string
	APIKey    string
}

// Load reads config from environment variables with sensible defaults for local dev.
func Load() Config {
	return Config{
		APIPort:            getEnv("API_PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/signalflow?sslmode=disable"),
		ClickHouseURL:      getEnv("CLICKHOUSE_URL", "clickhouse://localhost:9000/signalflow"),
		NATSURL:            getEnv("NATS_URL", "nats://localhost:4222"),
		NATSStream:         getEnv("NATS_STREAM", "READINGS"),
		NATSSubject:        getEnv("NATS_SUBJECT", "readings.ingest"),
		ImbalanceThreshold: parseFloat(getEnv("IMBALANCE_THRESHOLD", "0.20")),
		JWTSecret:          getEnv("JWT_SECRET", "dev-secret-change-in-production"),
		APIKey:             getEnv("API_KEY", "dev-api-key"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func parseFloat(s string) float64 {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.20
	}
	return val
}
