package server

import (
	"context"
	"fmt"
	"log/slog"
	"marketflow/internal/adapters/exchange"
	"marketflow/internal/domain"
	"net/http"
	"sync"
	"time"
)

type DataModeServiceImp struct {
	DataBuffer  []map[string]domain.ExchangeData
	Datafetcher domain.DataFetcher
	Cache       domain.CacheMemory
	cancel      context.CancelFunc
	DB          domain.Database
	wg          sync.WaitGroup
	mu          sync.Mutex
}

func NewDataFetcher(dataSource domain.DataFetcher, DataSaver domain.Database, Cache domain.CacheMemory) *DataModeServiceImp {
	return &DataModeServiceImp{
		Datafetcher: dataSource,
		DB:          DataSaver,
		Cache:       Cache,
		DataBuffer:  make([]map[string]domain.ExchangeData, 0),
	}
}

var _ (domain.DataModeService) = (*DataModeServiceImp)(nil)

// Mode switch core logic
func (serv *DataModeServiceImp) SwitchMode(mode string) (int, error) {
	serv.mu.Lock()
	defer serv.mu.Unlock()

	// Check if is current datafetcher mode equal to changing mode
	if _, ok := serv.Datafetcher.(*exchange.LiveMode); (ok && mode == "live") || (!ok && mode == "test") {
		return http.StatusBadRequest, fmt.Errorf("data mode is already switched to %s", mode)
	}

	switch mode {
	case "test":
		serv.Datafetcher.Close()
		serv.Datafetcher = exchange.NewTestModeFetcher()
		if err := serv.ListenAndSave(); err != nil {
			return http.StatusInternalServerError, err
		}
	case "live":
		serv.Datafetcher.Close()
		serv.Datafetcher = exchange.NewLiveModeFetcher()
		if err := serv.ListenAndSave(); err != nil {
			return http.StatusInternalServerError, err
		}
	default:
		return http.StatusBadRequest, domain.ErrInvalidModeVal
	}
	return http.StatusOK, nil
}

// Goroutines stop logic
func (serv *DataModeServiceImp) StopListening() {
	serv.cancel()
	serv.Datafetcher.Close()
	serv.wg.Wait()
	slog.Info("Listen and save goroutine has been finished...")
}

// Core logic: handle data retrieval, aggregation, and persistence for exchanges
func (serv *DataModeServiceImp) ListenAndSave() error {
	ctx, cancel := context.WithCancel(context.Background())
	serv.cancel = cancel

	aggregated, rawDataCh, err := serv.Datafetcher.SetupDataFetcher()
	if err != nil {
		return err
	}
	serv.wg.Add(3)

	go serv.listenAndSaveLatest(rawDataCh)
	go serv.aggregateAndSaveEveryMinute(ctx)
	go serv.collectAggregatedData(ctx, aggregated)

	return nil
}

// Stores the latest data into DB and Cache from raw data channel
func (serv *DataModeServiceImp) listenAndSaveLatest(rawDataCh chan []domain.Data) {
	defer serv.wg.Done()
	serv.SaveLatestData(rawDataCh)
}

// Every minute aggregates buffered data and saves it
func (serv *DataModeServiceImp) aggregateAndSaveEveryMinute(ctx context.Context) {
	defer serv.wg.Done()
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			serv.mu.Lock()
			merged := MergeAggregatedData(serv.DataBuffer)
			serv.DB.SaveAggregatedData(merged)
			serv.Cache.SaveAggregatedData(merged)
			serv.DataBuffer = nil
			serv.mu.Unlock()
		}
	}
}

// Collects aggregated data into buffer until context is cancelled
func (serv *DataModeServiceImp) collectAggregatedData(ctx context.Context, aggregated chan map[string]domain.ExchangeData) {
	defer serv.wg.Done()

	for {
		select {
		case <-ctx.Done():
			for data := range aggregated {
				serv.mu.Lock()
				serv.DataBuffer = append(serv.DataBuffer, data)
				slog.Debug("Received data", "buffer_size", len(serv.DataBuffer))
				serv.mu.Unlock()
			}
			return
		case data, ok := <-aggregated:
			if !ok {
				return
			}
			serv.mu.Lock()
			serv.DataBuffer = append(serv.DataBuffer, data)
			serv.mu.Unlock()
		}
	}
}

