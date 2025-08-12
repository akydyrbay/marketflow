package utils

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"marketflow/internal/domain"
)

func SendMsg(w http.ResponseWriter, code int, msg string) error {
	data := struct {
		Code int    `json:"Code"`
		Msg  string `json:"Message"`
	}{
		Code: code,
		Msg:  msg,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Failed to send message to the client", "error", err.Error())
		return err
	}
	return nil
}

func SendJSON(w http.ResponseWriter, code int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}
	return nil
}

func SendMetricData(w http.ResponseWriter, code int, rawdata domain.Data) error {
	data := struct {
		ExchangeName string  `json:"exchange"`
		Symbol       string  `json:"symbol"`
		Price        float64 `json:"price"`
		Timestamp    string  `json:"timestamp"`
	}{
		ExchangeName: rawdata.ExchangeName,
		Symbol:       rawdata.Symbol,
		Price:        rawdata.Price,
		Timestamp: time.Unix(0, rawdata.Timestamp*int64(time.Millisecond)).
			Format("2006-01-02 15:04:05"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}
	return nil
}
