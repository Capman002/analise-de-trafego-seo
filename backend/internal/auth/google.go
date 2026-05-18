package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

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

// googleToken representa o JSON do token OAuth2.
type googleToken struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
}

var (
	clientOnce sync.Once
	cachedHTTP *http.Client
	clientErr  error
)

// NewGoogleClient cria um http.Client autenticado via OAuth2 para APIs Google.
// Usa os JSONs de credenciais e token no formato gerado pelo OAuth2 flow.
// O client é singleton — chamadas subsequentes retornam o mesmo client.
func NewGoogleClient(credJSON, tokenJSON string) (*http.Client, error) {
	clientOnce.Do(func() {
		cachedHTTP, clientErr = buildClient(credJSON, tokenJSON)
	})
	return cachedHTTP, clientErr
}

func buildClient(credJSON, tokenJSON string) (*http.Client, error) {
	var creds googleCreds
	if err := json.Unmarshal([]byte(credJSON), &creds); err != nil {
		return nil, fmt.Errorf("erro ao parsear credenciais: %w", err)
	}

	app := creds.Installed
	if app == nil {
		app = creds.Web
	}
	if app == nil {
		return nil, fmt.Errorf("credenciais não contêm 'installed' nem 'web'")
	}

	var tok googleToken
	if err := json.Unmarshal([]byte(tokenJSON), &tok); err != nil {
		return nil, fmt.Errorf("erro ao parsear token: %w", err)
	}

	if tok.RefreshToken == "" {
		return nil, fmt.Errorf("refresh_token ausente no token JSON")
	}

	config := &oauth2.Config{
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			"https://www.googleapis.com/auth/webmasters.readonly",
			"https://www.googleapis.com/auth/analytics.readonly",
		},
	}

	oauthToken := &oauth2.Token{
		RefreshToken: tok.RefreshToken,
		TokenType:    "Bearer",
	}

	// O TokenSource renova o access_token automaticamente via refresh_token
	return config.Client(context.Background(), oauthToken), nil
}
