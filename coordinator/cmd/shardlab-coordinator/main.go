package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/marceterrone10/shardlab/coordinator"
)

func main() {
	if err := run(); err != nil {
		slog.Error("coordinator stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	config, err := coordinator.LoadConfig()
	if err != nil {
		return err
	}

	router, err := coordinator.NewRouter(config.Shards)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              config.HTTPAddress,
		Handler:           coordinator.NewServer(router).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errs := make(chan error, 1)
	go func() {
		slog.Info("coordinator listening",
			"address", config.HTTPAddress,
			"shards", len(config.Shards),
		)
		errs <- server.ListenAndServe()
	}()

	select {
	case err := <-errs:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}
