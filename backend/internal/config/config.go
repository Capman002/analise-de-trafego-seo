package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config centraliza todas as variáveis de ambiente do backend.
// Nenhum valor sensível é hardcoded — tudo vem de .env.
type Config struct {
	Port       int
	DBPath     string
	CORSOrigin string

	// Google Sheets (público — sem auth)
	SheetsCSVURL string

	// Google API — credenciais em formato JSON (Service Account)
	GoogleCredentialsJSON string

	// Bing Webmaster Tools
	BingAPIKey string

	// Basic Auth
	ApiUser string
	ApiPass string

	// Sync
	SyncOnStartup bool
}

func Load() (*Config, error) {
	// Tenta carregar .env da pasta atual ou da pasta pai (para dev local)
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	port, _ := strconv.Atoi(getEnv("PORT", "8080"))

	cfg := &Config{
		Port:                  port,
		DBPath:                getEnv("DB_PATH", "./data/analise-trafego.db"),
		CORSOrigin:            getEnv("CORS_ORIGIN", "http://localhost:5173"),
		SheetsCSVURL:          os.Getenv("SHEETS_CSV_URL"),
		GoogleCredentialsJSON: os.Getenv("GOOGLE_CREDENTIALS_JSON"),
		BingAPIKey:            os.Getenv("BING_API_KEY"),
		ApiUser:               getEnv("API_USER", "admin"),
		ApiPass:               getEnv("API_PASS", "admin"),
		SyncOnStartup:         getEnv("SYNC_ON_STARTUP", "true") == "true",
	}

	if cfg.SheetsCSVURL == "" {
		return nil, fmt.Errorf("SHEETS_CSV_URL é obrigatório — configure no .env")
	}

	if cfg.GoogleCredentialsJSON == "" {
		// Warning, não fatal — permite testar sync de clientes sem credenciais Google
		fmt.Println("[WARN] GOOGLE_CREDENTIALS_JSON não configurado — coleta GSC/GA4 indisponível")
	}

	if cfg.BingAPIKey == "" {
		fmt.Println("[WARN] BING_API_KEY não configurado — coleta Bing indisponível")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
