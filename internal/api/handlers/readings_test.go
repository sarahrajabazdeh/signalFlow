package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"signalflow/internal/api/handlers"
)

func TestHandleCreateReading(t *testing.T) {
	js := testJetStream(t)

	body := `{"asset_id":"00000000-0000-0000-0000-000000000001","actual_output":420,"recorded_at":"2026-06-03T10:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/readings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.HandleCreateReading(js, "readings.test")(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("want 202, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleCreateReading_InvalidBody(t *testing.T) {
	js := testJetStream(t)

	req := httptest.NewRequest(http.MethodPost, "/readings", strings.NewReader("bad json"))
	w := httptest.NewRecorder()

	handlers.HandleCreateReading(js, "readings.test")(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", w.Code)
	}
}
