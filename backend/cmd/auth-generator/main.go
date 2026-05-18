package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// googleCreds representa o JSON de credenciais OAuth2 do Google.
type googleCreds struct {
	Installed *oauthApp `json:"installed"`
	Web       *oauthApp `json:"web"`
}

type oauthApp struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// googleToken para imprimir no final
type googleToken struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
}

func main() {
	// 1. Carregar .env
	_ = godotenv.Load("../../.env")
	if os.Getenv("GOOGLE_CREDENTIALS_JSON") == "" {
		_ = godotenv.Load(".env")
	}

	credJSON := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	if credJSON == "" {
		log.Fatal("ERRO: Variável GOOGLE_CREDENTIALS_JSON não encontrada no ambiente ou .env")
	}

	// 2. Fazer parse do JSON de credenciais
	var creds googleCreds
	if err := json.Unmarshal([]byte(credJSON), &creds); err != nil {
		log.Fatalf("Erro ao parsear credenciais: %v", err)
	}

	app := creds.Installed
	if app == nil {
		app = creds.Web
	}
	if app == nil {
		log.Fatal("Credenciais não contêm 'installed' nem 'web'")
	}

	// 3. Configurar OAuth2 com Redirect URI apontando para o servidor local
	config := &oauth2.Config{
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:9999/callback", // Requisito: Cadastrar no GCP
		Scopes: []string{
			"https://www.googleapis.com/auth/webmasters.readonly",
			"https://www.googleapis.com/auth/analytics.readonly",
		},
	}

	// Canal para avisar a main thread que terminamos
	done := make(chan bool)

	// 4. Criar o servidor HTTP
	server := &http.Server{Addr: ":9999"}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Código não encontrado", http.StatusBadRequest)
			return
		}

		// Trocar código por token
		tok, err := config.Exchange(context.TODO(), code)
		if err != nil {
			http.Error(w, "Falha ao trocar código: "+err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "<h1>Autenticação Concluída!</h1><p>Você pode fechar esta aba e voltar ao terminal.</p>")

		// Montar e imprimir o JSON mágico
		finalToken := googleToken{
			RefreshToken: tok.RefreshToken,
			AccessToken:  tok.AccessToken,
			TokenType:    tok.TokenType,
		}

		tokenBytes, _ := json.Marshal(finalToken)
		fmt.Println("\n=========================================================================")
		fmt.Println("SUCESSO! Copie o JSON abaixo e cole no seu .env como GOOGLE_TOKEN_JSON:")
		fmt.Println("=========================================================================\n")
		fmt.Printf("GOOGLE_TOKEN_JSON='%s'\n\n", string(tokenBytes))
		fmt.Println("=========================================================================")

		// Avisa que terminou para encerrar o servidor
		done <- true
	})

	// 5. Iniciar Servidor e Orientar Usuário
	fmt.Println("Iniciando servidor de autenticação local...")
	fmt.Println("IMPORTANTE: Certifique-se de que 'http://localhost:9999/callback' está cadastrado como uma URI de redirecionamento autorizada no seu Google Cloud Console.")
	fmt.Println("")

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Printf("👉 CLIQUE NO LINK ABAIXO PARA AUTENTICAR:\n\n%v\n\n", authURL)
	fmt.Println("Aguardando aprovação no navegador...")

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro no servidor: %v", err)
		}
	}()

	// Bloqueia até o callback processar tudo
	<-done
	server.Shutdown(context.TODO())
}
