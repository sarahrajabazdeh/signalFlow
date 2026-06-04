package consumer

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"signalflow/internal/db"
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
