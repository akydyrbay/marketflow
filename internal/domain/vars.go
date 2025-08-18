package domain

import "flag"

// Currencies
const (
	BTCUSDT  string = "BTCUSDT"
	DOGEUSDT string = "DOGEUSDT"
	TONUSDT  string = "TONUSDT"
	SOLUSDT  string = "SOLUSDT"
	ETHUSDT  string = "ETHUSDT"
)

var Symbols = []string{BTCUSDT, DOGEUSDT, TONUSDT, SOLUSDT, ETHUSDT}

var Exchanges = []string{"Exchange1", "Exchange2", "Exchange3", "All"}

// Flags
var (
	Port        = flag.String("port", "8080", "Establishes server port number")
	HelpFlag    = flag.Bool("help", false, "Show help message")
	HelpMessage = "Usage:\n   marketflow [--port <N>]\n   marketflow --help\n\nOptions:\n   --port N\tPort number"
)
