package model

import "time"

// Client representa um cliente da Wicomm, sincronizado da planilha do Google Sheets.
type Client struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	GscURL          string    `json:"gscUrl"`
	GscType         string    `json:"gscType"` // "domain" | "url"
	Ga4ID           string    `json:"ga4Id"`
	BrandRegex      string    `json:"brandRegex"`
	SyncedAt        time.Time `json:"syncedAt"`
	PermissionError bool      `json:"permissionError"` // true se GSC retornou 403/permissão negada
}

// GSCRow armazena uma linha de dados brutos do Google Search Console.
type GSCRow struct {
	ID          int64     `json:"id"`
	ClientID    int64     `json:"clientId"`
	Date        string    `json:"date"`      // YYYY-MM-DD
	Dimension   string    `json:"dimension"` // "query" | "page"
	Key         string    `json:"key"`
	Clicks      int       `json:"clicks"`
	Impressions int       `json:"impressions"`
	CTR         float64   `json:"ctr"`
	Position    float64   `json:"position"`
	FetchedAt   time.Time `json:"fetchedAt"`
}

// GA4Row armazena uma linha de dados brutos do Google Analytics 4.
type GA4Row struct {
	ID              int64     `json:"id"`
	ClientID        int64     `json:"clientId"`
	Date            string    `json:"date"`
	ItemName        string    `json:"itemName"`
	ItemID          string    `json:"itemId"`
	Sessions        int       `json:"sessions"`
	EngagedSessions int       `json:"engagedSessions"`
	Conversions     int       `json:"conversions"`
	Revenue         float64   `json:"revenue"`
	ItemsPurchased  int       `json:"itemsPurchased"`
	FetchedAt       time.Time `json:"fetchedAt"`
}

// BingRow armazena uma linha de dados brutos do Bing Webmaster.
type BingRow struct {
	ID          int64     `json:"id"`
	ClientID    int64     `json:"clientId"`
	Date        string    `json:"date"`
	Dimension   string    `json:"dimension"`
	Key         string    `json:"key"`
	Clicks      int       `json:"clicks"`
	Impressions int       `json:"impressions"`
	Position    float64   `json:"position"`
	FetchedAt   time.Time `json:"fetchedAt"`
}

// SyncLog registra cada operação de sincronização.
type SyncLog struct {
	ID        int64      `json:"id"`
	ClientID  int64      `json:"clientId"`
	Source    string     `json:"source"` // "gsc" | "ga4" | "bing" | "sheets"
	Status    string     `json:"status"` // "success" | "error" | "running"
	Message   string     `json:"message"`
	StartedAt time.Time  `json:"startedAt"`
	EndedAt   *time.Time `json:"endedAt,omitempty"`
}

// TrafficResponse é a resposta enviada ao frontend com dados agregados.
type TrafficResponse struct {
	Client  Client      `json:"client"`
	Period  int         `json:"period"` // dias
	GSC     interface{} `json:"gsc"`
	GA4     interface{} `json:"ga4"`
	Bing    interface{} `json:"bing"`
	SyncLog []SyncLog   `json:"syncLog"`
}
