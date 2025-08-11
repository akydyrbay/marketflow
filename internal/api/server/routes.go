package server

import (
	"fmt"
	"marketflow/internal/api/handlers"
	"marketflow/internal/domain"

	"net/http"
	"time"
)

func Setup(db domain.Database, cacheMemory domain.CacheMemory, datafetchServ *DataModeServiceImp) *http.ServeMux {
	modeHandler := handlers.NewSwitchModeHandler(datafetchServ)
	marketHandler := handlers.NewMarketDataHandler(datafetchServ)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /mode/{mode}", modeHandler.SwitchMode) // Switch to MODE

	mux.HandleFunc("GET /health", modeHandler.CheckHealth) // Returns system status

	mux.HandleFunc("GET /prices/{metric}/{symbol}", marketHandler.ProcessMetricQueryByAll)
	mux.HandleFunc("GET /prices/{metric}/{exchange}/{symbol}", marketHandler.ProcessMetricQueryByExchange)
	fmt.Println(time.Now())
	return mux
}
