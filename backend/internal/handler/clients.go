package handler

import (
	"net/http"

	"github.com/wicomm/analise-trafego/internal/model"
	"github.com/wicomm/analise-trafego/internal/repository"
)

// ClientsHandler gerencia endpoints relacionados a clientes.
type ClientsHandler struct {
	repo        *repository.ClientRepo
	trafficRepo *repository.TrafficRepo
}

func NewClientsHandler(repo *repository.ClientRepo, trafficRepo *repository.TrafficRepo) *ClientsHandler {
	return &ClientsHandler{repo: repo, trafficRepo: trafficRepo}
}

// List retorna todos os clientes com status de permissão.
// GET /api/clients
func (h *ClientsHandler) List(w http.ResponseWriter, r *http.Request) {
	clients, err := h.repo.ListAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "falha ao listar clientes")
		return
	}

	if clients == nil {
		clients = []model.Client{}
	}

	// Enriquecer com status de permissão a partir da tabela client_sync_status
	for i := range clients {
		hasPermError, _ := h.trafficRepo.HasPermissionError(clients[i].ID)
		clients[i].PermissionError = hasPermError
	}

	writeJSON(w, http.StatusOK, clients)
}
