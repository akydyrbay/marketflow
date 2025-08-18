package server

import (
	"marketflow/internal/domain"
	"marketflow/internal/domain/utils"
	"marketflow/pkg/logger"
	"net/http"
)

// Latest data validation and service logic
func (serv *DataModeServiceImp) LatestData(exchange string, symbol string) (domain.Data, int, error) {
	var (
		latest domain.Data
		err    error
	)

	if err := utils.CheckExchangeName(exchange); err != nil {
		logger.Error("Failed to get latest data: ", "error", err.Error())
		return latest, http.StatusBadRequest, err
	}

	if err := utils.CheckSymbolName(symbol); err != nil {
		logger.Error("Failed to get latest data: ", "error", err.Error())
		return latest, http.StatusBadRequest, err
	}

	// first we look for data in the cache
	latest, err = serv.Cache.LatestData(exchange, symbol)
	if err != nil {
		// If Redis is not available, se look for data in the DB
		logger.Debug("Failed to get latest data from cache: ", "error", err.Error())
		if exchange == "All" {
			latest, err = serv.DB.LatestDataByAllExchanges(symbol)
			if err != nil {
				logger.Error("Failed to get latest data by all exchanges from Db: ", "error", err.Error())
				return latest, http.StatusInternalServerError, err
			}
		} else {
			latest, err = serv.DB.LatestDataByExchange(exchange, symbol)
			if err != nil {
				logger.Error("Failed to get latest data by exchange from Db: ", "error", err.Error())
				return latest, http.StatusInternalServerError, err
			}
		}
	}

	if latest.Price == 0 {
		return domain.Data{}, http.StatusNotFound, domain.ErrLatestPriceNotFound
	}

	return latest, http.StatusOK, nil
}
