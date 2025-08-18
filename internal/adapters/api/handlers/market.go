package handlers

import (
	"fmt"
	"marketflow/internal/domain"
	"marketflow/internal/domain/utils"
	"marketflow/pkg/logger"
	"net/http"
)

type MarketDataHTTPHandler struct {
	serv domain.DataModeService
}

func NewMarketDataHandler(serv domain.DataModeService) *MarketDataHTTPHandler {
	return &MarketDataHTTPHandler{serv: serv}
}

const (
	MetricHighest = "highest"
	MetricLowest  = "lowest"
	MetricAverage = "average"
	MetricLatest  = "latest"
)

// Core handler for processing metric-based queries by specific exchange
func (h *MarketDataHTTPHandler) ProcessMetricQueryByExchange(w http.ResponseWriter, r *http.Request) {
	var (
		data domain.Data
		msg  string
		code int = 200
		err  error
	)

	metric := r.PathValue("metric")
	if len(metric) == 0 {
		logger.Error("Failed to get metric value from path: ", "error", domain.ErrEmptyMetricVal.Error())
		utils.SendMsg(w, http.StatusBadRequest, domain.ErrEmptyMetricVal.Error())
		return
	}

	exchange := r.PathValue("exchange")
	if len(exchange) == 0 {
		logger.Error("Failed to get exchange value from path: ", "error", domain.ErrEmptyExchangeVal.Error())
		utils.SendMsg(w, http.StatusBadRequest, domain.ErrEmptyExchangeVal.Error())
		return
	}

	symbol := r.PathValue("symbol")
	if len(symbol) == 0 {
		logger.Error("Failed to get symbol value from path: ", "error", domain.ErrEmptySymbolVal)
		utils.SendMsg(w, http.StatusBadRequest, domain.ErrEmptySymbolVal.Error())
		return
	}

	switch metric {
	case MetricHighest:
		period := r.URL.Query().Get("period")
		if period == "" {
			data, code, err = h.serv.HighestPrice(exchange, symbol)
			if err != nil {
				logger.Error("Failed to get highest price: ", "exchange", exchange, "symbol", symbol, "error", err.Error())
				utils.SendMsg(w, code, err.Error())
				return
			}

		} else {
			data, code, err = h.serv.HighestPriceWithPeriod(exchange, symbol, period)
			if err != nil {
				logger.Error("Failed to get highest price: ", "exchange", exchange, "symbol", symbol, "error", err.Error())
				utils.SendMsg(w, code, err.Error())
				return
			}
		}
		msg = fmt.Sprintf("Highest price for %s at %s duration {%s}: %.2f", symbol, exchange, period, data.Price)
	case MetricLowest:
		period := r.URL.Query().Get("period")

		if period == "" {
			data, code, err = h.serv.LowestPrice(exchange, symbol)
			if err != nil {
				logger.Error("Failed to get lowest price: ", "exchange", exchange, "symbol", symbol, "error", err.Error())
				utils.SendMsg(w, code, err.Error())
				return
			}
		} else {
			data, code, err = h.serv.LowestPriceWithPeriod(exchange, symbol, period)
			if err != nil {
				logger.Error("Failed to get lowest price: ", "exchange", exchange, "symbol", symbol, "error", err.Error())
				utils.SendMsg(w, code, err.Error())
				return
			}
		}
		msg = fmt.Sprintf("Lowest price for %s at %s duration {%s}: %.2f", symbol, exchange, period, data.Price)
	case MetricAverage:
		period := r.URL.Query().Get("period")
		if period == "" {
			data, code, err = h.serv.AveragePrice(exchange, symbol)
			if err != nil {
				logger.Error("Failed to get average price: ", "exchange", exchange, "symbol", symbol, "error", err.Error())
				utils.SendMsg(w, code, err.Error())
				return
			}

		} else {
			data, code, err = h.serv.AveragePriceWithPeriod(exchange, symbol, period)
			if err != nil {
				logger.Error("Failed to get average price with period: ", "exchange", exchange, "symbol", symbol, "period", period, "error", err.Error())
				utils.SendMsg(w, code, err.Error())
				return
			}

		}
		msg = fmt.Sprintf("Average price for %s at %s duration {%s}: %.2f", symbol, exchange, period, data.Price)
	case MetricLatest:
		data, code, err = h.serv.LatestData(exchange, symbol)
		if err != nil {
			logger.Error("Failed to get latest data: ", "exchange", exchange, "symbol", symbol, "error", err.Error())
			utils.SendMsg(w, code, err.Error())
			return
		}
		msg = fmt.Sprintf("Latest price for %s at %s: %.2f", symbol, exchange, data.Price)

	default:
		logger.Error("Failed to get data by metric: ", "exchange", "All", "symbol", symbol, "metric", metric, "error", domain.ErrInvalidMetricVal.Error())
		utils.SendMsg(w, http.StatusBadRequest, domain.ErrInvalidMetricVal.Error())
		return
	}

	if err := utils.SendMetricData(w, code, data); err != nil {
		logger.Error("Failed to send JSON message: ", "data", data, "error", err.Error())
		utils.SendMsg(w, code, err.Error())
		return
	}

	logger.Info(msg)
}

