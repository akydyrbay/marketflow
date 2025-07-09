package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"marketflow/internal/adapters/cache"
	"marketflow/internal/adapters/db"
	"marketflow/internal/api"
	"marketflow/internal/app/aggr"
	"marketflow/internal/app/mode"
	"marketflow/internal/domain"
	"marketflow/pkg/logger"
)

func main() {
	
	logger.Init()
	
	// connect to the postgres
	repo := db.NewPostgres()

	// defer repo.Close()

	// connect to the redis
	cache := cache.NewRedis()
	// defer cache.Close()

	// create aggregation for processing price updates
	inputChan := make(chan domain.PriceUpdate, 10000)

	// start the manager live/test
	manager := mode.NewManager()
	agg := aggr.NewAggregator(inputChan, repo, cache)

	go agg.Start(context.Background())
	// start the api
	server := api.NewServer(repo, cache, manager)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: server.Router(inputChan),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("API server failed", "error", err)
			log.Fatalf("API server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.Info("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("API shutdown error", "error", err)
	}
	logger.Info("shutdown complete")
}
