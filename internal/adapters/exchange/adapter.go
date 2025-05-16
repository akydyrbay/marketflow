package exchange

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	m "marketflow/internal/domain"
)

type ExchangeAdapter struct {
	Addr        string
	Outbound    chan m.PriceUpdate
	BackoffBase time.Duration
}

func (a *ExchangeAdapter) connect() (net.Conn, error) {
	conn, err := net.Dial("tcp", a.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to exchange %s: %w", a.Addr, err)
	}
	fmt.Println("connected to exchange at", a.Addr)
	return conn, nil
}

func (a *ExchangeAdapter) readLoop(ctx context.Context, conn net.Conn) error {
	reader := bufio.NewReader(conn)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			line, err := reader.ReadBytes('\n')
			if err != nil {
				return fmt.Errorf("read error: %w", err)
			}
			var update m.PriceUpdate
			if err := json.Unmarshal(line, &update); err != nil {
				fmt.Println("Invalid json:", string(line))
				continue
			}
			if isValidSymbol(update.Symbol) {
				a.Outbound <- update
			}
		}
	}
}

func (a *ExchangeAdapter) Run(ctx context.Context) {
	backoff := a.BackoffBase
	for {
		conn, err := a.connect()
		if err != nil {
			fmt.Println("retrying in", backoff)
			time.Sleep(backoff)
			if backoff < 5*time.Second {
				backoff *= 2
			}
			continue
		}
		backoff = a.BackoffBase
		err = a.readLoop(ctx, conn)
		fmt.Println("Disconnected from", a.Addr, "due to", err)
		conn.Close()
	}
}

func isValidSymbol(symbol string) bool {
	switch strings.ToUpper(symbol) {
	case "BTCUSDT", "DOGEUSDT", "TONUSDT", "SOLUSDT", "ETHUSDT":
		return true
	default:
		return false
	}
}
