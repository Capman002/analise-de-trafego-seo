package auth

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	clientOnce sync.Once
	cachedHTTP *http.Client
	clientErr  error
)

// NewGoogleClient cria um http.Client autenticado via OAuth2/Service Account para APIs Google.
// O client é singleton — chamadas subsequentes retornam o mesmo client.
func NewGoogleClient(credJSON string) (*http.Client, error) {
	clientOnce.Do(func() {
		cachedHTTP, clientErr = buildClient(credJSON)
	})
	return cachedHTTP, clientErr
}

func buildClient(credJSON string) (*http.Client, error) {
	if credJSON == "" {
		return nil, fmt.Errorf("credenciais google vazias")
	}

	scopes := []string{
		"https://www.googleapis.com/auth/webmasters.readonly",
		"https://www.googleapis.com/auth/analytics.readonly",
	}

	creds, err := google.CredentialsFromJSON(context.Background(), []byte(credJSON), scopes...)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler credenciais do JSON (Service Account/OAuth): %w", err)
	}

	// Retorna um client HTTP que injeta o Bearer token automaticamente
	return oauth2.NewClient(context.Background(), creds.TokenSource), nil
}
