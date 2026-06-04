package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

type healthResponse struct {
	Status   string `json:"status"`
	Postgres string `json:"postgres"`
	NATS     string `json:"nats"`
}

// HandleHealth returns the health status of the API and its dependencies.
//
// @Summary      Health check
// @Tags         health
// @Produce      json
// @Success      200  {object}  healthResponse
// @Failure      503  {object}  healthResponse
// @Router       /health [get]
func HandleHealth(pool *pgxpool.Pool, nc *nats.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		resp := healthResponse{Status: "ok", Postgres: "up", NATS: "up"}
		code := http.StatusOK

		if err := pool.Ping(ctx); err != nil {
			resp.Postgres = "down"
			resp.Status = "degraded"
			code = http.StatusServiceUnavailable
		}

		if nc.Status() != nats.CONNECTED {
			resp.NATS = "down"
			resp.Status = "degraded"
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(resp)
	}
}
