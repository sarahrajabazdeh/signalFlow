package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"signalflow/internal/api/handlers"
	"signalflow/internal/models"
)

func TestHandleListAlerts_Empty(t *testing.T) {
	pool := testPostgresPool(t)

	req := httptest.NewRequest(http.MethodGet, "/alerts", nil)
	w := httptest.NewRecorder()

	handlers.HandleListAlerts(pool)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var got []models.Alert
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got == nil {
		t.Error("want empty slice [], got nil (would serialize as JSON null)")
	}
}

func TestHandleListAlerts_InvalidAssetID(t *testing.T) {
	pool := testPostgresPool(t)

	req := httptest.NewRequest(http.MethodGet, "/alerts?asset_id=not-a-uuid", nil)
	w := httptest.NewRecorder()

	handlers.HandleListAlerts(pool)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", w.Code)
	}
}
