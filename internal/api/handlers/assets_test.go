package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"signalflow/internal/api/handlers"
	"signalflow/internal/models"
)

func TestHandleCreateAsset(t *testing.T) {
	pool := testPostgresPool(t)

	body := `{"name":"Solar Farm Alpha","expected_output":500}`
	req := httptest.NewRequest(http.MethodPost, "/assets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.HandleCreateAsset(pool)(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d: %s", w.Code, w.Body.String())
	}
	var got models.Asset
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Name != "Solar Farm Alpha" {
		t.Errorf("want name Solar Farm Alpha, got %s", got.Name)
	}
	if got.ExpectedOutput != 500 {
		t.Errorf("want expected_output 500, got %f", got.ExpectedOutput)
	}
}

func TestHandleCreateAsset_InvalidBody(t *testing.T) {
	pool := testPostgresPool(t)

	req := httptest.NewRequest(http.MethodPost, "/assets", strings.NewReader("not json"))
	w := httptest.NewRecorder()

	handlers.HandleCreateAsset(pool)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", w.Code)
	}
}
