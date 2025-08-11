package app

import (
	"context"
	"log"
	"log/slog"
	"marketflow/internal/adapters/cache"
	"marketflow/internal/adapters/db"
	"marketflow/internal/adapters/exchange"
	"marketflow/internal/api/server"
	"marketflow/internal/domain"
	"marketflow/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func SetupApp() (*http.Server, func()) {
	// initialize logger
	logger.Init()

	// connect to Postgres
	repo := db.NewPostgres()

	// connect to Redis
	cache := cache.NewRedis()

	exchange := exchange.NewLiveModeFetcher()

	datafetch := server.NewDataFetcher(exchange, repo, cache)

	if err := datafetch.ListenAndSave(); err != nil {
		logger.Error("Failed to start data fetcher", "error", err)
		exchange.Close()
		os.Exit(1)
	}

	router := server.Setup(repo, cache, datafetch)
	srv := &http.Server{
		Addr:    ":" + *domain.Port,
		Handler: router,
	}

	cleanup := func() {
		slog.Info("Cleaning up resources...")
		cache.Close()
		repo.Close()
		datafetch.StopListening()
	}

	return srv, cleanup
}

func StartServer(srv *http.Server) {
	go func() {
		slog.Info("Starting server at " + *domain.Port + "...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error: ", err.Error())
		}
	}()
}

func WaitForShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	slog.Info("Shutdown signal received...")
}

func ShutdownServer(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slog.Info("Shutting down HTTP server...")
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
	} else {
		slog.Info("Server gracefully stopped.")
	}
}
