package web

import "embed"

// DistFS contém todos os assets estáticos do frontend SvelteKit,
// buildados via `bun run build` e copiados para este diretório.
//
// Em desenvolvimento, o diretório `dist/` pode estar vazio —
// o build do Dockerfile (ou um script local) popula-o antes de `go build`.
//
//go:embed all:dist
var DistFS embed.FS
