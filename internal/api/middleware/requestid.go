package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// RequestID attaches a request ID to every request.
// Reads X-Request-ID header if present; otherwise generates a new UUID.
// Sets the ID on the response header and injects it into the zerolog context.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", id)
		logger := log.With().Str("request_id", id).Logger()
		r = r.WithContext(logger.WithContext(r.Context()))
		next.ServeHTTP(w, r)
	})
}
