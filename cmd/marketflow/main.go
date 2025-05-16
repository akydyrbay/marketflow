package main

import (
	"context"
	"fmt"
	"time"

	"marketflow/internal/adapters/exchange"
	m "marketflow/internal/domain"
)

func main() {
	ctx := context.Background()
	updates := make(chan m.PriceUpdate, 100)

	adapter := exchange.ExchangeAdapter{
		Addr:        "localhost:40101",
		Outbound:    updates,
		BackoffBase: 500 * time.Millisecond,
	}

	go adapter.Run(ctx)

	for update := range updates {
		fmt.Printf("[DATA] %s → %.2f @ %d\n", update.Symbol, update.Price, update.Timestamp)
	}
}
