package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wicomm/analise-trafego/internal/auth"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	credJSON string
}

func NewAuthHandler(credJSON string) *AuthHandler {
	return &AuthHandler{
		credJSON: credJSON,
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

	html := fmt.Sprintf(`
		<html>
		<body style="font-family: sans-serif; text-align: center; padding: 50px; background: #f9fafb;">
			<div style="background: white; padding: 40px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); max-width: 800px; margin: 0 auto;">
				<h1 style="color: #10b981; margin-bottom: 20px;">✅ Autenticação Concluída com Sucesso!</h1>
				<p style="color: #374151; font-size: 16px; margin-bottom: 30px;">
					Copie o código abaixo e cole na aba <b>Environment Variables</b> do seu Dokploy,<br>
					junto com as outras variáveis, e clique em <b>Deploy/Restart</b>.
				</p>
				
				<div style="background: #1f2937; color: #10b981; padding: 20px; border-radius: 8px; text-align: left; overflow-x: auto; margin-bottom: 20px; position: relative;">
					<code id="tokenCode" style="font-family: monospace; font-size: 14px; word-break: break-all;">GOOGLE_TOKEN_JSON='%s'</code>
				</div>
				
				<button onclick="copyToClipboard()" id="copyBtn" style="background: #3b82f6; color: white; border: none; padding: 12px 24px; border-radius: 6px; font-size: 16px; cursor: pointer; font-weight: bold; transition: background 0.2s;">
					Copiar Código
				</button>
			</div>

			<script>
				function copyToClipboard() {
					var code = document.getElementById("tokenCode").innerText;
					navigator.clipboard.writeText(code).then(function() {
						var btn = document.getElementById("copyBtn");
						btn.innerText = "Copiado!";
						btn.style.background = "#10b981";
						setTimeout(function() {
							btn.innerText = "Copiar Código";
							btn.style.background = "#3b82f6";
						}, 3000);
					});
				}
			</script>
		</body>
		</html>
	`, string(tokenJSON))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