// Core handler for processing metric-based queries across all exchanges
func (h *MarketDataHTTPHandler) ProcessMetricQueryByAll(w http.ResponseWriter, r *http.Request) {
	var (
		data     domain.Data
		exchange = "All"
		msg      string
		code     int = 200
		err      error
	)
	metric := r.PathValue("metric")
	if len(metric) == 0 {
		logger.Error("Failed to get metric value from path: ", "error", domain.ErrEmptyMetricVal.Error())
		utils.SendMsg(w, http.StatusBadRequest, domain.ErrEmptyMetricVal.Error())
		return
	}

	symbol := r.PathValue("symbol")
	if len(symbol) == 0 {
		logger.Error("Failed to get symbol value from path: ", "error", domain.ErrEmptyExchangeVal)
		utils.SendMsg(w, http.StatusBadRequest, domain.ErrEmptySymbolVal.Error())
		return
	}

	switch metric {
	case MetricHighest:
		period := r.URL.Query().Get("period")
		if period == "" {
			data, code, err = h.serv.HighestPrice(exchange, symbol)
			if err != nil {
				logger.Error("Failed to get highest price: ", "exchange", exchange, "symbol", symbol, "error", err.Error())
				utils.SendMsg(w, code, err.Error())
				return
			}
		} else {
			data, code, err = h.serv.HighestPriceByAllExchangesWithPeriod(symbol, period)
			if err != nil {
				logger.Error("Failed to get highest price with period: ", "exchange", exchange, "symbol", symbol, "period", period, "error", err.Error())
				utils.SendMsg(w, code, err.Error())
				return
			}
		}
		msg = fmt.Sprintf("Highest price for %s at %s: %.2f", symbol, exchange, data.Price)
	case MetricLowest:
		period := r.URL.Query().Get("period")
		if period == "" {
			data, code, err = h.serv.LowestPrice(exchange, symbol)
			if err != nil {
				logger.Error("Failed to get lowest price: ", "exchange", exchange, "symbol", symbol, "error", err.Error())
				utils.SendMsg(w, code, err.Error())
				return
			}
		} else {
			data, code, err = h.serv.LowestPriceByAllExchangesWithPeriod(symbol, period)
			if err != nil {
				logger.Error("Failed to get lowest price: ", "exchange", exchange, "symbol", symbol, "error", err.Error())
				utils.SendMsg(w, code, err.Error())
				return
			}
		}

		msg = fmt.Sprintf("Lowest price for %s at %s: %.2f", symbol, exchange, data.Price)
	case MetricAverage:
		data, code, err = h.serv.AveragePrice(exchange, symbol)
		if err != nil {
			logger.Error("Failed to get average price: ", "exchange", exchange, "symbol", symbol, "error", err.Error())
			utils.SendMsg(w, code, err.Error())
			return
		}

		msg = fmt.Sprintf("Average price for %s at %s: %.2f", symbol, exchange, data.Price)
	case MetricLatest:
		data, code, err = h.serv.LatestData(exchange, symbol)
		if err != nil {
			logger.Error("Failed to get latest data: ", "exchange", exchange, "symbol", symbol, "error", err.Error())
			utils.SendMsg(w, code, err.Error())
			return
		}

		msg = fmt.Sprintf("Latest price for %s at %s: %.2f", symbol, exchange, data.Price)
	default:
		logger.Error("Failed to get data by metric: ", "exchange", exchange, "symbol", symbol, "metric", metric, "error", domain.ErrInvalidMetricVal.Error())
		utils.SendMsg(w, http.StatusBadRequest, domain.ErrInvalidMetricVal.Error())
		return
	}

	if err := utils.SendMetricData(w, code, data); err != nil {
		logger.Error("Failed to send JSON message: ", "data", data, "error", err.Error())
		utils.SendMsg(w, code, err.Error())
		return
	}
	logger.Info(msg)
}
