package server

import (
	"log/slog"
	"marketflow/internal/domain"
)

// Services health checking logic
func (serv *DataModeServiceImp) CheckHealth() []domain.ConnMsg {
	data := make([]domain.ConnMsg, 0)

	if err := serv.Datafetcher.CheckHealth(); err != nil {
		slog.Error("Cathed error from Datafetcher health: ", "error", err.Error())
		data = append(data, domain.ConnMsg{Connection: "Datafetcher", Status: err.Error()})
	}

	if err := serv.DB.CheckHealth(); err != nil {
		slog.Info("Cathed error from Database health: ", "error", err.Error())
		data = append(data, domain.ConnMsg{Connection: "Database", Status: "unhealthy"})
	}

	if err := serv.Cache.CheckHealth(); err != nil {
		slog.Info("Cathed error from Cache health: ", "error", err.Error())
		data = append(data, domain.ConnMsg{Connection: "Cache", Status: "unhealthy"})
	}

	if len(data) == 0 {
		data = append(data, domain.ConnMsg{Status: "all connections are healthy"})
	}

	return data
}
