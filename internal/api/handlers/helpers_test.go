package handlers_test

import (
	"context"
	"strings"
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
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "READINGS_TEST",
		Subjects: []string{"readings.test"},
	})
	if err != nil {
		// stream may already exist — that's fine, anything else is a setup failure
		if !strings.Contains(err.Error(), "stream name already in use") {
			t.Fatalf("create test stream: %v", err)
		}
	}
	return js
}

func testNATSConn(t *testing.T) *nats.Conn {
	t.Helper()
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		t.Skipf("nats not available: %v", err)
	}
	t.Cleanup(func() { nc.Close() })
	return nc
}
