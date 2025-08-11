package server

import (
	"log/slog"
	"marketflow/internal/domain"
	"net/http"
	"time"
)

// Fetches the highest price for a specific exchange and symbol
func (serv *DataModeServiceImp) HighestPrice(exchange, symbol string) (domain.Data, int, error) {
	var (
		highest domain.Data
		err     error
	)

	if err := domain.CheckExchangeName(exchange); err != nil {
		return domain.Data{}, http.StatusBadRequest, err
	}

	if err := domain.CheckSymbolName(symbol); err != nil {
		return domain.Data{}, http.StatusBadRequest, err
	}

	switch exchange {
	case "All":
		highest, err = serv.DB.MaxPriceByAllExchanges(symbol)
		if err != nil {
			slog.Error("Failed to get highest price by all exchanges", "error", err.Error())
			return domain.Data{}, http.StatusInternalServerError, err
		}

	default:
		highest, err = serv.DB.MaxPriceByExchange(exchange, symbol)
		if err != nil {
			slog.Error("Failed to get highest price from exchange", "error", err.Error())
			return domain.Data{}, http.StatusInternalServerError, err
		}
	}

	serv.mu.Lock()
	merged := MergeAggregatedData(serv.DataBuffer)
	serv.mu.Unlock()

	key := exchange + " " + symbol
	if agg, ok := merged[key]; ok {
		if agg.Max_price > highest.Price {
			highest.Price = agg.Max_price
			highest.Timestamp = agg.Timestamp.UnixMilli()
		}
	} else {
		slog.Warn("Aggregated data not found for key", "key", key)
	}

	if highest.Price == 0 {
		return domain.Data{}, http.StatusNotFound, domain.ErrHighPriceNotFound
	}

	return highest, http.StatusOK, nil
}

// Fetches the average price for a specific exchange and symbol over a given period
func (serv *DataModeServiceImp) HighestPriceWithPeriod(exchange, symbol string, period string) (domain.Data, int, error) {
	if err := domain.CheckExchangeName(exchange); err != nil {
		return domain.Data{}, http.StatusBadRequest, err
	}

	if err := domain.CheckSymbolName(symbol); err != nil {
		return domain.Data{}, http.StatusBadRequest, err
	}

	if exchange == "All" {
		return domain.Data{}, http.StatusBadRequest, domain.ErrAllNotSupported
	}

	duration, err := time.ParseDuration(period)
	if err != nil {
		return domain.Data{}, http.StatusBadRequest, err
	}

	startTime := time.Now()

	highest, err := serv.DB.MaxPriceByExchangeWithDuration(exchange, symbol, startTime, duration)
	if err != nil {
		slog.Error("Failed to get highest price from Exchange by period", "error", err.Error())
		return domain.Data{}, http.StatusInternalServerError, err
	}

	aggregated := serv.AggregatedDataByDuration(exchange, symbol, duration)
	merged := MergeAggregatedData(aggregated)

	key := exchange + " " + symbol
	if agg, ok := merged[key]; ok {
		if agg.Max_price > highest.Price {
			highest.Price = agg.Max_price
			highest.Timestamp = agg.Timestamp.UnixMilli()
		}
	} else {
		slog.Warn("Aggregated data not found for key", "key", key)
	}
	highest.Timestamp = startTime.Add(-duration).UnixMilli()

	if highest.Price == 0 {
		return domain.Data{}, http.StatusNotFound, domain.ErrHighPriceWithPeriodNotFound
	}

	return highest, http.StatusOK, nil
}

// Fetches the average price across all exchanges for a given symbol over a specified period
func (serv *DataModeServiceImp) HighestPriceByAllExchangesWithPeriod(symbol string, period string) (domain.Data, int, error) {
	exchange := "All"
	if err := domain.CheckSymbolName(symbol); err != nil {
		return domain.Data{}, http.StatusBadRequest, err
	}

	duration, err := time.ParseDuration(period)
	if err != nil {
		return domain.Data{}, http.StatusBadRequest, err
	}

	startTime := time.Now()

	highest, err := serv.DB.MaxPriceByAllExchangesWithDuration(symbol, startTime, duration)
	if err != nil {
		slog.Error("Failed to get highest price from Exchange by period", "error", err.Error())
		return domain.Data{}, http.StatusInternalServerError, err
	}

	aggregated := serv.AggregatedDataByDuration(exchange, symbol, duration)
	merged := MergeAggregatedData(aggregated)

	key := exchange + " " + symbol
	if agg, ok := merged[key]; ok {
		if agg.Max_price > highest.Price {
			highest.Price = agg.Max_price
			highest.Timestamp = agg.Timestamp.UnixMilli()
		}
	} else {
		slog.Warn("Aggregated data not found for key", "key", key)
	}

	if highest.Price == 0 {
		return domain.Data{}, http.StatusNotFound, domain.ErrHighPriceWithPeriodNotFound
	}

	return highest, http.StatusOK, nil
}
