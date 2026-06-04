package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"signalflow/config"
	"signalflow/internal/consumer"
	"signalflow/internal/db"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := db.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("connect postgres")
	}
	defer pool.Close()

	chConn, err := db.NewClickHouseConn(ctx, cfg.ClickHouseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("connect clickhouse")
	}
	defer chConn.Close()

	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		log.Fatal().Err(err).Msg("connect nats")
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		log.Fatal().Err(err).Msg("jetstream context")
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     cfg.NATSStream,
		Subjects: []string{cfg.NATSSubject},
	})
	if err != nil && !errors.Is(err, nats.ErrStreamNameAlreadyInUse) {
		log.Fatal().Err(err).Msg("create nats stream")
	}

	sub, err := js.PullSubscribe(cfg.NATSSubject, "readings-consumer", nats.AckExplicit())
	if err != nil {
		log.Fatal().Err(err).Msg("pull subscribe")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Info().Msg("shutdown signal received")
		cancel()
	}()

	log.Info().Msg("consumer starting")
	c := consumer.New(sub, pool, chConn, cfg.ImbalanceThreshold)
	if err := c.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("consumer error")
	}
	log.Info().Msg("consumer stopped")
}
