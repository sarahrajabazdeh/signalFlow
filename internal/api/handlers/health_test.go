package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"signalflow/internal/api/handlers"
)

func TestHandleHealth_Up(t *testing.T) {
	pool := testPostgresPool(t)
	nc := testNATSConn(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handlers.HandleHealth(pool, nc)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("want status ok, got %s", resp["status"])
	}
	if resp["postgres"] != "up" {
		t.Errorf("want postgres up, got %s", resp["postgres"])
	}
	if resp["nats"] != "up" {
		t.Errorf("want nats up, got %s", resp["nats"])
	}
}
