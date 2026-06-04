package consumer

import (
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Consumer subscribes to NATS JetStream and processes meter reading events.
type Consumer struct {
	sub       *nats.Subscription
	pool      *pgxpool.Pool
	ch        clickhouse.Conn
	threshold float64
	log       zerolog.Logger
}

// New creates a Consumer. sub must be a pull subscriber with explicit ACK policy.
func New(sub *nats.Subscription, pool *pgxpool.Pool, ch clickhouse.Conn, threshold float64) *Consumer {
	return &Consumer{
		sub:       sub,
		pool:      pool,
		ch:        ch,
		threshold: threshold,
		log:       log.With().Str("component", "consumer").Logger(),
	}
}
