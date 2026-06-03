package db

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// NewClickHouseConn creates a connection to ClickHouse.
// Used by the consumer to insert readings (time series data).
// clickhouseURL format: clickhouse://host:port/database
func NewClickHouseConn(ctx context.Context, clickhouseURL string) (clickhouse.Conn, error) {
	parsed, err := url.Parse(clickhouseURL)
	if err != nil {
		return nil, fmt.Errorf("parse clickhouse url: %w", err)
	}

	host := parsed.Hostname()
	port := parsed.Port()
	if port == "" {
		port = "9000"
	}
	database := parsed.Path
	if database != "" && database[0] == '/' {
		database = database[1:]
	}
	if database == "" {
		database = "signalflow"
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", host, port)},
		Auth: clickhouse.Auth{
			Database: database,
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
