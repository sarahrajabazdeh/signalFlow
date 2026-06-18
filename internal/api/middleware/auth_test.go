package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"signalflow/internal/api/middleware"
)

const testSecret = "test-secret"

func makeToken(secret string, exp time.Time) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "device",
		ExpiresAt: jwt.NewNumericDate(exp),
	})
	signed, _ := token.SignedString([]byte(secret))
	return signed
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestAuth_MissingHeader_Returns401(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/alerts", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuth_MalformedHeader_Returns401(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))

	req := httptest.NewRequest(http.MethodGet, "/alerts", nil)
	req.Header.Set("Authorization", "NotBearer sometokenvalue")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuth_InvalidSignature_Returns401(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))

	token := makeToken("wrong-secret", time.Now().Add(time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/alerts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuth_ExpiredToken_Returns401(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))

	token := makeToken(testSecret, time.Now().Add(-time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/alerts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuth_ValidToken_PassesThrough(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(okHandler))

	token := makeToken(testSecret, time.Now().Add(time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/alerts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
