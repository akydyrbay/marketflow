package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"marketflow/internal/domain"
	"marketflow/pkg/logger"
)

func (s *Server) handleLatestPrice(input chan<- domain.PriceUpdate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 {
			http.Error(w, "invalid URL", http.StatusBadRequest)
			return
		}
		symbol := parts[3]
		exchange := r.URL.Query().Get("exchange")
		if exchange == "" {
			exchange = "ex1" // Default exchange
		}

		update, err := s.cache.GetLatest(ctx, exchange, symbol)
		if err != nil {
			logger.Warn("cache miss, falling back to postgres", "symbol", symbol, "exchange", exchange)
			stats, err := s.repo.GetLatest(ctx, exchange, symbol)
			if err != nil {
				logger.Error("failed to get latest price", "symbol", symbol, "exchange", exchange, "error", err)
				http.Error(w, "failed to get latest price", http.StatusInternalServerError)
				return
			}
			update = domain.PriceUpdate{
				Exchange: stats.Exchange,
				Pair:     stats.Pair,
				Price:    stats.Average,
				Time:     stats.Timestamp,
			}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"exchange": update.Exchange,
			"pair":     update.Pair,
			"price":    update.Price,
			"time":     update.Time,
		})
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("failed to encode response", "error", err)
	}
}
