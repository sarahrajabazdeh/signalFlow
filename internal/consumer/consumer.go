package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"signalflow/internal/db/queries"
	"signalflow/internal/models"
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

// process handles a single NATS message: decodes it, validates the asset,
// writes the reading to ClickHouse, and creates a PostgreSQL alert if the
// actual output falls below the imbalance threshold.
func (c *Consumer) process(ctx context.Context, msg *nats.Msg) {
	var req models.CreateReadingRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		c.log.Warn().Err(err).Msg("invalid message payload")
		msg.Nak()
		return
	}

	asset, err := queries.GetAssetByID(ctx, c.pool, req.AssetID)
	if err != nil {
		c.log.Warn().Err(err).Str("asset_id", req.AssetID.String()).Msg("asset not found")
		msg.Nak()
		return
	}

	reading := models.Reading{
		ID:           uuid.New(),
		AssetID:      req.AssetID,
		ActualOutput: req.ActualOutput,
		RecordedAt:   req.RecordedAt,
		CreatedAt:    time.Now().UTC(),
	}

	if err := queries.SaveReading(ctx, c.ch, reading); err != nil {
		c.log.Error().Err(err).
			Str("asset_id", req.AssetID.String()).
			Str("reading_id", reading.ID.String()).
			Msg("failed to save reading")
		msg.Nak()
		return
	}

	if req.ActualOutput < asset.ExpectedOutput*(1-c.threshold) {
		alert := models.Alert{
			ID:             uuid.New(),
			AssetID:        asset.ID,
			ActualOutput:   req.ActualOutput,
			ExpectedOutput: asset.ExpectedOutput,
			DeviationPct:   (asset.ExpectedOutput - req.ActualOutput) / asset.ExpectedOutput,
			TriggeredAt:    time.Now().UTC(),
		}
		if err := queries.SaveAlert(ctx, c.pool, alert); err != nil {
			c.log.Error().Err(err).
				Str("asset_id", req.AssetID.String()).
				Str("reading_id", reading.ID.String()).
				Msg("failed to save alert")
			// do not NAK — reading is already committed to ClickHouse
		}
	}

	msg.Ack()
}

// Run starts the pull-subscribe loop and blocks until ctx is cancelled.
// A fetch timeout (no messages available) is normal idle behaviour — the loop continues.
func (c *Consumer) Run(ctx context.Context) error {
	c.log.Info().Msg("consumer loop starting")
	for {
		select {
		case <-ctx.Done():
			c.log.Info().Msg("consumer loop stopped")
			return nil
		default:
		}

		msgs, err := c.sub.Fetch(10, nats.MaxWait(5*time.Second))
		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				continue // no messages — normal when queue is empty
			}
			if ctx.Err() != nil {
				return nil // context cancelled while waiting
			}
			if errors.Is(err, nats.ErrConnectionClosed) || errors.Is(err, nats.ErrBadSubscription) {
				c.log.Error().Err(err).Msg("subscription lost")
				return nil
			}
			c.log.Error().Err(err).Msg("fetch error")
			continue
		}

		for _, msg := range msgs {
			c.process(ctx, msg)
		}
	}
}
