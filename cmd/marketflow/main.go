package main

// func main() {
// 	ctx := context.Background()
// 	updates := make(chan m.PriceUpdate, 100)

// 	adapter := exchange.ExchangeAdapter{
// 		Addr:        "localhost:40101",
// 		Outbound:    updates,
// 		BackoffBase: 500 * time.Millisecond,
// 	}

// 	go adapter.Run(ctx)

// 	for update := range updates {
// 		fmt.Printf("[DATA] %s → %.2f @ %d\n", update.Symbol, update.Price, update.Timestamp)
// 	}
// }

// func main() {
// 	cache := &c.RedisCache{Addr: "localhost:6379"}
// 	if err := cache.Connect(); err != nil {
// 		panic(err)
// 	}
// 	now := time.Now().Unix()
// 	cache.AddPrice("exchange1", "BTCUSDT", 25000.5, now)
// 	cache.AddPrice("exchange1", "BTCUSDT", 25010.7, now+1)
// 	cache.Cleanup("exchange1", "BTCUSDT", now-60)
// 	// Inspect in redis-cli:
// 	// > redis-cli ZRANGE prices:exchange1:BTCUSDT 0 -1 WITHSCORES
// }
