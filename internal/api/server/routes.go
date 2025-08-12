package server

import (
	"fmt"
	"net/http"
	"time"

	"marketflow/internal/api/handlers"
	"marketflow/internal/domain"
)

func Setup(db domain.Database, cacheMemory domain.CacheMemory, datafetch *DataModeServiceImp) *http.ServeMux {
	modeHandler := handlers.NewSwitchModeHandler(datafetch)
	marketHandler := handlers.NewMarketDataHandler(datafetch)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /mode/{mode}", modeHandler.SwitchMode) // Switch to MODE

	mux.HandleFunc("GET /health", modeHandler.CheckHealth) // Returns system status

	mux.HandleFunc("GET /prices/{metric}/{symbol}", marketHandler.ProcessMetricQueryByAll)
	mux.HandleFunc("GET /prices/{metric}/{exchange}/{symbol}", marketHandler.ProcessMetricQueryByExchange)
	fmt.Println(time.Now())
	return mux
}
