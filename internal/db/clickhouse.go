package db

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// NewClickHouseConn creates a connection to ClickHouse.
// Used by the consumer to insert readings (time series data).
func NewClickHouseConn(ctx context.Context, clickhouseURL string) (clickhouse.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"localhost:9000"},
		Auth: clickhouse.Auth{
			Database: "signalflow",
		},
		Protocol: clickhouse.Native, // native binary protocol — faster than HTTP
	})
	if err != nil {
		return nil, fmt.Errorf("open clickhouse: %w", err)
	}

	// verify the connection is working
	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping clickhouse: %w", err)
	}

	return conn, nil
}
