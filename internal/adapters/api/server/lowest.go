package server

import (
	"marketflow/internal/domain"
	"marketflow/internal/domain/utils"
	"marketflow/pkg/logger"
	"net/http"
	"time"
)

// Fetches the lowest price by specific exchange and given symbol
func (serv *DataModeServiceImp) LowestPrice(exchange, symbol string) (domain.Data, int, error) {
	if err := utils.CheckExchangeName(exchange); err != nil {
		return domain.Data{}, http.StatusBadRequest, err
	}

	if err := utils.CheckSymbolName(symbol); err != nil {
		return domain.Data{}, http.StatusBadRequest, err
	}

	var (
		lowest domain.Data
		err    error
	)
	switch exchange {
	case "All":
		lowest, err = serv.DB.MinPriceByAllExchanges(symbol)
		if err != nil {
			logger.Error("Failed to get lowest price by all exchanges", "error", err.Error())
			return domain.Data{}, http.StatusInternalServerError, err
		}
	default:
		lowest, err = serv.DB.MinPriceByExchange(exchange, symbol)
		if err != nil {
			logger.Error("Failed to get lowest price from exchange", "error", err.Error())
			return domain.Data{}, http.StatusInternalServerError, err
		}
	}

	serv.mu.Lock()
	merged := MergeAggregatedData(serv.DataBuffer)
	serv.mu.Unlock()

	key := exchange + " " + symbol
	if agg, ok := merged[key]; ok {
		if lowest.Price == 0 || lowest.Price > agg.Min_price {
			lowest.Price = agg.Min_price
			lowest.Timestamp = agg.Timestamp.UnixMilli()
		}
	} else {
		logger.Warn("Aggregated data not found for key", "key", key)
	}

	if lowest.Price == 0 {
		return domain.Data{}, http.StatusNotFound, domain.ErrLowestPriceNotFound
	}

	return lowest, http.StatusOK, nil
}

// Fetches the lowest price by specific exchange and symbol over a specified period
func (serv *DataModeServiceImp) LowestPriceWithPeriod(exchange, symbol string, period string) (domain.Data, int, error) {
	if err := utils.CheckExchangeName(exchange); err != nil {
		return domain.Data{}, http.StatusBadRequest, err
	}

	if err := utils.CheckSymbolName(symbol); err != nil {
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

	lowest, err := serv.DB.MinPriceByExchangeWithDuration(exchange, symbol, startTime, duration)
	if err != nil {
		logger.Error("Failed to get lowest price from Exchange by period", "error", err.Error())
		return domain.Data{}, http.StatusInternalServerError, err
	}

	aggregated := serv.AggregatedDataByDuration(exchange, symbol, duration)
	merged := MergeAggregatedData(aggregated)

	key := exchange + " " + symbol
	if agg, ok := merged[key]; ok {
		if lowest.Price == 0 || lowest.Price > agg.Min_price {
			lowest.Price = agg.Min_price
			lowest.Timestamp = agg.Timestamp.UnixMilli()
		}
	} else {
		logger.Warn("Aggregated data not found for key", "key", key)
	}

	if lowest.Price == 0 {
		return domain.Data{}, http.StatusNotFound, domain.ErrLowestPriceWithPeriodNotFound
	}

	return lowest, http.StatusOK, nil
}

// Fetches the lowest price across all exchanges for a given symbol over a specified period
func (serv *DataModeServiceImp) LowestPriceByAllExchangesWithPeriod(symbol string, period string) (domain.Data, int, error) {
	exchange := "All"
	if err := utils.CheckSymbolName(symbol); err != nil {
		return domain.Data{}, http.StatusBadRequest, err
	}

	duration, err := time.ParseDuration(period)
	if err != nil {
		return domain.Data{}, http.StatusBadRequest, err
	}

	startTime := time.Now()

	lowest, err := serv.DB.MinPriceByAllExchangesWithDuration(symbol, startTime, duration)
	if err != nil {
		logger.Error("Failed to get lowest price from Exchange by period", "error", err.Error())
		return domain.Data{}, http.StatusInternalServerError, err
	}

	aggregated := serv.AggregatedDataByDuration(exchange, symbol, duration)
	merged := MergeAggregatedData(aggregated)

	key := exchange + " " + symbol
	if agg, ok := merged[key]; ok {
		if lowest.Price == 0 || lowest.Price > agg.Min_price {
			lowest.Price = agg.Min_price
			lowest.Timestamp = agg.Timestamp.UnixMilli()
		}
	} else {
		logger.Warn("Aggregated data not found for key", "key", key)
	}

	if lowest.Price == 0 {
		return domain.Data{}, http.StatusNotFound, domain.ErrLowestPriceWithPeriodNotFound
	}

	return lowest, http.StatusOK, nil
}
