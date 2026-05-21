package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/wicomm/analise-trafego/internal/model"
)

// DateRange define as janelas de tempo exatas para a consulta SQL, permitindo datas customizadas ou relativas.
type DateRange struct {
	CurrentStart string
	CurrentEnd   string
	PrevStart    string
	PrevEnd      string
}

// periodToStartDate converte um período em dias para a data de início correspondente.
// Usa meses calendário quando o período corresponde a intervalos padrão (3, 6, 12, 16 meses).
func periodToStartDate(periodDays int) string {
	now := time.Now()
	var start time.Time

	switch periodDays {
	case 90:
		start = now.AddDate(0, -3, 0)
	case 180:
		start = now.AddDate(0, -6, 0)
	case 365:
		start = now.AddDate(0, -12, 0)
	case 480:
		start = now.AddDate(0, -16, 0)
	default:
		start = now.AddDate(0, 0, -periodDays)
	}

	return start.Format("2006-01-02")
}

// CalculateDateRange gera a estrutura DateRange a partir de datas exatas (YYYY-MM-DD) ou de um período (dias).
// Se startDate e endDate estiverem preenchidos, calcula o período anterior com a exata duração (em dias) do atual.
func CalculateDateRange(periodDays int, startDate, endDate string) DateRange {
	if startDate != "" && endDate != "" {
		s, errS := time.Parse("2006-01-02", startDate)
		e, errE := time.Parse("2006-01-02", endDate)
		if errS == nil && errE == nil && !e.Before(s) {
			durationDays := int(e.Sub(s).Hours() / 24)
			pEnd := s.AddDate(0, 0, -1)
			pStart := pEnd.AddDate(0, 0, -durationDays)
			return DateRange{
				CurrentStart: s.Format("2006-01-02"),
				CurrentEnd:   e.Format("2006-01-02"),
				PrevStart:    pStart.Format("2006-01-02"),
				PrevEnd:      pEnd.Format("2006-01-02"),
			}
		}
	}

	// Fallback: usar periodDays relativo
	now := time.Now()
	var cStart, pStart time.Time

	switch periodDays {
	case 90:
		cStart = now.AddDate(0, -3, 0)
		pStart = now.AddDate(0, -6, 0)
	case 180:
		cStart = now.AddDate(0, -6, 0)
		pStart = now.AddDate(0, -12, 0)
	case 365:
		cStart = now.AddDate(0, -12, 0)
		pStart = now.AddDate(0, -24, 0)
	case 480:
		cStart = now.AddDate(0, -16, 0)
		pStart = now.AddDate(0, -32, 0)
	default:
		cStart = now.AddDate(0, 0, -periodDays)
		pStart = now.AddDate(0, 0, -(periodDays * 2))
	}
	cEnd := now
	pEnd := cStart.AddDate(0, 0, -1)

	return DateRange{
		CurrentStart: cStart.Format("2006-01-02"),
		CurrentEnd:   cEnd.Format("2006-01-02"),
		PrevStart:    pStart.Format("2006-01-02"),
		PrevEnd:      pEnd.Format("2006-01-02"),
	}
}

// TrafficRepo gerencia operações CRUD sobre as tabelas de dados de tráfego.
type TrafficRepo struct {
	db *sql.DB
}

func NewTrafficRepo(db *sql.DB) *TrafficRepo {
	return &TrafficRepo{db: db}
}

// HasRecentData verifica se existem dados de uma fonte para um cliente
// mais recentes que maxAgeHours horas.
func (r *TrafficRepo) HasRecentData(clientID int64, source string, maxAgeHours int) (bool, error) {
	table := tableForSource(source)
	if table == "" {
		return false, nil
	}

	var count int
	err := r.db.QueryRow(
		fmt.Sprintf(`SELECT COUNT(*) FROM %s
		WHERE client_id = ?
		AND fetched_at > datetime('now', '-%d hours')`, table, maxAgeHours),
		clientID,
	).Scan(&count)

	return count > 0, err
}

