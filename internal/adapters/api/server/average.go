package server

import (
	"marketflow/internal/domain"
	"marketflow/internal/domain/utils"
	"marketflow/pkg/logger"
	"net/http"
	"time"
)

// Fetches the average price for a specific exchange and symbol
func (serv *DataModeServiceImp) AveragePrice(exchange, symbol string) (domain.Data, int, error) {
	var (
		data domain.Data
		err  error
	)

	if err := utils.CheckExchangeName(exchange); err != nil {
		return data, http.StatusBadRequest, err
	}

	if err := utils.CheckSymbolName(symbol); err != nil {
		return data, http.StatusBadRequest, err
	}

	switch exchange {
	case "All":
		data, err = serv.DB.AveragePriceByAllExchanges(symbol)
		if err != nil {
			return data, http.StatusInternalServerError, err
		}
	default:
		data, err = serv.DB.AveragePriceByExchange(exchange, symbol)
		if err != nil {
			return data, http.StatusInternalServerError, err
		}
	}

	// we also search it in the DataBuffer
	serv.mu.Lock()
	merged := MergeAggregatedData(serv.DataBuffer)
	serv.mu.Unlock()

	data.Timestamp = time.Now().UnixMilli()
	key := exchange + " " + symbol
	if avg, ok := merged[key]; ok {
		if avg.Average_price != 0 {
			data.Price = (avg.Average_price + data.Price) / 2
		}
	} else {
		logger.Warn("Aggregated data not found for key", "key", key)
	}

	if data.Price == 0 {
		return domain.Data{}, http.StatusNotFound, domain.ErrAveragePriceNotFound
	}

	return data, http.StatusOK, nil
}

// Fetches the average price for a specific exchange and symbol over a given period
func (serv *DataModeServiceImp) AveragePriceWithPeriod(exchange, symbol, period string) (domain.Data, int, error) {
	var (
		data domain.Data
		err  error
	)

	if err := utils.CheckExchangeName(exchange); err != nil {
		return data, http.StatusBadRequest, err
	}

	if err := utils.CheckSymbolName(symbol); err != nil {
		return data, http.StatusBadRequest, err
	}

	if exchange == "All" {
		return data, http.StatusBadRequest, domain.ErrAllNotSupported
	}

	duration, err := time.ParseDuration(period)
	if err != nil {
		return data, http.StatusBadRequest, err
	}
	startTime := time.Now()

	data, err = serv.DB.AveragePriceWithDuration(exchange, symbol, startTime, duration)
	if err != nil {
		return data, http.StatusInternalServerError, err
	}

	data.Timestamp = startTime.Add(-duration).UnixMilli()

	aggregated := serv.AggregatedDataByDuration(exchange, symbol, duration)
	merged := MergeAggregatedData(aggregated)

	key := exchange + " " + symbol
	if agg, ok := merged[key]; ok {
		if agg.Average_price != 0 {
			data.Price = (agg.Average_price + data.Price) / 2
		}
	} else {
		logger.Warn("Aggregated data not found for key", "key", key)
	}

	if data.Price == 0 {
		return domain.Data{}, http.StatusNotFound, domain.ErrAveragePriceWithPeriodNotFound
	}

	return data, http.StatusOK, nil
}
