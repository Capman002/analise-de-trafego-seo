package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/wicomm/analise-trafego/internal/model"
	"github.com/wicomm/analise-trafego/internal/repository"
)

// SheetsService sincroniza a lista de clientes a partir de uma planilha
// pública do Google Sheets exportada como CSV.
type SheetsService struct {
	csvURL     string
	clientRepo *repository.ClientRepo
	httpClient *http.Client
}

func NewSheetsService(csvURL string, clientRepo *repository.ClientRepo) *SheetsService {
	return &SheetsService{
		csvURL:     csvURL,
		clientRepo: clientRepo,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SyncClients busca o CSV e faz upsert de cada cliente no banco.
// Retorna o número de clientes sincronizados.
func (s *SheetsService) SyncClients() (int, error) {
	slog.Info("sincronizando clientes da planilha", "url", s.csvURL)

	resp, err := s.httpClient.Get(s.csvURL)
	if err != nil {
		return 0, fmt.Errorf("falha ao buscar planilha: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("planilha retornou status %d", resp.StatusCode)
	}

	clients, err := parseCSV(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("falha ao parsear CSV: %w", err)
	}

	count := 0
	for _, c := range clients {
		if c.Name == "" {
			continue
		}
		if err := s.clientRepo.Upsert(c); err != nil {
			slog.Error("falha ao upsert cliente", "name", c.Name, "err", err)
			continue
		}
		count++
	}

	slog.Info("clientes sincronizados", "total", count)
	return count, nil
}

// parseCSV lê o CSV e converte cada linha em um model.Client.
// Colunas esperadas: Cliente, Roadmap, Url do GSC, Tipo, Conta SC, Conta GA4, ID do GA4, Regex marca
func parseCSV(r io.Reader) ([]model.Client, error) {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Lê header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("falha ao ler header: %w", err)
	}

	// Mapeia índices por nome da coluna (resiliente à reordenação)
	idx := make(map[string]int)
	for i, col := range header {
		idx[strings.TrimSpace(col)] = i
	}

	// Valida colunas obrigatórias
	required := []string{"Cliente"}
	for _, col := range required {
		if _, ok := idx[col]; !ok {
			return nil, fmt.Errorf("coluna obrigatória %q não encontrada no header", col)
		}
	}

	var clients []model.Client
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Warn("linha inválida no CSV, pulando", "err", err)
			continue
		}

		name := getField(record, idx, "Cliente")
		if name == "" {
			continue
		}

		gscURL := getField(record, idx, "Url do GSC")
		gscType := normalizeGscType(getField(record, idx, "Tipo"))
		ga4ID := getField(record, idx, "ID do GA4")
		brandRegex := getField(record, idx, "Regex marca")

		clients = append(clients, model.Client{
			Name:       strings.TrimSpace(name),
			GscURL:     strings.TrimSpace(gscURL),
			GscType:    gscType,
			Ga4ID:      strings.TrimSpace(ga4ID),
			BrandRegex: strings.TrimSpace(brandRegex),
		})
	}

	return clients, nil
}

func getField(record []string, idx map[string]int, col string) string {
	i, ok := idx[col]
	if !ok || i >= len(record) {
		return ""
	}
	return record[i]
}

// normalizeGscType converte "Domínio"/"Url" para "domain"/"url".
func normalizeGscType(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "domínio", "dominio", "domain":
		return "domain"
	case "url":
		return "url"
	default:
		return ""
	}
}
