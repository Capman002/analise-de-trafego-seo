package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/wicomm/analise-trafego/internal/auth"
	"github.com/wicomm/analise-trafego/internal/config"
	"github.com/wicomm/analise-trafego/internal/database"
	"github.com/wicomm/analise-trafego/internal/handler"
	"github.com/wicomm/analise-trafego/internal/middleware"
	"github.com/wicomm/analise-trafego/internal/repository"
	"github.com/wicomm/analise-trafego/internal/service"
	"github.com/wicomm/analise-trafego/web"
)

func main() {
	// Contexto global para Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Logger estruturado
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Configuração
	cfg, err := config.Load()
	if err != nil {
		slog.Error("falha ao carregar configuração", "err", err)
		os.Exit(1)
	}

	// Database
	db, err := database.Open(cfg.DBPath)
	if err != nil {
		slog.Error("falha ao abrir banco de dados", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Repositórios
	clientRepo := repository.NewClientRepo(db)
	trafficRepo := repository.NewTrafficRepo(db)

	// Services
	sheetsService := service.NewSheetsService(cfg.SheetsCSVURL, clientRepo)

	// ── Google OAuth2 Client (opcional) ──────────────────────────
	var gscService *service.GSCService
	var ga4Service *service.GA4Service

	if cfg.GoogleCredentialsJSON != "" && cfg.GoogleTokenJSON != "" {
		googleClient, err := auth.NewGoogleClient(cfg.GoogleCredentialsJSON, cfg.GoogleTokenJSON)
		if err != nil {
			slog.Warn("falha ao criar Google OAuth2 client — coleta GSC/GA4 desabilitada", "err", err)
		} else {
			gscService = service.NewGSCService(googleClient)
			ga4Service = service.NewGA4Service(googleClient)
			slog.Info("Google OAuth2 client inicializado — coleta GSC/GA4 ativa")
		}
	}

	// ── Bing Service (opcional) ─────────────────────────────────
	var bingService *service.BingService
	if cfg.BingAPIKey != "" {
		bingService = service.NewBingService(cfg.BingAPIKey)
		slog.Info("Bing Webmaster service inicializado")
	}

	// ── Collector ───────────────────────────────────────────────
	collector := service.NewCollector(gscService, ga4Service, bingService, trafficRepo, clientRepo)

	// Sync inicial (se habilitado)
	if cfg.SyncOnStartup {
		go func() {
			count, err := sheetsService.SyncClients()
			if err != nil {
				slog.Error("falha no sync inicial de clientes", "err", err)
			} else {
				slog.Info("sync inicial completo", "clientes", count)
			}

			// Iniciar rotina de pre-warming (Primeiro Povoamento + Cron Diário)
			preWarmer := service.NewPreWarmer(collector, clientRepo, trafficRepo)
			preWarmer.Run(ctx) // vai bloquear a goroutine, mas é exatamente o que queremos (loop infinito de cron)
		}()
	}

	// Handlers
	clientsHandler := handler.NewClientsHandler(clientRepo, trafficRepo)
	trafficHandler := handler.NewTrafficHandler(clientRepo, trafficRepo, collector)
	syncHandler := handler.NewSyncHandler(sheetsService)

	// Router
	r := chi.NewRouter()

	// Middleware global
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Compress(5))
	r.Use(middleware.SecurityHeaders)
	
	// Basic Auth para proteger a aplicação inteira
	r.Use(chimiddleware.BasicAuth("Painel Restrito", map[string]string{
		cfg.ApiUser: cfg.ApiPass,
	}))

	// CORS — restrito ao frontend
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.CORSOrigin},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Accept"},
		ExposedHeaders:   []string{"Content-Disposition"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Rate limiting — 60 req/min por IP
	r.Use(httprate.LimitByIP(60, time.Minute))

	// Rotas API
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", handler.Health())
		r.Get("/clients", clientsHandler.List)
		r.Get("/traffic/{id}", trafficHandler.GetTraffic)
		r.Post("/sync/clients", syncHandler.SyncClients)
	})

	// ── Frontend SPA (embutido no binário) ─────────────────────
	r.Handle("/*", web.NewSPAHandler())

	// Start
	addr := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("servidor iniciando", "addr", addr, "cors", cfg.CORSOrigin)

	server := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second, // Reduzido para 15s (trabalho pesado movido para worker)
		IdleTimeout:  120 * time.Second,
	}

	// Inicia o servidor em uma goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("erro no servidor HTTP", "err", err)
			os.Exit(1)
		}
	}()

	// Aguarda sinal de encerramento (SIGINT, SIGTERM)
	<-ctx.Done()
	slog.Info("iniciando graceful shutdown do servidor...")

	// Cria um contexto com timeout para o Shutdown finalizar conexões ativas
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("erro ao encerrar servidor graciosamente", "err", err)
	} else {
		slog.Info("servidor encerrado com sucesso")
	}
}
