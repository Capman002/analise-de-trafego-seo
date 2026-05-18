package web

import (
	"io/fs"
	"net/http"
	"strings"
)

// NewSPAHandler retorna um http.Handler que serve os arquivos estáticos
// do frontend embutido, com fallback SPA para index.html.
//
// Lógica:
//  1. Se o path começa com /api/, não intercepta (deixa chi resolver)
//  2. Tenta servir arquivo estático exato (JS, CSS, imagens, etc.)
//  3. Se o arquivo não existe, serve index.html (SPA client-side routing)
func NewSPAHandler() http.Handler {
	// Sub-FS para remover o prefixo "dist/" do embed
	distFS, err := fs.Sub(DistFS, "dist")
	if err != nil {
		panic("web: falha ao acessar dist/ embutido: " + err.Error())
	}

	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Nunca interceptar rotas API
		if strings.HasPrefix(path, "/api/") || path == "/api" {
			http.NotFound(w, r)
			return
		}

		// Limpar o path (remover leading slash para fs.Open)
		cleanPath := strings.TrimPrefix(path, "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}

		// Tentar abrir o arquivo no FS embutido
		f, err := distFS.(fs.ReadFileFS).ReadFile(cleanPath)
		if err == nil && f != nil {
			// Arquivo existe — servir normalmente (com content-type, cache, etc.)
			fileServer.ServeHTTP(w, r)
			return
		}

		// Arquivo não encontrado — SPA fallback: servir index.html
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
