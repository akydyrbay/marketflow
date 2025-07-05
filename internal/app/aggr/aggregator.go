package aggr

import (
	"context"
	"os"
	"strings"
	"time"

	"marketflow/internal/domain"
	"marketflow/pkg/logger"
)

type Aggregator struct {
	Input  <-chan domain.PriceUpdate
	Repo   domain.PriceRepository
	Cache  domain.Cache
	Window time.Duration
}

func NewAggregator(input <-chan domain.PriceUpdate, repo domain.PriceRepository, cache domain.Cache) *Aggregator {
	windowStr := os.Getenv("AGGREGATOR_WINDOW")
	if windowStr == "" {
		windowStr = "1m" // Default to 1 minute
	}

	window, err := time.ParseDuration(windowStr)
	if err != nil {
		logger.Error("failed to parse AGGREGATOR_WINDOW, using default", "error", err, "default", "1m")
		window = time.Minute
	}

	return &Aggregator{
		Input:  input,
		Repo:   repo,
		Cache:  cache,
		Window: window,
	}
}

func (a *Aggregator) Start(ctx context.Context) {
	buffer := make(map[string][]domain.PriceUpdate)
	ticker := time.NewTicker(a.Window)
	defer ticker.Stop()

	logger.Info("starting price aggregator", "window", a.Window)

	for {
		select {
		case <-ctx.Done():
			a.flush(ctx, buffer, time.Now())
			logger.Info("aggregator stopped by context")
			return

		case update, ok := <-a.Input:
			if !ok {
				a.flush(ctx, buffer, time.Now())
				logger.Info("aggregator channel closed, stopping")
				return
			}

			// Store immediately in cache
			if err := a.Cache.StoreLatestPrice(ctx, update); err != nil {
				logger.Error("failed to store in cache", "error", err)
			}

			// Add to buffer for aggregation
			key := update.Exchange + ":" + update.Pair
			buffer[key] = append(buffer[key], update)
			logger.Debug("price added to buffer", "exchange", update.Exchange, "pair", update.Pair, "price", update.Price)

		case tickTime := <-ticker.C:
			a.flush(ctx, buffer, tickTime)
			// Reset buffer but keep existing maps for next period
			for key := range buffer {
				buffer[key] = nil // Clear slice but keep map entry
			}
			logger.Info("flushed aggregation buffer", "time", tickTime)
		}
	}
}

func (a *Aggregator) flush(ctx context.Context, buffer map[string][]domain.PriceUpdate, ts time.Time) {
	if len(buffer) == 0 {
		logger.Debug("flush called with empty buffer")
		return
	}

	for key, updates := range buffer {
		if len(updates) == 0 {
			continue
		}

		parts := strings.Split(key, ":")
		if len(parts) < 2 {
			logger.Warn("invalid buffer key", "key", key)
			continue
		}

		exchange, pair := parts[0], parts[1]
		stats := a.aggregateUpdates(updates, exchange, pair, ts)

		// Store in database
		if err := a.Repo.StoreStatsBatch(ctx, []domain.PriceStats{stats}); err != nil {
			logger.Error("failed to store aggregated stats",
				"exchange", exchange, "pair", pair, "error", err)
		}

		logger.Info("stored aggregated stats",
			"exchange", exchange, "pair", pair,
			"avg", stats.Average, "min", stats.Min, "max", stats.Max)
	}
}

func (a *Aggregator) aggregateUpdates(updates []domain.PriceUpdate, exchange, pair string, ts time.Time) domain.PriceStats {
	if len(updates) == 0 {
		return domain.PriceStats{}
	}

	var sum float64
	min := updates[0].Price
	max := updates[0].Price

	for _, update := range updates {
		sum += update.Price
		if update.Price < min {
			min = update.Price
		}
		if update.Price > max {
			max = update.Price
		}
	}

	avg := sum / float64(len(updates))

	return domain.PriceStats{
		Exchange:  exchange,
		Pair:      pair,
		Timestamp: ts,
		Average:   avg,
		Min:       min,
		Max:       max,
	}
}
