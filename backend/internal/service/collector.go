package service

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/wicomm/analise-trafego/internal/model"
	"github.com/wicomm/analise-trafego/internal/repository"
)

// Collector orquestra a coleta de dados de todas as fontes para um cliente.
// Verifica cache antes de buscar: se dados existem com < maxAgeHours, pula a fonte.
type Collector struct {
	gsc         *GSCService
	ga4         *GA4Service
	bing        *BingService
	trafficRepo *repository.TrafficRepo
	clientRepo  *repository.ClientRepo
	maxAgeHours int
}

func NewCollector(
	gsc *GSCService,
	ga4 *GA4Service,
	bing *BingService,
	trafficRepo *repository.TrafficRepo,
	clientRepo *repository.ClientRepo,
) *Collector {
	return &Collector{
		gsc:         gsc,
		ga4:         ga4,
		bing:        bing,
		trafficRepo: trafficRepo,
		clientRepo:  clientRepo,
		maxAgeHours: 6,
	}
}

// CollectResult informa o status da coleta de cada fonte.
type CollectResult struct {
	GSC  string // "ok", "cached", "error: ...", "skip"
	GA4  string
	Bing string
}

// CollectAll busca dados das 3 fontes em paralelo para um cliente.
// Se force = false e dados recentes (<6h) existem no SQLite, pula a coleta da fonte.
func (c *Collector) CollectAll(client *model.Client, periodDays int, force bool) CollectResult {
	result := CollectResult{GSC: "skip", GA4: "skip", Bing: "skip"}

	// Calcular date range — GSC tem ~2 dias de atraso na disponibilização dos dados
	end := time.Now().AddDate(0, 0, -2)
	start := end.AddDate(0, 0, -periodDays)
	startDate := start.Format("2006-01-02")
	endDate := end.Format("2006-01-02")

	// GA4 tem ~1 dia de atraso
	ga4End := time.Now().UTC().AddDate(0, 0, -1)
	ga4Start := ga4End.AddDate(0, 0, -periodDays)
	ga4StartDate := ga4Start.Format("2006-01-02")
	ga4EndDate := ga4End.Format("2006-01-02")

	var wg sync.WaitGroup
	var mu sync.Mutex

	// ── GSC ──────────────────────────────────────────────────────
	if c.gsc != nil && client.GscURL != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if !force {
				cached, _ := c.trafficRepo.HasRecentData(client.ID, "gsc", c.maxAgeHours)
				if cached {
					mu.Lock()
					result.GSC = "cached"
					mu.Unlock()
					slog.Info("[Collector] GSC cached", "client", client.Name)
					return
				}
			}

			slog.Info("[Collector] coletando GSC", "client", client.Name, "url", client.GscURL)

			// GSC API: domain properties usam prefixo "sc-domain:"
			gscSiteURL := client.GscURL
			if client.GscType == "domain" {
				gscSiteURL = "sc-domain:" + client.GscURL
			}

			// Buscar pages, queries e totais em paralelo
			var pages, queries, totals []GSCRow
			var errPages, errQueries, errTotals error
			var inner sync.WaitGroup

			inner.Add(3)
			go func() {
				defer inner.Done()
				pages, errPages = c.gsc.FetchDailyData(gscSiteURL, "page", startDate, endDate, "")
			}()
			go func() {
				defer inner.Done()
				queries, errQueries = c.gsc.FetchDailyData(gscSiteURL, "query", startDate, endDate, "")
			}()
			go func() {
				defer inner.Done()
				totals, errTotals = c.gsc.FetchDailyTotals(gscSiteURL, startDate, endDate, "")
			}()
			inner.Wait()

			if errPages != nil && errQueries != nil {
				mu.Lock()
				result.GSC = fmt.Sprintf("error: pages=%v, queries=%v", errPages, errQueries)
				mu.Unlock()
				c.trafficRepo.UpsertSyncStatus(client.ID, "gsc", "error", result.GSC)
				slog.Error("[Collector] GSC falhou", "client", client.Name, "errPages", errPages, "errQueries", errQueries)
				return
			}
			if errTotals != nil {
				slog.Warn("[Collector] GSC totais falhou (non-fatal)", "client", client.Name, "err", errTotals)
			}

			// Converter service.GSCRow → model.GSCRow para o repository
			allRows := append(pages, queries...)
			allRows = append(allRows, totals...)
			modelRows := make([]model.GSCRow, len(allRows))
			for i, r := range allRows {
				modelRows[i] = model.GSCRow{
					Date: r.Date, Dimension: r.Dimension, Key: r.Key,
					Clicks: r.Clicks, Impressions: r.Impressions,
					CTR: r.CTR, Position: r.Position,
				}
			}

			count, err := c.trafficRepo.UpsertGSCData(client.ID, modelRows)
			if err != nil {
				mu.Lock()
				result.GSC = fmt.Sprintf("error: persist=%v", err)
				mu.Unlock()
				c.trafficRepo.UpsertSyncStatus(client.ID, "gsc", "error", result.GSC)
				return
			}

			mu.Lock()
			result.GSC = fmt.Sprintf("ok (%d rows)", count)
			mu.Unlock()
			c.trafficRepo.UpsertSyncStatus(client.ID, "gsc", "success", "")
			slog.Info("[Collector] GSC completo", "client", client.Name, "rows", count)
		}()
	}

	// ── GA4 ──────────────────────────────────────────────────────
	if c.ga4 != nil && client.Ga4ID != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if !force {
				cached, _ := c.trafficRepo.HasRecentData(client.ID, "ga4", c.maxAgeHours)
				if cached {
					mu.Lock()
					result.GA4 = "cached"
					mu.Unlock()
					slog.Info("[Collector] GA4 cached", "client", client.Name)
					return
				}
			}

			slog.Info("[Collector] coletando GA4", "client", client.Name, "propertyID", client.Ga4ID)

			rows, err := c.ga4.FetchLandingPages(client.Ga4ID, ga4StartDate, ga4EndDate)
			if err != nil {
				mu.Lock()
				result.GA4 = fmt.Sprintf("error: %v", err)
				mu.Unlock()
				c.trafficRepo.UpsertSyncStatus(client.ID, "ga4", "error", result.GA4)
				slog.Error("[Collector] GA4 falhou", "client", client.Name, "err", err)
				return
			}

			// Converter service.GA4Row → model.GA4Row
			modelRows := make([]model.GA4Row, len(rows))
			for i, r := range rows {
				modelRows[i] = model.GA4Row{
					Date: r.Date, ItemName: r.ItemName,
					Sessions: r.Sessions, EngagedSessions: r.EngagedSessions,
					Conversions: r.Conversions, Revenue: r.Revenue,
					ItemsPurchased: r.ItemsPurchased,
				}
			}

			count, err := c.trafficRepo.UpsertGA4Data(client.ID, modelRows)
			if err != nil {
				mu.Lock()
				result.GA4 = fmt.Sprintf("error: persist=%v", err)
				mu.Unlock()
				c.trafficRepo.UpsertSyncStatus(client.ID, "ga4", "error", result.GA4)
				return
			}

			mu.Lock()
			result.GA4 = fmt.Sprintf("ok (%d rows)", count)
			mu.Unlock()
			c.trafficRepo.UpsertSyncStatus(client.ID, "ga4", "success", "")
			slog.Info("[Collector] GA4 completo", "client", client.Name, "rows", count)
		}()
	}

	// ── Bing ─────────────────────────────────────────────────────
	if c.bing != nil && client.GscURL != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if !force {
				cached, _ := c.trafficRepo.HasRecentData(client.ID, "bing", c.maxAgeHours)
				if cached {
					mu.Lock()
					result.Bing = "cached"
					mu.Unlock()
					slog.Info("[Collector] Bing cached", "client", client.Name)
					return
				}
			}

			slog.Info("[Collector] coletando Bing", "client", client.Name)

			rows, err := c.bing.FetchQueryStats(client.GscURL, startDate, endDate)
			if err != nil {
				mu.Lock()
				result.Bing = fmt.Sprintf("error: %v", err)
				mu.Unlock()
				c.trafficRepo.UpsertSyncStatus(client.ID, "bing", "error", result.Bing)
				slog.Error("[Collector] Bing falhou", "client", client.Name, "err", err)
				return
			}

			// Converter service.BingRow → model.BingRow
			modelRows := make([]model.BingRow, len(rows))
			for i, r := range rows {
				modelRows[i] = model.BingRow{
					Date: r.Date, Dimension: r.Dimension, Key: r.Key,
					Clicks: r.Clicks, Impressions: r.Impressions,
					Position: r.Position,
				}
			}

			count, err := c.trafficRepo.UpsertBingData(client.ID, modelRows)
			if err != nil {
				mu.Lock()
				result.Bing = fmt.Sprintf("error: persist=%v", err)
				mu.Unlock()
				c.trafficRepo.UpsertSyncStatus(client.ID, "bing", "error", result.Bing)
				return
			}

			mu.Lock()
			result.Bing = fmt.Sprintf("ok (%d rows)", count)
			mu.Unlock()
			c.trafficRepo.UpsertSyncStatus(client.ID, "bing", "success", "")
			slog.Info("[Collector] Bing completo", "client", client.Name, "rows", count)
		}()
	}

	wg.Wait()
	return result
}