// Retrieves the latest data from the channel and stores it in both PostgreSQL and Redis
func (serv *DataModeServiceImp) SaveLatestData(rawDataCh chan []domain.Data) {
	for rawData := range rawDataCh {
		latestData := make(map[string]domain.Data)
		for i := len(rawData) - 1; i >= 0; i-- {
			if rawData[i].ExchangeName == "" || rawData[i].Symbol == "" {
				continue
			}

			exchKey := "latest " + rawData[i].ExchangeName + " " + rawData[i].Symbol
			allKey := "latest " + "All" + " " + rawData[i].Symbol

			if _, exist := latestData[exchKey]; !exist {
				latestData[exchKey] = rawData[i]
			}

			if _, exist := latestData[allKey]; !exist {
				latestData[allKey] = rawData[i]
			}

			maxLatest := len(domain.Exchanges) * len(domain.Symbols)

			// Break loop if we find all latest prices
			if len(latestData) == maxLatest {
				break
			}
		}

		if err := serv.Cache.SaveLatestData(latestData); err != nil {
			slog.Debug("Failed to save latest data to cache: " + err.Error())

			if err := serv.DB.SaveLatestData(latestData); err != nil {
				slog.Error("Failed to save latest data to Db: " + err.Error())
			}
		}

	}
}

// Merges multiple aggregated exchange data entries into a single aggregated result
func MergeAggregatedData(DataBuffer []map[string]domain.ExchangeData) map[string]domain.ExchangeData {
	result := make(map[string]domain.ExchangeData)
	sums := make(map[string]float64)
	counts := make(map[string]int)

	for _, dataMap := range DataBuffer {
		for key, val := range dataMap {
			agg, exists := result[key]
			if !exists {
				agg = domain.ExchangeData{
					Pair_name: val.Pair_name,
					Exchange:  val.Exchange,
					Min_price: val.Min_price,
					Max_price: val.Max_price,
					Timestamp: val.Timestamp,
				}
			}

			if val.Min_price < agg.Min_price {
				agg.Min_price = val.Min_price
			}
			if val.Max_price > agg.Max_price {
				agg.Max_price = val.Max_price
			}

			sums[key] += val.Average_price
			counts[key]++

			if val.Timestamp.After(agg.Timestamp) {
				agg.Timestamp = val.Timestamp
			}

			result[key] = agg
		}
	}

	// Count average
	for key, item := range result {
		if count := counts[key]; count > 0 {
			item.Average_price = sums[key] / float64(count)
			result[key] = item
		}
	}
	return result
}

// Fetches aggregated market data for a specific exchange and symbol within a time period
func (serv *DataModeServiceImp) AggregatedDataByDuration(exchange, symbol string, duration time.Duration) []map[string]domain.ExchangeData {
	serv.mu.Lock()
	defer serv.mu.Unlock()

	cutoff := time.Now().Add(-duration - 10*time.Second)

	var latest []map[string]domain.ExchangeData
	var lastSeen *domain.ExchangeData

	for i := len(serv.DataBuffer) - 1; i >= 0; i-- {
		m := serv.DataBuffer[i]
		data, ok := m[exchange+" "+symbol]
		if ok {
			lastSeen = &data
			if !data.Timestamp.Before(cutoff) {
				latest = append(latest, m)
			}
		}
	}
	if len(latest) == 0 && lastSeen != nil {
		fmt.Println("DEBUG: nothing matched cutoff =", cutoff)
		fmt.Println("DEBUG: buffer length =", len(serv.DataBuffer))
		for i := len(serv.DataBuffer) - 1; i >= 0; i-- {
			m := serv.DataBuffer[i]
			if d, ok := m[exchange+" "+symbol]; ok {
				fmt.Println("BUFFER ENTRY:", d.Exchange, d.Pair_name, d.Timestamp)
			}
		}
	}

	return latest
}
