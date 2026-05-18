package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strconv"
)

const ga4APIBase = "https://analyticsdata.googleapis.com/v1beta/properties"

// GA4Service busca dados do Google Analytics 4 via Data API REST.
type GA4Service struct {
	client *http.Client
}

func NewGA4Service(client *http.Client) *GA4Service {
	return &GA4Service{client: client}
}

// GA4Row é o formato normalizado para persistência (landing pages orgânicas).
type GA4Row struct {
	Date            string
	ItemName        string // path da landing page
	Sessions        int
	EngagedSessions int
	Conversions     int
	Revenue         float64
	ItemsPurchased  int
}

// ── Tipos internos da GA4 Data API ──────────────────────────────

type ga4ReportRequest struct {
	DateRanges      []ga4DateRange      `json:"dateRanges"`
	Dimensions      []ga4Dimension      `json:"dimensions"`
	Metrics         []ga4Metric         `json:"metrics"`
	Limit           string              `json:"limit"`
	Offset          string              `json:"offset"`
	OrderBys        []ga4OrderBy        `json:"orderBys"`
	DimensionFilter interface{}         `json:"dimensionFilter,omitempty"`
}

type ga4DateRange struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type ga4Dimension struct {
	Name string `json:"name"`
}

type ga4Metric struct {
	Name string `json:"name"`
}

type ga4OrderBy struct {
	Metric ga4MetricOrder `json:"metric"`
	Desc   bool           `json:"desc"`
}

type ga4MetricOrder struct {
	MetricName string `json:"metricName"`
}

type ga4ReportResponse struct {
	Rows     []ga4Row `json:"rows"`
	RowCount int      `json:"rowCount"`
}

type ga4Row struct {
	DimensionValues []ga4Value `json:"dimensionValues"`
	MetricValues    []ga4Value `json:"metricValues"`
}

type ga4Value struct {
	Value string `json:"value"`
}

// Filtro para tráfego orgânico Google
var filterOrganicSEO = map[string]interface{}{
	"filter": map[string]interface{}{
		"fieldName": "sessionSourceMedium",
		"stringFilter": map[string]interface{}{
			"matchType": "EXACT",
			"value":     "google / organic",
		},
	},
}

// FetchLandingPages busca dados de landing pages orgânicas do GA4.
// Métricas: sessions, activeUsers, engagedSessions, engagementRate, pageViews, revenue, keyEvents
func (s *GA4Service) FetchLandingPages(propertyID, startDate, endDate string) ([]GA4Row, error) {
	slog.Info("[GA4] iniciando coleta landing pages",
		"propertyID", propertyID,
		"range", fmt.Sprintf("%s → %s", startDate, endDate),
	)

	dimensions := []ga4Dimension{
		{Name: "date"},
		{Name: "landingPagePlusQueryString"},
	}

	metrics := []ga4Metric{
		{Name: "sessions"},
		{Name: "activeUsers"},
		{Name: "engagedSessions"},
		{Name: "engagementRate"},
		{Name: "screenPageViews"},
		{Name: "purchaseRevenue"},
		{Name: "keyEvents"},
	}

	apiRows, err := s.runPaginatedReport(propertyID, dimensions, metrics, startDate, endDate, "sessions", filterOrganicSEO)
	if err != nil {
		return nil, err
	}

	// Converter para formato normalizado
	rows := make([]GA4Row, 0, len(apiRows))
	for _, r := range apiRows {
		if len(r.DimensionValues) < 2 || len(r.MetricValues) < 7 {
			continue
		}

		date := normalizeGA4Date(r.DimensionValues[0].Value)
		path := r.DimensionValues[1].Value

		sessions := parseIntSafe(r.MetricValues[0].Value)
		engaged := parseIntSafe(r.MetricValues[2].Value)
		revenue := parseFloatSafe(r.MetricValues[5].Value)
		keyEvents := parseIntSafe(r.MetricValues[6].Value)

		rows = append(rows, GA4Row{
			Date:            date,
			ItemName:        path,
			Sessions:        sessions,
			EngagedSessions: engaged,
			Conversions:     keyEvents,
			Revenue:         math.Round(revenue*100) / 100,
			ItemsPurchased:  0, // landing pages não têm esta métrica diretamente
		})
	}

	slog.Info("[GA4] coleta completa", "total_rows", len(rows))
	return rows, nil
}

// runPaginatedReport executa um report GA4 com paginação automática.
// Limite por página: 250.000 linhas.
func (s *GA4Service) runPaginatedReport(
	propertyID string,
	dimensions []ga4Dimension,
	metrics []ga4Metric,
	startDate, endDate string,
	orderByMetric string,
	dimensionFilter interface{},
) ([]ga4Row, error) {
	limit := 250000
	var allRows []ga4Row
	offset := 0

	for {
		body := ga4ReportRequest{
			DateRanges:      []ga4DateRange{{StartDate: startDate, EndDate: endDate}},
			Dimensions:      dimensions,
			Metrics:         metrics,
			Limit:           strconv.Itoa(limit),
			Offset:          strconv.Itoa(offset),
			OrderBys:        []ga4OrderBy{{Metric: ga4MetricOrder{MetricName: orderByMetric}, Desc: true}},
			DimensionFilter: dimensionFilter,
		}

		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar request: %w", err)
		}

		apiURL := fmt.Sprintf("%s/%s:runReport", ga4APIBase, propertyID)
		req, err := http.NewRequest("POST", apiURL, bytes.NewReader(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("erro ao criar request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("erro na requisição GA4: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("GA4 API retornou %d: %s", resp.StatusCode, string(respBody[:min(len(respBody), 500)]))
		}

		if err != nil {
			return nil, fmt.Errorf("erro ao ler resposta: %w", err)
		}

		var result ga4ReportResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("erro ao parsear resposta: %w", err)
		}

		allRows = append(allRows, result.Rows...)
		slog.Info("[GA4] página coletada", "rows", len(result.Rows), "total", len(allRows), "rowCount", result.RowCount)

		if len(allRows) >= result.RowCount || len(result.Rows) < limit {
			break
		}
		offset += limit
	}

	return allRows, nil
}

// normalizeGA4Date converte "20250101" para "2025-01-01".
func normalizeGA4Date(value string) string {
	if len(value) == 8 {
		return value[:4] + "-" + value[4:6] + "-" + value[6:8]
	}
	return value
}

func parseIntSafe(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func parseFloatSafe(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
