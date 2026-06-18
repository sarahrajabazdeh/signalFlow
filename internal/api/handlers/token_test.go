package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"signalflow/internal/api/handlers"
)

const (
	testJWTSecret = "test-jwt-secret"
	testAPIKey    = "test-api-key"
)

func TestHandleToken_CorrectAPIKey_ReturnsToken(t *testing.T) {
	handler := handlers.HandleToken(testJWTSecret, testAPIKey)

	body, _ := json.Marshal(map[string]string{"api_key": testAPIKey})
	req := httptest.NewRequest(http.MethodPost, "/token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal("response is not valid JSON")
	}
	if resp["token"] == "" {
		t.Error("expected non-empty token in response")
	}
}

func TestHandleToken_WrongAPIKey_Returns401(t *testing.T) {
	handler := handlers.HandleToken(testJWTSecret, testAPIKey)

	body, _ := json.Marshal(map[string]string{"api_key": "wrong-key"})
	req := httptest.NewRequest(http.MethodPost, "/token", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestHandleToken_MissingBody_Returns400(t *testing.T) {
	handler := handlers.HandleToken(testJWTSecret, testAPIKey)

	req := httptest.NewRequest(http.MethodPost, "/token", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
