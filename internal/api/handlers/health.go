package handlers

import (
	"net/http"

	"marketflow/internal/api/utils"
	"marketflow/pkg/logger"
)

// Core handler for service health checking
func (h *ModeHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	res := h.serv.CheckHealth()

	if err := utils.SendJSON(w, http.StatusOK, res); err != nil {
		logger.Error("Failed to send checkhealth data: " + err.Error())
		utils.SendMsg(w, http.StatusInternalServerError, err.Error())
	}
}
