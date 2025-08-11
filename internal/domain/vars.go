package domain

const (
	BTCUSDT  string = "BTCUSDT"
	DOGEUSDT string = "DOGEUSDT"
	TONUSDT  string = "TONUSDT"
	SOLUSDT  string = "SOLUSDT"
	ETHUSDT  string = "ETHUSDT"
)

var Symbols = []string{BTCUSDT, DOGEUSDT, TONUSDT, SOLUSDT, ETHUSDT}

var Exchanges = []string{"Exchange1", "Exchange2", "Exchange3", "All"}

func CheckExchangeName(exchange string) error {
	for _, val := range Exchanges {
		if exchange == val {
			return nil
		}
	}
	return ErrInvalidExchangeVal
}

func CheckSymbolName(symbol string) error {
	for _, val := range Symbols {
		if symbol == val {
			return nil
		}
	}
	return ErrInvalidSymbolVal
}
