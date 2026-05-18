package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

const bingAPIBase = "https://ssl.bing.com/webmaster/api.svc/json"

// BingService busca dados do Bing Webmaster Tools via API REST.
type BingService struct {
	apiKey string
}

func NewBingService(apiKey string) *BingService {
	return &BingService{apiKey: apiKey}
}

// BingRow é o formato normalizado para persistência.
type BingRow struct {
	Date        string
	Dimension   string // "query"
	Key         string
	Clicks      int
	Impressions int
	Position    float64
}

// bingAPIResponse é a resposta da Bing API (formato .NET WCF).
type bingAPIResponse struct {
	D []bingAPIRecord `json:"d"`
}

type bingAPIRecord struct {
	Query                string      `json:"Query"`
	Date                 interface{} `json:"Date"`
	Clicks               int        `json:"Clicks"`
	Impressions           int        `json:"Impressions"`
	AvgImpressionPosition float64   `json:"AvgImpressionPosition"`
}

// dotNetDateRegex captura timestamps no formato /Date(1234567890)/
var dotNetDateRegex = regexp.MustCompile(`/Date\((\d+)\)/`)

// normalizeDate converte datas .NET (/Date(timestamp)/) ou ISO para YYYY-MM-DD.
func normalizeDate(raw interface{}) string {
	s := fmt.Sprintf("%v", raw)

	// Formato .NET: /Date(1234567890)/
	if matches := dotNetDateRegex.FindStringSubmatch(s); len(matches) == 2 {
		ts, err := strconv.ParseInt(matches[1], 10, 64)
		if err == nil {
			return time.UnixMilli(ts).UTC().Format("2006-01-02")
		}
	}

	// ISO ou já formatada: pegar apenas os primeiros 10 caracteres
	if len(s) >= 10 {
		return s[:10]
	}

	return s
}

// FetchQueryStats busca estatísticas de queries do Bing para um site.
// Retorna dados diários (sem agregação).
func (s *BingService) FetchQueryStats(siteUrl, startDate, endDate string) ([]BingRow, error) {
	params := url.Values{
		"apikey":    {s.apiKey},
		"siteUrl":   {siteUrl},
		"startDate": {startDate},
		"endDate":   {endDate},
	}

	apiURL := fmt.Sprintf("%s/GetQueryStats?%s", bingAPIBase, params.Encode())

	slog.Info("[Bing] iniciando coleta",
		"siteUrl", siteUrl,
		"range", fmt.Sprintf("%s → %s", startDate, endDate),
	)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na requisição Bing: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Bing API retornou %d: %s", resp.StatusCode, string(body[:min(len(body), 500)]))
	}

	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	var result bingAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("erro ao parsear resposta: %w", err)
	}

	rows := make([]BingRow, 0, len(result.D))
	for _, r := range result.D {
		if r.Query == "" {
			continue
		}
		date := normalizeDate(r.Date)
		if date == "" {
			continue
		}

		position := r.AvgImpressionPosition
		if position == -1 {
			position = 0
		}

		rows = append(rows, BingRow{
			Date:        date,
			Dimension:   "query",
			Key:         r.Query,
			Clicks:      r.Clicks,
			Impressions: r.Impressions,
			Position:    position,
		})
	}

	slog.Info("[Bing] coleta completa", "total_rows", len(rows))
	return rows, nil
}
