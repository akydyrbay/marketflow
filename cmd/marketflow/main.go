package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"marketflow/internal/adapters/cache"
	"marketflow/internal/adapters/db"
	"marketflow/internal/api"
	"marketflow/internal/app/mode"
	"marketflow/internal/domain"
	"marketflow/pkg/config"
	"marketflow/pkg/logger"
)

// func main() {
// 	ctx := context.Background()
// 	updates := make(chan m.PriceUpdate, 100)

// 	adapter := exchange.ExchangeAdapter{
// 		Addr:        "localhost:40101",
// 		Outbound:    updates,
// 		BackoffBase: 500 * time.Millisecond,
// 	}

// 	go adapter.Run(ctx)

// 	for update := range updates {
// 		fmt.Printf("[DATA] %s → %.2f @ %d\n", update.Symbol, update.Price, update.Timestamp)
// 	}
// }

// func main() {
// 	cache := &c.RedisCache{Addr: "localhost:6379"}
// 	if err := cache.Connect(); err != nil {
// 		panic(err)
// 	}
// 	now := time.Now().Unix()
// 	cache.AddPrice("exchange1", "BTCUSDT", 25000.5, now)
// 	cache.AddPrice("exchange1", "BTCUSDT", 25010.7, now+1)
// 	cache.Cleanup("exchange1", "BTCUSDT", now-60)
// 	// Inspect in redis-cli:
// 	// > redis-cli ZRANGE prices:exchange1:BTCUSDT 0 -1 WITHSCORES
// }

func main() {
	// loads all the configs
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config: %w", err)
	}

	// connect to the postgres
	repo, err := db.NewPostgres()
	if err != nil {
		log.Fatalf("failed to init postgres: %v", err)
	}
	// defer repo.Close()

	// connect to the redis
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	cache := cache.NewRedis(cfg.Redis.DB, redisAddr, cfg.Redis.Password, cfg.RedisTTL)
	defer cache.Close()

	// create aggregation for processing price updates
	inputChan := make(chan domain.PriceUpdate, 10000)
	// aggr := aggr.NewAggregator(inputChan, repo, cache, cfg.AggregatorWindow)

	// start the manager live/test
	manager := mode.NewManager(cfg)
	// go aggr.Start(context.Background())

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// if err := manager.Start(ctx, inputChan, mode.Test); err != nil {
	// 	log.Fatalf("failed to start test mode: %v", err)
	// }
	// start the api
	server := api.NewServer(repo, cache, manager)
	srv := &http.Server{
		Addr:    cfg.APIAddr,
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
