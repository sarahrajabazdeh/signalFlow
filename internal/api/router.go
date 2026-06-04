package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"signalflow/config"
	_ "signalflow/internal/api/docs"
	"signalflow/internal/api/handlers"
	apimid "signalflow/internal/api/middleware"
)

func NewRouter(pool *pgxpool.Pool, js nats.JetStreamContext, nc *nats.Conn, cfg config.Config) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(apimid.RequestID)
	r.Use(apimid.Metrics)

	r.Get("/health", handlers.HandleHealth(pool, nc))
	r.Handle("/metrics", promhttp.Handler())
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Post("/assets", handlers.HandleCreateAsset(pool))
	r.Post("/readings", handlers.HandleCreateReading(js, cfg.NATSSubject))
	r.Get("/alerts", handlers.HandleListAlerts(pool))

	return r
}
