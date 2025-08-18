package service

import (
	"marketflow/internal/domain"
	"math"
	"strings"
	"time"
)

func Aggregate(mergedCh chan []domain.Data) (chan map[string]domain.ExchangeData, chan []domain.Data) {
	aggregatedCh := make(chan map[string]domain.ExchangeData)
	rawDataCh := make(chan []domain.Data)

	go func() {
		for dataBatch := range mergedCh {

			// To prevent the main thread from being delayed
			go func() {
				rawDataCh <- dataBatch
			}()

			exchangesData := make(map[string]domain.ExchangeData)
			counts := make(map[string]int)
			sums := make(map[string]float64)

			for _, data := range dataBatch {
				keys := []string{
					data.ExchangeName + " " + data.Symbol, // by exchange
					"All " + data.Symbol,                  // by all exchanges
				}

				for _, key := range keys {
					val, exists := exchangesData[key]
					if !exists {
						val = domain.ExchangeData{
							Exchange:  strings.Split(key, " ")[0],
							Pair_name: data.Symbol,
							Min_price: math.Inf(1),
							Max_price: math.Inf(-1),
						}
					}

					// обновление мин/макс
					if data.Price < val.Min_price {
						val.Min_price = data.Price
					}
					if data.Price > val.Max_price {
						val.Max_price = data.Price
					}

					sums[key] += data.Price
					counts[key]++

					exchangesData[key] = val
				}
			}

			// Counting avg price
			for key, ed := range exchangesData {
				if count, ok := counts[key]; ok && count > 0 {
					ed.Average_price = sums[key] / float64(count)
					ed.Timestamp = time.Now()
					exchangesData[key] = ed
				}
			}
			aggregatedCh <- exchangesData
		}
		close(aggregatedCh)
		close(rawDataCh)
	}()

	return aggregatedCh, rawDataCh
}
