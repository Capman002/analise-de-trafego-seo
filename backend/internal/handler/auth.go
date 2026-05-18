package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wicomm/analise-trafego/internal/auth"
	"github.com/wicomm/analise-trafego/internal/repository"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	settingsRepo *repository.SettingsRepo
	credJSON     string
}

func NewAuthHandler(settingsRepo *repository.SettingsRepo, credJSON string) *AuthHandler {
	return &AuthHandler{
		settingsRepo: settingsRepo,
		credJSON:     credJSON,
	}
}

// Login redireciona o admin para a tela de consentimento do Google
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if h.credJSON == "" {
		http.Error(w, "GOOGLE_CREDENTIALS_JSON não configurado no servidor", http.StatusInternalServerError)
		return
	}

	config, err := auth.GetOAuthConfig(h.credJSON)
	if err != nil {
		http.Error(w, "Erro ao carregar configurações OAuth: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Usamos a URL base do Request para adivinhar a URL de callback dinâmica
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := r.Host
	// Se estiver atrás de um proxy, tenta pegar o host real
	if xfh := r.Header.Get("X-Forwarded-Host"); xfh != "" {
		host = xfh
	}

	config.RedirectURL = fmt.Sprintf("%s://%s/api/auth/google/callback", scheme, host)

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// Callback recebe o código do Google e salva o Token no banco de dados
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Código não encontrado na resposta do Google", http.StatusBadRequest)
		return
	}

	config, err := auth.GetOAuthConfig(h.credJSON)
	if err != nil {
		http.Error(w, "Erro ao carregar configurações OAuth", http.StatusInternalServerError)
		return
	}

	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := r.Host
	if xfh := r.Header.Get("X-Forwarded-Host"); xfh != "" {
		host = xfh
	}
	config.RedirectURL = fmt.Sprintf("%s://%s/api/auth/google/callback", scheme, host)

	tok, err := config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Falha ao trocar código pelo token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Monta o JSON mínimo
	tokenData := map[string]string{
		"access_token":  tok.AccessToken,
		"refresh_token": tok.RefreshToken,
		"token_type":    tok.TokenType,
	}
	tokenJSON, _ := json.Marshal(tokenData)

	// Salva no banco de dados
	if err := h.settingsRepo.Set("google_token_json", string(tokenJSON)); err != nil {
		http.Error(w, "Falha ao salvar token no banco de dados: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Tenta atualizar o Singleton HTTP Client na memória para passar a coletar dados instantaneamente
	if _, err := auth.UpdateGoogleClient(h.credJSON, string(tokenJSON)); err != nil {
		fmt.Printf("[ERRO] Falha ao injetar novo token em memória: %v\n", err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`
		<html>
		<body style="font-family: sans-serif; text-align: center; padding: 50px;">
			<h1 style="color: #4CAF50;">✅ Autenticação Concluída com Sucesso!</h1>
			<p>O token do Google foi salvo com segurança no banco de dados.</p>
			<p>O sistema já tem permissão para ler os dados do GSC e GA4.</p>
			<button onclick="window.location.href='/'" style="padding: 10px 20px; font-size: 16px; cursor: pointer;">Voltar para o Dashboard</button>
		</body>
		</html>
	`))
}
