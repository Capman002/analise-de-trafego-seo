package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wicomm/analise-trafego/internal/repository"
	"github.com/wicomm/analise-trafego/internal/service"
)

// TrafficHandler gerencia endpoints de dados de tráfego.
type TrafficHandler struct {
	clientRepo  *repository.ClientRepo
	trafficRepo *repository.TrafficRepo
	collector   *service.Collector
}

func NewTrafficHandler(cr *repository.ClientRepo, tr *repository.TrafficRepo, col *service.Collector) *TrafficHandler {
	return &TrafficHandler{clientRepo: cr, trafficRepo: tr, collector: col}
}

// GetTraffic retorna dados agregados de todas as fontes para um cliente.
// GET /api/traffic/:id?period=28&location=BRA
//
// Fluxo:
// 1. Busca o cliente no DB
// 2. Dispara coleta (collector) se não houver dados recentes
// 3. Retorna dados do SQLite
func (h *TrafficHandler) GetTraffic(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	clientID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID de cliente inválido")
		return
	}

	// Período e Datas (Custom)
	period := 28
	if p := r.URL.Query().Get("period"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 && v <= 730 {
			period = v
		}
	}
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	if startDate != "" && endDate != "" {
		period = 0
	}

	dr := repository.CalculateDateRange(period, startDate, endDate)

	// Buscar cliente
	client, err := h.clientRepo.GetByID(clientID)
	if err != nil {
		writeError(w, http.StatusNotFound, "cliente não encontrado")
		return
	}

	// ❌ Fallback de coleta removido conforme diretriz da arquitetura.
	// O backend agora é 100% Read-Only na interação do usuário.
	// O background worker é o único responsável por injetar dados no SQLite.

	// Buscar dados de cada fonte do SQLite
	overview, err := h.trafficRepo.GetTrafficOverview(clientID, dr)
	if err != nil {
		slog.Error("GetTrafficOverview failed", "client_id", clientID, "err", err)
	}
	overviewPrev, err := h.trafficRepo.GetTrafficOverviewPrev(clientID, dr)
	if err != nil {
		slog.Error("GetTrafficOverviewPrev failed", "client_id", clientID, "err", err)
	}
	posDistribution, err := h.trafficRepo.GetPositionDistribution(clientID, dr)
	if err != nil {
		slog.Error("GetPositionDistribution failed", "client_id", clientID, "err", err)
	}
	gscQueries, err := h.trafficRepo.GetGSCData(clientID, dr, "query")
	if err != nil {
		slog.Error("GetGSCData (query) failed", "client_id", clientID, "err", err)
	}
	gscPages, err := h.trafficRepo.GetGSCData(clientID, dr, "page")
	if err != nil {
		slog.Error("GetGSCData (page) failed", "client_id", clientID, "err", err)
	}
	gscChart, err := h.trafficRepo.GetGSCChartData(clientID, dr)
	if err != nil {
		slog.Error("GetGSCChartData failed", "client_id", clientID, "err", err)
	}
	gscChartPrev, err := h.trafficRepo.GetGSCChartDataPrev(clientID, dr)
	if err != nil {
		slog.Error("GetGSCChartDataPrev failed", "client_id", clientID, "err", err)
	}

	gscRiseQueries, err := h.trafficRepo.GetGSCTrending(clientID, dr, "query", "DESC")
	if err != nil {
		slog.Error("GetGSCTrending (query, DESC) failed", "client_id", clientID, "err", err)
	}
	gscRisePages, err := h.trafficRepo.GetGSCTrending(clientID, dr, "page", "DESC")
	if err != nil {
		slog.Error("GetGSCTrending (page, DESC) failed", "client_id", clientID, "err", err)
	}
	gscFallQueries, err := h.trafficRepo.GetGSCTrending(clientID, dr, "query", "ASC")
	if err != nil {
		slog.Error("GetGSCTrending (query, ASC) failed", "client_id", clientID, "err", err)
	}
	gscFallPages, err := h.trafficRepo.GetGSCTrending(clientID, dr, "page", "ASC")
	if err != nil {
		slog.Error("GetGSCTrending (page, ASC) failed", "client_id", clientID, "err", err)
	}

	ga4Data, err := h.trafficRepo.GetGA4Data(clientID, dr)
	if err != nil {
		slog.Error("GetGA4Data failed", "client_id", clientID, "err", err)
	}
	bingData, err := h.trafficRepo.GetBingData(clientID, dr)
	if err != nil {
		slog.Error("GetBingData failed", "client_id", clientID, "err", err)
	}

	// Status de sincronização gravado pelo Worker
	syncStatus, err := h.trafficRepo.GetSyncStatus(clientID)
	if err != nil {
		slog.Error("GetSyncStatus failed", "client_id", clientID, "err", err)
	}

	// Garantir arrays vazios em vez de null
	if gscQueries == nil {
		gscQueries = []map[string]interface{}{}
	}
	if gscPages == nil {
		gscPages = []map[string]interface{}{}
	}
	if gscChart == nil {
		gscChart = []map[string]interface{}{}
	}
	if gscChartPrev == nil {
		gscChartPrev = []map[string]interface{}{}
	}
	if gscRiseQueries == nil {
		gscRiseQueries = []map[string]interface{}{}
	}
	if gscRisePages == nil {
		gscRisePages = []map[string]interface{}{}
	}
	if gscFallQueries == nil {
		gscFallQueries = []map[string]interface{}{}
	}
	if gscFallPages == nil {
		gscFallPages = []map[string]interface{}{}
	}
	if ga4Data == nil {
		ga4Data = []map[string]interface{}{}
	}
	if bingData == nil {
		bingData = []map[string]interface{}{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"client":                client,
		"period":                period,
		"startDate":             dr.CurrentStart,
		"endDate":               dr.CurrentEnd,
		"overview":              overview,
		"overview_prev":         overviewPrev,
		"position_distribution": posDistribution,
		"gsc_queries":           gscQueries,
		"gsc_pages":             gscPages,
		"gsc_chart":             gscChart,
		"gsc_chart_prev":        gscChartPrev,
		"gsc_rise_queries":      gscRiseQueries,
		"gsc_rise_pages":        gscRisePages,
		"gsc_fall_queries":      gscFallQueries,
		"gsc_fall_pages":        gscFallPages,
		"ga4":                   ga4Data,
		"bing":                  bingData,
		"sync_status":           syncStatus,
	})
}
