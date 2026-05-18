package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/wicomm/analise-trafego/internal/repository"
)

// PreWarmer é responsável por popular o cache do banco em background (Cron Job).
type PreWarmer struct {
	collector   *Collector
	clientRepo  *repository.ClientRepo
	trafficRepo *repository.TrafficRepo
}

func NewPreWarmer(collector *Collector, clientRepo *repository.ClientRepo, trafficRepo *repository.TrafficRepo) *PreWarmer {
	return &PreWarmer{
		collector:   collector,
		clientRepo:  clientRepo,
		trafficRepo: trafficRepo,
	}
}

// Run inicia o loop do background worker.
// Primeiro realiza um povoamento inicial (se necessário), e depois agenda para rodar diariamente às 00:00.
func (pw *PreWarmer) Run(ctx context.Context) {
	slog.Info("[PreWarmer] Iniciando rotina de background para aquecimento de cache...")

	// 1. Povoamento Inicial (Logo após o boot do servidor)
	pw.processAllClients(ctx)

	// 2. Loop Diário (Cron)
	for {
		now := time.Now()
		// Calcular tempo até a próxima meia-noite
		nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		durationUntilMidnight := nextMidnight.Sub(now)

		slog.Info("[PreWarmer] Próximo povoamento agendado", "duracao", durationUntilMidnight, "hora", nextMidnight)

		// Dormir até a meia-noite ou cancelar
		select {
		case <-ctx.Done():
			slog.Info("[PreWarmer] Desligamento solicitado. Encerrando worker...")
			return
		case <-time.After(durationUntilMidnight):
			slog.Info("[PreWarmer] Executando povoamento diário (00:00)")
			pw.processAllClients(ctx)
		}
	}
}

// processAllClients itera sequencialmente para evitar esgotar a cota da API.
func (pw *PreWarmer) processAllClients(ctx context.Context) {
	clients, err := pw.clientRepo.ListAll()
	if err != nil {
		slog.Error("[PreWarmer] Falha ao buscar clientes para pre-warming", "err", err)
		return
	}

	for i, client := range clients {
		select {
		case <-ctx.Done():
			slog.Info("[PreWarmer] Povoamento cancelado durante execução.")
			return
		default:
		}

		hasData, err := pw.trafficRepo.HasAnyData(client.ID)
		if err != nil {
			slog.Error("[PreWarmer] Erro ao verificar histórico", "client", client.Name, "err", err)
			continue
		}

		// Definir período de busca
		// Se não tem dados, buscar os últimos 16 meses (~480 dias)
		// Se já tem dados, buscar apenas os últimos 7 dias (o UPSERT atualiza a base mantendo o passado).
		period := 7
		if !hasData {
			period = 480
			slog.Info("[PreWarmer] 1º Povoamento detectado (16 meses)", "client", client.Name)
		} else {
			slog.Info("[PreWarmer] Atualização Diária detectada (7 dias)", "client", client.Name)
		}

		slog.Info("[PreWarmer] Processando", "client", client.Name, "progresso", i+1, "total", len(clients))

		// O CollectAll ignora a coleta se a flag de cache de 6 horas for válida.
		// Como queremos forçar a atualização às 00:00, ou buscar os 16 meses,
		// precisamos garantir que o CollectAll execute. O HasRecentData do Collector
		// pode bloquear a execução diária se tiver sido coletado < 6h.
		// Porém, à meia noite, os dados do dia viraram (nova data no GSC), então a UI pedirá novos dias de qualquer modo.
		// Para o prewarmer funcionar melhor, chamamos CollectAll (ele vai respeitar o cache de 6h local se não forçarmos).
		// Aqui nós usamos force=true para BYPASS do cache local e baixar do Google de qualquer modo.

		result := pw.collector.CollectAll(&client, period, true)

		slog.Info("[PreWarmer] Finalizado", "client", client.Name, "gsc", result.GSC, "ga4", result.GA4)

		// Delay de 2 segundos entre clientes para proteger a cota da API (Rate Limiting Profile)
		if i < len(clients)-1 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}

	slog.Info("[PreWarmer] Povoamento completo para todos os clientes.")
}
