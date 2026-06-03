package handlers_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"signalflow/internal/db"
)

func testPostgresPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := db.NewPostgresPool(context.Background(),
		"postgres://postgres:postgres@localhost:5432/signalflow?sslmode=disable")
	if err != nil {
		t.Skipf("postgres not available: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
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
	js.AddStream(&nats.StreamConfig{
		Name:     "READINGS_TEST",
		Subjects: []string{"readings.test"},
	})
	return js
}
