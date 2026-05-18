package handler

import (
	"fmt"
	"net/http"

	"github.com/wicomm/analise-trafego/internal/service"
)

// SyncHandler gerencia endpoints de sincronização.
type SyncHandler struct {
	sheets *service.SheetsService
}

func NewSyncHandler(sheets *service.SheetsService) *SyncHandler {
	return &SyncHandler{sheets: sheets}
}

// SyncClients força re-sync da planilha de clientes.
// POST /api/sync/clients
func (h *SyncHandler) SyncClients(w http.ResponseWriter, r *http.Request) {
	count, err := h.sheets.SyncClients()
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("falha no sync: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"synced": count,
		"status": "ok",
	})
}
