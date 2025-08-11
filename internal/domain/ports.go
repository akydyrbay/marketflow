package domain

import "time"

// For adapters
type DataFetcher interface {
	SetupDataFetcher() (chan map[string]ExchangeData, chan []Data, error)
	CheckHealth() error
	Close()
}

type CacheMemory interface {
	CheckHealth() error
	LatestData(exchange, symbol string) (Data, error)
	SaveAggregatedData(aggregatedData map[string]ExchangeData) error
	SaveLatestData(latestData map[string]Data) error
}

type Database interface {
	DatabaseSaver
	LatestDataReader
	AvgPriceReader
	MinPriceReader
	MaxPriceReader
	DatabaseHealthChecker
}

type DatabaseSaver interface {
	SaveAggregatedData(aggregatedData map[string]ExchangeData) error
	SaveLatestData(latestData map[string]Data) error
}

type LatestDataReader interface {
	LatestDataByExchange(exchange, symbol string) (Data, error)
	LatestDataByAllExchanges(symbol string) (Data, error)
}

type AvgPriceReader interface {
	AveragePriceByExchange(exchange, symbol string) (Data, error)
	AveragePriceByAllExchanges(symbol string) (Data, error)
	AveragePriceWithDuration(exchange, symbol string, startTime time.Time, duration time.Duration) (Data, error)
}

type MinPriceReader interface {
	MinPriceByAllExchanges(symbol string) (Data, error)
	MinPriceByExchange(exchange, symbol string) (Data, error)
	MinPriceByExchangeWithDuration(exchange, symbol string, startTime time.Time, duration time.Duration) (Data, error)
	MinPriceByAllExchangesWithDuration(symbol string, startTime time.Time, duration time.Duration) (Data, error)
}

type MaxPriceReader interface {
	MaxPriceByAllExchanges(symbol string) (Data, error)
	MaxPriceByExchange(exchange, symbol string) (Data, error)
	MaxPriceByExchangeWithDuration(exchange, symbol string, startTime time.Time, duration time.Duration) (Data, error)
	MaxPriceByAllExchangesWithDuration(symbol string, startTime time.Time, duration time.Duration) (Data, error)
}

type DatabaseHealthChecker interface {
	CheckHealth() error
}

// For services
type DataModeService interface {
	LatestDataSaver
	LatestDataGetter
	DataAggregator
	AvgPriceGetter
	HighestPriceGetter
	LowestPriceGetter
	DataManager
}

type LatestDataSaver interface {
	SaveLatestData(rawDataCh chan []Data)
}

type LatestDataGetter interface {
	LatestData(exchange string, symbol string) (Data, int, error)
}

type DataAggregator interface {
	AggregatedDataByDuration(exchange, symbol string, duration time.Duration) []map[string]ExchangeData
}

type AvgPriceGetter interface {
	AveragePrice(exchange, symbol string) (Data, int, error)
	AveragePriceWithPeriod(exchange, symbol, period string) (Data, int, error)
}

type HighestPriceGetter interface {
	HighestPrice(exchange, symbol string) (Data, int, error)
	HighestPriceWithPeriod(exchange, symbol string, period string) (Data, int, error)
	HighestPriceByAllExchangesWithPeriod(symbol string, period string) (Data, int, error)
}

type LowestPriceGetter interface {
	LowestPrice(exchange, symbol string) (Data, int, error)
	LowestPriceWithPeriod(exchange, symbol string, period string) (Data, int, error)
	LowestPriceByAllExchangesWithPeriod(symbol string, period string) (Data, int, error)
}

type DataManager interface {
	SwitchMode(mode string) (int, error)
	CheckHealth() []ConnMsg
	ListenAndSave() error
	StopListening()
}
