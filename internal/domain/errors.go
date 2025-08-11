package domain

import "errors"

var (
	ErrInvalidExchangeVal             = errors.New("exchange value is invalid , must be (Exchange1, Exchange2, Exchange3, All)")
	ErrInvalidMetricVal               = errors.New("metric value is invalid , must be (highest, lowest, latest, average)")
	ErrInvalidSymbolVal               = errors.New("symbol value is invalid , must be (BTCUSDT, DOGEUSDT, TONUSDT, ETHUSDT, SOLUSDT)")
	ErrInvalidModeVal                 = errors.New("mode value is invalid, must be (test or live)")
	ErrAllNotSupported                = errors.New(`"All" is not supported for this period-based query`)
	ErrEmptyMetricVal                 = errors.New("metric value is empty")
	ErrEmptyExchangeVal               = errors.New("exchange value is empty")
	ErrEmptySymbolVal                 = errors.New("symbol value is empty")
	ErrHighPriceNotFound              = errors.New("highest price is not found")
	ErrHighPriceWithPeriodNotFound    = errors.New("highest price data is unavailable for the selected period")
	ErrLowestPriceNotFound            = errors.New("lowest price is not found")
	ErrLowestPriceWithPeriodNotFound  = errors.New("lowest price data is unavailable for the selected period")
	ErrLatestPriceNotFound            = errors.New("latest price is not found")
	ErrAveragePriceNotFound           = errors.New("average price is not found")
	ErrAveragePriceWithPeriodNotFound = errors.New("average price data is unavailable for the selected period")
)