// HasAnyData verifica se o cliente possui qualquer registro histórico (útil para decidir entre 7 ou 480 dias).
func (r *TrafficRepo) HasAnyData(clientID int64) (bool, error) {
	var count int
	// Checar GSC que costuma ser a base principal
	err := r.db.QueryRow(`
		SELECT COUNT(1) FROM gsc_data WHERE client_id = ? LIMIT 1
	`, clientID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpsertSyncStatus salva o status de coleta de uma fonte para o cliente.
func (r *TrafficRepo) UpsertSyncStatus(clientID int64, source string, status string, errMsg string) error {
	_, err := r.db.Exec(`
		INSERT INTO client_sync_status (client_id, source, status, error_message, last_sync_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(client_id, source) DO UPDATE SET
			status = EXCLUDED.status,
			error_message = EXCLUDED.error_message,
			last_sync_at = CURRENT_TIMESTAMP
	`, clientID, source, status, errMsg)
	return err
}

// GetSyncStatus retorna os status das 3 fontes para um cliente (gsc, ga4, bing).
func (r *TrafficRepo) GetSyncStatus(clientID int64) (map[string]map[string]string, error) {
	rows, err := r.db.Query(`
		SELECT source, status, COALESCE(error_message, ''), last_sync_at
		FROM client_sync_status
		WHERE client_id = ?
	`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]map[string]string)
	for rows.Next() {
		var source, status, errMsg, lastSync string
		if err := rows.Scan(&source, &status, &errMsg, &lastSync); err != nil {
			continue
		}
		result[source] = map[string]string{
			"status":       status,
			"error":        errMsg,
			"last_sync_at": lastSync,
		}
	}
	return result, nil
}

// GetGSCData retorna dados AGREGADOS do Search Console para um cliente.
// Agrupa por key (query ou page) com SUM de cliques/impressões e médias ponderadas
// de CTR e posição — exatamente como o Google Search Console exibe.
// Complexidade: O(n log n) onde n = linhas no período. Com índice cobertura, O(k log n) onde k = keys únicas.
func (r *TrafficRepo) GetGSCData(clientID int64, dr DateRange, dimension string) ([]map[string]interface{}, error) {
	cStart, cEnd, pStart, pEnd := dr.CurrentStart, dr.CurrentEnd, dr.PrevStart, dr.PrevEnd

	return r.queryTrafficData(`
		SELECT 
			k.key,
			c.clicks, c.impressions, c.ctr, c.position,
			p.prev_clicks, p.prev_impressions, p.prev_ctr, p.prev_position
		FROM (
			SELECT key FROM gsc_data
			WHERE client_id = ? AND dimension = ? AND ((date >= ? AND date <= ?) OR (date >= ? AND date <= ?))
			GROUP BY key
		) k
		LEFT JOIN (
			SELECT 
				key,
				SUM(clicks) as clicks,
				SUM(impressions) as impressions,
				CASE WHEN SUM(impressions) > 0 THEN CAST(SUM(clicks) AS REAL) / SUM(impressions) ELSE 0 END as ctr,
				CASE WHEN SUM(impressions) > 0 THEN SUM(position * impressions) / SUM(impressions) ELSE 0 END as position
			FROM gsc_data
			WHERE client_id = ? AND dimension = ? AND date >= ? AND date <= ?
			GROUP BY key
		) c ON k.key = c.key
		LEFT JOIN (
			SELECT 
				key,
				SUM(clicks) as prev_clicks,
				SUM(impressions) as prev_impressions,
				CASE WHEN SUM(impressions) > 0 THEN CAST(SUM(clicks) AS REAL) / SUM(impressions) ELSE 0 END as prev_ctr,
				CASE WHEN SUM(impressions) > 0 THEN SUM(position * impressions) / SUM(impressions) ELSE 0 END as prev_position
			FROM gsc_data
			WHERE client_id = ? AND dimension = ? AND date >= ? AND date <= ?
			GROUP BY key
		) p ON k.key = p.key
		ORDER BY COALESCE(c.clicks, 0) DESC, COALESCE(p.prev_clicks, 0) DESC
		LIMIT 200
	`, clientID, dimension, cStart, cEnd, pStart, pEnd, clientID, dimension, cStart, cEnd, clientID, dimension, pStart, pEnd)
}

// GetGSCTrending retorna a variação de cliques (maiores altas ou maiores quedas)
// cruzando as chaves (query ou page) do período atual com o período imediatamente anterior.
// sortDirection deve ser "DESC" para altas e "ASC" para quedas.
func (r *TrafficRepo) GetGSCTrending(clientID int64, dr DateRange, dimension string, sortDirection string) ([]map[string]interface{}, error) {
	cStart, cEnd, pStart, pEnd := dr.CurrentStart, dr.CurrentEnd, dr.PrevStart, dr.PrevEnd

	// Previne SQL Injection no Sort
	orderClause := "DESC"
	if sortDirection == "ASC" {
		orderClause = "ASC"
	}

	query := fmt.Sprintf(`
		SELECT 
			k.key, 
			COALESCE(c.clicks, 0) as clicks, 
			COALESCE(c.impressions, 0) as impressions, 
			COALESCE(p.prev_clicks, 0) as prev_clicks,
			(COALESCE(c.clicks, 0) - COALESCE(p.prev_clicks, 0)) as clicks_diff
		FROM (
			SELECT key FROM gsc_data
			WHERE client_id = ? AND dimension = ? AND ((date >= ? AND date <= ?) OR (date >= ? AND date <= ?))
			GROUP BY key
		) k
		LEFT JOIN (
			SELECT key, SUM(clicks) as clicks, SUM(impressions) as impressions
			FROM gsc_data
			WHERE client_id = ? AND dimension = ? AND date >= ? AND date <= ?
			GROUP BY key
		) c ON k.key = c.key
		LEFT JOIN (
			SELECT key, SUM(clicks) as prev_clicks, SUM(impressions) as prev_impressions
			FROM gsc_data
			WHERE client_id = ? AND dimension = ? AND date >= ? AND date <= ?
			GROUP BY key
		) p ON k.key = p.key
		WHERE (COALESCE(c.clicks, 0) - COALESCE(p.prev_clicks, 0)) != 0
		ORDER BY clicks_diff %s, COALESCE(c.clicks, 0) DESC
		LIMIT 100
	`, orderClause)

	return r.queryTrafficData(query, clientID, dimension, cStart, cEnd, pStart, pEnd, clientID, dimension, cStart, cEnd, clientID, dimension, pStart, pEnd)
}

// GetGSCChartData retorna a soma de cliques e impressões por dia para o gráfico.
// Usa dimension='total' (coleta sem dimensão cruzada) para números idênticos ao GSC oficial.
func (r *TrafficRepo) GetGSCChartData(clientID int64, dr DateRange) ([]map[string]interface{}, error) {
	return r.queryTrafficData(`
		SELECT 
			date, 
			SUM(clicks) as clicks, 
			SUM(impressions) as impressions,
			CASE WHEN SUM(impressions) > 0 THEN CAST(SUM(clicks) AS REAL) / SUM(impressions) ELSE 0 END as ctr,
			CASE WHEN SUM(impressions) > 0 THEN SUM(position * impressions) / SUM(impressions) ELSE 0 END as position
		FROM gsc_data
		WHERE client_id = ?
		AND dimension = 'total'
		AND date >= ? AND date <= ?
		GROUP BY date
		ORDER BY date ASC
	`, clientID, dr.CurrentStart, dr.CurrentEnd)
}

// GetGSCChartDataPrev retorna a soma de cliques e impressões por dia para o gráfico no período anterior.
func (r *TrafficRepo) GetGSCChartDataPrev(clientID int64, dr DateRange) ([]map[string]interface{}, error) {
	return r.queryTrafficData(`
		SELECT 
			date, 
			SUM(clicks) as clicks, 
			SUM(impressions) as impressions,
			CASE WHEN SUM(impressions) > 0 THEN CAST(SUM(clicks) AS REAL) / SUM(impressions) ELSE 0 END as ctr,
			CASE WHEN SUM(impressions) > 0 THEN SUM(position * impressions) / SUM(impressions) ELSE 0 END as position
		FROM gsc_data
		WHERE client_id = ?
		AND dimension = 'total'
		AND date >= ? AND date <= ?
		GROUP BY date
		ORDER BY date ASC
	`, clientID, dr.PrevStart, dr.PrevEnd)
}

// GetGA4Data retorna dados do GA4 para um cliente, filtrados por período.
func (r *TrafficRepo) GetGA4Data(clientID int64, dr DateRange) ([]map[string]interface{}, error) {
	return r.queryTrafficData(`
		SELECT date, item_name, item_id, sessions, engaged_sessions,
		       conversions, revenue, items_purchased
		FROM ga4_data
		WHERE client_id = ?
		AND date >= ? AND date <= ?
		ORDER BY date DESC, sessions DESC
		LIMIT 100
	`, clientID, dr.CurrentStart, dr.CurrentEnd)
}

// GetBingData retorna dados do Bing para um cliente, filtrados por período. Limitado a 100 linhas para UI.
func (r *TrafficRepo) GetBingData(clientID int64, dr DateRange) ([]map[string]interface{}, error) {
	return r.queryTrafficData(`
		SELECT date, dimension, key, clicks, impressions, position
		FROM bing_data
		WHERE client_id = ?
		AND date >= ? AND date <= ?
		ORDER BY date DESC, clicks DESC
		LIMIT 100
	`, clientID, dr.CurrentStart, dr.CurrentEnd)
}

// GetTrafficOverview retorna métricas agregadas (SUM) calculadas diretamente no SQLite.
// Usa dimension='total' — dados coletados sem dimensão cruzada (apenas date),
// idênticos aos totais que o GSC oficial exibe no topo do dashboard.
func (r *TrafficRepo) GetTrafficOverview(clientID int64, dr DateRange) (map[string]interface{}, error) {
	// GSC: dimension='total' = totais reais sem inflação por dimensão
	gscRow := r.db.QueryRow(`
		SELECT COALESCE(SUM(clicks), 0), COALESCE(SUM(impressions), 0),
			CASE WHEN COALESCE(SUM(impressions), 0) > 0 
				THEN CAST(COALESCE(SUM(clicks), 0) AS REAL) / SUM(impressions) 
				ELSE 0 
			END,
			CASE WHEN COALESCE(SUM(impressions), 0) > 0 
				THEN SUM(position * impressions) / SUM(impressions) 
				ELSE 0 
			END
		FROM gsc_data
		WHERE client_id = ? 
		AND dimension = 'total'
		AND date >= ? AND date <= ?
	`, clientID, dr.CurrentStart, dr.CurrentEnd)

	var gscClicks, gscImpressions int64
	var gscCTR, gscPosition float64
	_ = gscRow.Scan(&gscClicks, &gscImpressions, &gscCTR, &gscPosition)

	ga4Row := r.db.QueryRow(`
		SELECT COALESCE(SUM(sessions), 0), COALESCE(SUM(revenue), 0), COALESCE(SUM(items_purchased), 0)
		FROM ga4_data
		WHERE client_id = ? AND date >= ? AND date <= ?
	`, clientID, dr.CurrentStart, dr.CurrentEnd)

	var ga4Sessions, ga4ItemsPurchased int64
	var ga4Revenue float64
	_ = ga4Row.Scan(&ga4Sessions, &ga4Revenue, &ga4ItemsPurchased)

	return map[string]interface{}{
		"gsc_clicks":          gscClicks,
		"gsc_impressions":     gscImpressions,
		"gsc_ctr":             gscCTR,
		"gsc_position":        gscPosition,
		"ga4_sessions":        ga4Sessions,
		"ga4_revenue":         ga4Revenue,
		"ga4_items_purchased": ga4ItemsPurchased,
	}, nil
}

// GetTrafficOverviewPrev retorna as mesmas métricas agregadas mas para o período
// imediatamente anterior (ex: se period=28, retorna os 28 dias antes dos 28 atuais).
// Usado para calcular Δ% de comparação no frontend.
func (r *TrafficRepo) GetTrafficOverviewPrev(clientID int64, dr DateRange) (map[string]interface{}, error) {
	prevStart, prevEnd := dr.PrevStart, dr.PrevEnd

	gscRow := r.db.QueryRow(`
		SELECT COALESCE(SUM(clicks), 0), COALESCE(SUM(impressions), 0),
			CASE WHEN COALESCE(SUM(impressions), 0) > 0 
				THEN CAST(COALESCE(SUM(clicks), 0) AS REAL) / SUM(impressions) 
				ELSE 0 
			END,
			CASE WHEN COALESCE(SUM(impressions), 0) > 0 
				THEN SUM(position * impressions) / SUM(impressions) 
				ELSE 0 
			END
		FROM gsc_data
		WHERE client_id = ? 
		AND dimension = 'total'
		AND date >= ? AND date <= ?
	`, clientID, prevStart, prevEnd)

	var gscClicks, gscImpressions int64
	var gscCTR, gscPosition float64
	_ = gscRow.Scan(&gscClicks, &gscImpressions, &gscCTR, &gscPosition)

	ga4Row := r.db.QueryRow(`
		SELECT COALESCE(SUM(sessions), 0), COALESCE(SUM(revenue), 0), COALESCE(SUM(items_purchased), 0)
		FROM ga4_data
		WHERE client_id = ? AND date >= ? AND date <= ?
	`, clientID, prevStart, prevEnd)

	var ga4Sessions, ga4ItemsPurchased int64
	var ga4Revenue float64
	_ = ga4Row.Scan(&ga4Sessions, &ga4Revenue, &ga4ItemsPurchased)

	return map[string]interface{}{
		"gsc_clicks":          gscClicks,
		"gsc_impressions":     gscImpressions,
		"gsc_ctr":             gscCTR,
		"gsc_position":        gscPosition,
		"ga4_sessions":        ga4Sessions,
		"ga4_revenue":         ga4Revenue,
		"ga4_items_purchased": ga4ItemsPurchased,
	}, nil
}

// GetPositionDistribution retorna a contagem de queries únicas por faixa de posição
// no período atual e no período anterior, para visualização de distribuição.
// Faixas: 1-3, 4-10, 11-20, 20+
func (r *TrafficRepo) GetPositionDistribution(clientID int64, dr DateRange) (map[string]interface{}, error) {
	cStart, cEnd, pStart, pEnd := dr.CurrentStart, dr.CurrentEnd, dr.PrevStart, dr.PrevEnd

	// Período atual
	currentRow := r.db.QueryRow(`
		SELECT
			COUNT(DISTINCT CASE WHEN avg_pos <= 3 THEN key END),
			COUNT(DISTINCT CASE WHEN avg_pos > 3 AND avg_pos <= 10 THEN key END),
			COUNT(DISTINCT CASE WHEN avg_pos > 10 AND avg_pos <= 20 THEN key END),
			COUNT(DISTINCT CASE WHEN avg_pos > 20 THEN key END)
		FROM (
			SELECT key,
				CASE WHEN SUM(impressions) > 0 
					THEN SUM(position * impressions) / SUM(impressions) 
					ELSE 0 
				END as avg_pos
			FROM gsc_data
			WHERE client_id = ? AND dimension = 'query'
			AND date >= ? AND date <= ?
			GROUP BY key
		)
	`, clientID, cStart, cEnd)

	var top3, top10, top20, beyond20 int64
	_ = currentRow.Scan(&top3, &top10, &top20, &beyond20)

	// Período anterior
	prevRow := r.db.QueryRow(`
		SELECT
			COUNT(DISTINCT CASE WHEN avg_pos <= 3 THEN key END),
			COUNT(DISTINCT CASE WHEN avg_pos > 3 AND avg_pos <= 10 THEN key END),
			COUNT(DISTINCT CASE WHEN avg_pos > 10 AND avg_pos <= 20 THEN key END),
			COUNT(DISTINCT CASE WHEN avg_pos > 20 THEN key END)
		FROM (
			SELECT key,
				CASE WHEN SUM(impressions) > 0 
					THEN SUM(position * impressions) / SUM(impressions) 
					ELSE 0 
				END as avg_pos
			FROM gsc_data
			WHERE client_id = ? AND dimension = 'query'
			AND date >= ? AND date <= ?
			GROUP BY key
		)
	`, clientID, pStart, pEnd)

	var prevTop3, prevTop10, prevTop20, prevBeyond20 int64
	_ = prevRow.Scan(&prevTop3, &prevTop10, &prevTop20, &prevBeyond20)

	return map[string]interface{}{
		"top3":          top3,
		"top10":         top10,
		"top20":         top20,
		"beyond20":      beyond20,
		"prev_top3":     prevTop3,
		"prev_top10":    prevTop10,
		"prev_top20":    prevTop20,
		"prev_beyond20": prevBeyond20,
	}, nil
}

// ── Upsert Methods ──────────────────────────────────────────────

// UpsertGSCData insere ou atualiza linhas de dados GSC em batch via transação.
// Usa ON CONFLICT para evitar duplicatas (client_id, date, dimension, key).
func (r *TrafficRepo) UpsertGSCData(clientID int64, rows []model.GSCRow) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO gsc_data (client_id, date, dimension, key, clicks, impressions, ctr, position, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(client_id, date, dimension, key) DO UPDATE SET
			clicks = EXCLUDED.clicks,
			impressions = EXCLUDED.impressions,
			ctr = EXCLUDED.ctr,
			position = EXCLUDED.position,
			fetched_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return 0, fmt.Errorf("erro ao preparar statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for _, row := range rows {
		_, err := stmt.Exec(clientID, row.Date, row.Dimension, row.Key, row.Clicks, row.Impressions, row.CTR, row.Position)
		if err != nil {
			return count, fmt.Errorf("erro ao inserir GSC row: %w", err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("erro ao commitar transação: %w", err)
	}
	return count, nil
}

// UpsertGA4Data insere linhas de dados GA4 em batch via transação.
func (r *TrafficRepo) UpsertGA4Data(clientID int64, rows []model.GA4Row) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO ga4_data (client_id, date, item_name, sessions, engaged_sessions, conversions, revenue, items_purchased, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(client_id, date, item_name) DO UPDATE SET
			sessions = EXCLUDED.sessions,
			engaged_sessions = EXCLUDED.engaged_sessions,
			conversions = EXCLUDED.conversions,
			revenue = EXCLUDED.revenue,
			items_purchased = EXCLUDED.items_purchased,
			fetched_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return 0, fmt.Errorf("erro ao preparar statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for _, row := range rows {
		_, err := stmt.Exec(clientID, row.Date, row.ItemName, row.Sessions, row.EngagedSessions, row.Conversions, row.Revenue, row.ItemsPurchased)
		if err != nil {
			return count, fmt.Errorf("erro ao inserir GA4 row: %w", err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("erro ao commitar transação: %w", err)
	}
	return count, nil
}

// UpsertBingData insere linhas de dados Bing em batch via transação.
func (r *TrafficRepo) UpsertBingData(clientID int64, rows []model.BingRow) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO bing_data (client_id, date, dimension, key, clicks, impressions, position, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(client_id, date, dimension, key) DO UPDATE SET
			clicks = EXCLUDED.clicks,
			impressions = EXCLUDED.impressions,
			position = EXCLUDED.position,
			fetched_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return 0, fmt.Errorf("erro ao preparar statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for _, row := range rows {
		_, err := stmt.Exec(clientID, row.Date, row.Dimension, row.Key, row.Clicks, row.Impressions, row.Position)
		if err != nil {
			return count, fmt.Errorf("erro ao inserir Bing row: %w", err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("erro ao commitar transação: %w", err)
	}
	return count, nil
}

// ── Helpers ─────────────────────────────────────────────────────

// HasPermissionError verifica se o cliente teve erro de permissão na última coleta GSC.
// Analisa a coluna error_message em client_sync_status procurando indicadores de 403/permission denied.
func (r *TrafficRepo) HasPermissionError(clientID int64) (bool, error) {
	var errMsg string
	err := r.db.QueryRow(`
		SELECT COALESCE(error_message, '')
		FROM client_sync_status
		WHERE client_id = ? AND source = 'gsc' AND status = 'error'
	`, clientID).Scan(&errMsg)
	if err != nil {
		return false, nil // Sem registro = sem erro de permissão
	}

	// Detectar padrões de erro de permissão da API Google
	if errMsg == "" {
		return false, nil
	}

	// GSC API retorna 403 com "User does not have sufficient permission" ou similar
	permissionPatterns := []string{"403", "permission", "forbidden", "User does not have"}
	for _, pattern := range permissionPatterns {
		if containsInsensitive(errMsg, pattern) {
			return true, nil
		}
	}

	return false, nil
}

// containsInsensitive verifica substring case-insensitive.
func containsInsensitive(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > 0 && len(substr) > 0 &&
				containsLower(toLower(s), toLower(substr)))
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// queryTrafficData é um helper genérico que executa uma query e retorna
// os resultados como slice de maps (flexível para diferentes schemas).
func (r *TrafficRepo) queryTrafficData(query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range cols {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results, rows.Err()
}

func tableForSource(source string) string {
	switch source {
	case "gsc":
		return "gsc_data"
	case "ga4":
		return "ga4_data"
	case "bing":
		return "bing_data"
	default:
		return ""
	}
}
