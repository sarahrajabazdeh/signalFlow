package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"signalflow/config"
	"signalflow/internal/api/handlers"
)

func NewRouter(pool *pgxpool.Pool, js nats.JetStreamContext, cfg config.Config) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/assets", handlers.HandleCreateAsset(pool))
	r.Post("/readings", handlers.HandleCreateReading(js, cfg.NATSSubject))
	r.Get("/alerts", handlers.HandleListAlerts(pool))

	return r
}
