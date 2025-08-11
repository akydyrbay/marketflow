package handlers

import (
	"fmt"
	"marketflow/internal/api/utils"
	"marketflow/internal/domain"
	"marketflow/pkg/logger"
	"net/http"
)

type ModeHandler struct {
	serv domain.DataModeService
}

func NewSwitchModeHandler(serv domain.DataModeService) *ModeHandler {
	return &ModeHandler{serv: serv}
}

// Core handler for switching datafetcher mode
func (h *ModeHandler) SwitchMode(w http.ResponseWriter, r *http.Request) {
	mode := r.PathValue("mode")
	if code, err := h.serv.SwitchMode(mode); err != nil {
		logger.Error("Failed to switch mode", "message", err.Error())
		utils.SendMsg(w, code, err.Error())
		return
	}

	// Sending message to the client
	msg := fmt.Sprintf("Datafetcher mode switched to %s", mode)
	utils.SendMsg(w, http.StatusOK, msg)
	logger.Info(msg)
}
