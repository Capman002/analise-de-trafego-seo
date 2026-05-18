package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

const gscAPIBase = "https://www.googleapis.com/webmasters/v3/sites"

// GSCService busca dados do Google Search Console via API REST.
type GSCService struct {
	client *http.Client
}

func NewGSCService(client *http.Client) *GSCService {
	return &GSCService{client: client}
}

// gscRequest é o corpo da requisição para searchAnalytics.query.
type gscRequest struct {
	StartDate             string                    `json:"startDate"`
	EndDate               string                    `json:"endDate"`
	Dimensions            []string                  `json:"dimensions"`
	RowLimit              int                       `json:"rowLimit"`
	StartRow              int                       `json:"startRow"`
	DimensionFilterGroups []gscDimensionFilterGroup  `json:"dimensionFilterGroups,omitempty"`
}

type gscDimensionFilterGroup struct {
	Filters []gscFilter `json:"filters"`
}

type gscFilter struct {
	Dimension  string `json:"dimension"`
	Operator   string `json:"operator"`
	Expression string `json:"expression"`
}

type gscAPIResponse struct {
	Rows []gscAPIRow `json:"rows"`
}

type gscAPIRow struct {
	Keys        []string `json:"keys"`
	Clicks      int      `json:"clicks"`
	Impressions int      `json:"impressions"`
	CTR         float64  `json:"ctr"`
	Position    float64  `json:"position"`
}

// GSCRow é o formato normalizado para persistência.
type GSCRow struct {
	Date        string
	Dimension   string // "page" | "query"
	Key         string
	Clicks      int
	Impressions int
	CTR         float64
	Position    float64
}

// FetchDailyData busca dados diários do GSC com paginação.
// dimension: "page" ou "query"
// Retorna dados com granularidade diária (date + dimension).
// Cap de 500k linhas por dimensão para evitar bloquear o pipeline com clientes de alto volume.
// Complexidade: O(n) onde n = total de linhas retornadas pela API (max 500k).
func (s *GSCService) FetchDailyData(siteUrl, dimension, startDate, endDate, location string) ([]GSCRow, error) {
	var allRows []gscAPIRow
	startRow := 0
	rowLimit := 25000
	maxTotalRows := 500000 // Cap para evitar coletas de horas em clientes com milhões de queries

	slog.Info("[GSC] iniciando coleta",
		"siteUrl", siteUrl,
		"dimension", dimension,
		"range", fmt.Sprintf("%s → %s", startDate, endDate),
		"location", location,
	)

	for {
		body := gscRequest{
			StartDate:  startDate,
			EndDate:    endDate,
			Dimensions: []string{"date", dimension},
			RowLimit:   rowLimit,
			StartRow:   startRow,
		}

		if location != "" && location != "ALL" {
			body.DimensionFilterGroups = []gscDimensionFilterGroup{{
				Filters: []gscFilter{{
					Dimension:  "country",
					Operator:   "equals",
					Expression: strings.ToLower(location),
				}},
			}}
		}

		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar request: %w", err)
		}

		// URL encode do siteUrl conforme GSC API exige
		apiURL := fmt.Sprintf("%s/%s/searchAnalytics/query", gscAPIBase, url.QueryEscape(siteUrl))
		req, err := http.NewRequest("POST", apiURL, bytes.NewReader(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("erro ao criar request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("erro na requisição GSC: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("GSC API retornou %d: %s", resp.StatusCode, string(respBody[:min(len(respBody), 500)]))
		}

		if err != nil {
			return nil, fmt.Errorf("erro ao ler resposta: %w", err)
		}

		var result gscAPIResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("erro ao parsear resposta: %w", err)
		}

		allRows = append(allRows, result.Rows...)
		slog.Info("[GSC] página coletada", "dimension", dimension, "rows", len(result.Rows), "total", len(allRows))

		if len(allRows) >= maxTotalRows {
			slog.Warn("[GSC] cap de linhas atingido — truncando", "dimension", dimension, "total", len(allRows), "max", maxTotalRows)
			break
		}
		if len(result.Rows) < rowLimit {
			break
		}
		startRow += rowLimit
	}

	// Converter para formato normalizado.
	// Filtro de relevância: descarta linhas com < 100 impressões diárias (ruído de long-tail).
	minImpressions := 100
	rows := make([]GSCRow, 0, len(allRows)/2) // pre-alloc conservador após filtro
	skipped := 0
	for _, r := range allRows {
		if len(r.Keys) < 2 {
			continue
		}
		if r.Impressions < minImpressions {
			skipped++
			continue
		}
		rows = append(rows, GSCRow{
			Date:        r.Keys[0],
			Dimension:   dimension,
			Key:         r.Keys[1],
			Clicks:      r.Clicks,
			Impressions: r.Impressions,
			CTR:         r.CTR,
			Position:    r.Position,
		})
	}

	slog.Info("[GSC] coleta completa", "dimension", dimension, "total_api", len(allRows), "filtradas", len(rows), "descartadas", skipped)
	return rows, nil
}

// FetchDailyTotals busca dados agregados por dia SEM dimensão page/query.
// Retorna os mesmos totais que o GSC oficial exibe no overview.
// Sem dimensão cruzada, cada impressão é contada uma única vez.
func (s *GSCService) FetchDailyTotals(siteUrl, startDate, endDate, location string) ([]GSCRow, error) {
	var allRows []gscAPIRow

	slog.Info("[GSC] coletando totais diários",
		"siteUrl", siteUrl,
		"range", fmt.Sprintf("%s → %s", startDate, endDate),
	)

	// Totais diários não precisam de paginação — max ~480 linhas (1 por dia)
	body := gscRequest{
		StartDate:  startDate,
		EndDate:    endDate,
		Dimensions: []string{"date"},
		RowLimit:   25000,
		StartRow:   0,
	}

	if location != "" && location != "ALL" {
		body.DimensionFilterGroups = []gscDimensionFilterGroup{{
			Filters: []gscFilter{{
				Dimension:  "country",
				Operator:   "equals",
				Expression: strings.ToLower(location),
			}},
		}}
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar request: %w", err)
	}

	apiURL := fmt.Sprintf("%s/%s/searchAnalytics/query", gscAPIBase, url.QueryEscape(siteUrl))
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na requisição GSC: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GSC API retornou %d: %s", resp.StatusCode, string(respBody[:min(len(respBody), 500)]))
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	var result gscAPIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("erro ao parsear resposta: %w", err)
	}
	allRows = result.Rows

	// Converter — cada row tem apenas Keys[0] = date
	rows := make([]GSCRow, 0, len(allRows))
	for _, r := range allRows {
		if len(r.Keys) < 1 {
			continue
		}
		rows = append(rows, GSCRow{
			Date:        r.Keys[0],
			Dimension:   "total",
			Key:         "_total",
			Clicks:      r.Clicks,
			Impressions: r.Impressions,
			CTR:         r.CTR,
			Position:    r.Position,
		})
	}

	slog.Info("[GSC] totais diários coletados", "dias", len(rows))
	return rows, nil
}
