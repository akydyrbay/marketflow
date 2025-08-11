package exchange

import (
	"math/rand"
	"time"

	"marketflow/internal/domain"
	"marketflow/internal/service"
)

type TestMode struct {
	stop chan struct{}
}

func NewTestModeFetcher() *TestMode {
	return &TestMode{stop: make(chan struct{})}
}

func (m *TestMode) SetupDataFetcher() (chan map[string]domain.ExchangeData, chan []domain.Data, error) {
	rawFlow := make(chan []domain.Data, 100)

	exchanges := []string{"Exchange1", "Exchange2", "Exchange3"}
	basePrices := map[string]float64{
		"BTCUSDT": 60000.0, "DOGEUSDT": 0.15, "TONUSDT": 5.0, "SOLUSDT": 150.0, "ETHUSDT": 3000.0,
	}

	go func() {
		ticker := time.NewTicker(1000 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-m.stop:
				close(rawFlow)
				return
			case <-ticker.C:
				var rawData []domain.Data
				now := time.Now()

				for i := 0; i < len(exchanges); i++ {
					ex := exchanges[rand.Intn(len(exchanges))]
					for _, pair := range domain.Symbols {
						// Generate random price fluctuation (±15%)
						price := basePrices[pair] * (1 + (rand.Float64()-0.5)*0.3)
						rawData = append(rawData, domain.Data{
							ExchangeName: ex,
							Symbol:       pair,
							Price:        price,
							Timestamp:    now.UnixNano() / int64(time.Millisecond),
						})
					}
				}

				rawFlow <- rawData

			}
		}
	}()

	aggregatedCh, rawCh := service.Aggregate(rawFlow)
	return aggregatedCh, rawCh, nil
}

func AggregateFromTestMode(input chan []domain.Data) (chan map[string]domain.ExchangeData, chan []domain.Data) {
	aggregated := make(chan map[string]domain.ExchangeData, 100)
	raw := make(chan []domain.Data, 100)

	go func() {
		for data := range input {
			// просто пробрасываем
			raw <- data

			agg := make(map[string]domain.ExchangeData)
			now := time.Now()

			for _, d := range data {
				key := d.ExchangeName + " " + d.Symbol
				agg[key] = domain.ExchangeData{
					Pair_name:     d.Symbol,
					Exchange:      d.ExchangeName,
					Timestamp:     now,
					Average_price: d.Price,
					Min_price:     d.Price,
					Max_price:     d.Price,
				}
			}
			aggregated <- agg
		}
	}()

	return aggregated, raw
}

func (g *TestMode) Close() {
	close(g.stop)
}
func (m *TestMode) CheckHealth() error {
	return nil
}
