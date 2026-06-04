package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"signalflow/internal/db"
	"signalflow/internal/db/queries"
	"signalflow/internal/models"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := db.NewPostgresPool(context.Background(),
		"postgres://postgres:postgres@localhost:5432/signalflow?sslmode=disable")
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

func testClickHouse(t *testing.T) clickhouse.Conn {
	t.Helper()
	conn, err := db.NewClickHouseConn(context.Background(), "clickhouse://localhost:9000/signalflow")
	if err != nil {
		t.Skipf("clickhouse not available: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func testJetStream(t *testing.T) nats.JetStreamContext {
	t.Helper()
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		t.Skipf("nats not available: %v", err)
	}
	t.Cleanup(func() { nc.Close() })
	js, err := nc.JetStream()
	if err != nil {
		t.Fatalf("jetstream context: %v", err)
	}
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "READINGS_TEST",
		Subjects: []string{"readings.test"},
	})
	if err != nil && !errors.Is(err, nats.ErrStreamNameAlreadyInUse) {
		t.Fatalf("add test stream: %v", err)
	}
	return js
}

// testConsumer creates a Consumer and returns it together with its pull subscription
// and JetStream context (both needed to publish and fetch test messages).
func testConsumer(t *testing.T) (*Consumer, *nats.Subscription, nats.JetStreamContext) {
	t.Helper()
	pool := testPool(t)
	ch := testClickHouse(t)
	js := testJetStream(t)

	// unique durable name per test so parallel tests don't share a consumer
	durable := "test-" + strings.ReplaceAll(t.Name(), "/", "-")
	sub, err := js.PullSubscribe("readings.test", durable, nats.AckExplicit())
	if err != nil {
		t.Fatalf("pull subscribe: %v", err)
	}
	t.Cleanup(func() { sub.Unsubscribe() })

	return New(sub, pool, ch, 0.20), sub, js
}

func TestProcess_HappyPath(t *testing.T) {
	c, sub, js := testConsumer(t)

	asset := models.Asset{
		ID:             uuid.New(),
		Name:           "Happy Path Asset",
		ExpectedOutput: 500,
		CreatedAt:      time.Now().UTC(),
	}
	if err := queries.SaveAsset(context.Background(), c.pool, asset); err != nil {
		t.Fatalf("save test asset: %v", err)
	}

	// 450 > 400 (80% of 500) — above threshold, no alert expected
	req := models.CreateReadingRequest{
		AssetID:      asset.ID,
		ActualOutput: 450,
		RecordedAt:   time.Now().UTC(),
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, err := js.Publish("readings.test", data); err != nil {
		t.Fatalf("publish: %v", err)
	}
	msgs, err := sub.Fetch(1, nats.MaxWait(2*time.Second))
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}

	c.process(context.Background(), msgs[0])

	var count uint64
	if err := c.ch.QueryRow(context.Background(),
		"SELECT count() FROM readings WHERE asset_id = ?", asset.ID,
	).Scan(&count); err != nil {
		t.Fatalf("query clickhouse: %v", err)
	}
	if count == 0 {
		t.Error("want reading in ClickHouse, got 0")
	}

	alerts, err := queries.ListAlerts(context.Background(), c.pool, &asset.ID)
	if err != nil {
		t.Fatalf("list alerts: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("want 0 alerts, got %d", len(alerts))
	}
}

func TestProcess_Imbalance(t *testing.T) {
	c, sub, js := testConsumer(t)

	asset := models.Asset{
		ID:             uuid.New(),
		Name:           "Imbalance Asset",
		ExpectedOutput: 500,
		CreatedAt:      time.Now().UTC(),
	}
	if err := queries.SaveAsset(context.Background(), c.pool, asset); err != nil {
		t.Fatalf("save test asset: %v", err)
	}

	// 300 < 400 (80% of 500) — below threshold, alert expected
	req := models.CreateReadingRequest{
		AssetID:      asset.ID,
		ActualOutput: 300,
		RecordedAt:   time.Now().UTC(),
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, err = js.Publish("readings.test", data); err != nil {
		t.Fatalf("publish: %v", err)
	}
	msgs, err := sub.Fetch(1, nats.MaxWait(2*time.Second))
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}

	c.process(context.Background(), msgs[0])

	alerts, err := queries.ListAlerts(context.Background(), c.pool, &asset.ID)
	if err != nil {
		t.Fatalf("list alerts: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("want 1 alert, got %d", len(alerts))
	}
	if alerts[0].ActualOutput != 300 {
		t.Errorf("want actual_output 300, got %f", alerts[0].ActualOutput)
	}
	if alerts[0].ExpectedOutput != 500 {
		t.Errorf("want expected_output 500, got %f", alerts[0].ExpectedOutput)
	}
}

func TestProcess_BadJSON(t *testing.T) {
	c, sub, js := testConsumer(t)

	if _, err := js.Publish("readings.test", []byte("not valid json")); err != nil {
		t.Fatalf("publish: %v", err)
	}
	msgs, err := sub.Fetch(1, nats.MaxWait(2*time.Second))
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}

	c.process(context.Background(), msgs[0])
	// NAK is called internally — no DB writes expected, no panic
}

func TestProcess_AssetNotFound(t *testing.T) {
	c, sub, js := testConsumer(t)

	unknownID := uuid.New()
	req := models.CreateReadingRequest{
		AssetID:      unknownID,
		ActualOutput: 450,
		RecordedAt:   time.Now().UTC(),
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if _, err = js.Publish("readings.test", data); err != nil {
		t.Fatalf("publish: %v", err)
	}
	msgs, err := sub.Fetch(1, nats.MaxWait(2*time.Second))
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}

	c.process(context.Background(), msgs[0])

	var count uint64
	if err := c.ch.QueryRow(context.Background(),
		"SELECT count() FROM readings WHERE asset_id = ?", unknownID,
	).Scan(&count); err != nil {
		t.Fatalf("query clickhouse: %v", err)
	}
	if count != 0 {
		t.Errorf("want 0 readings for unknown asset, got %d", count)
	}
}
