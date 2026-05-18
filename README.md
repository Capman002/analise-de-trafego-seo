# Análise de Tráfego SEO 🚀

![License](https://img.shields.io/badge/License-MIT%20+%20Commons%20Clause-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go)
![SvelteKit](https://img.shields.io/badge/SvelteKit-2.0-FF3E00?logo=svelte)

Um Dashboard de inteligência e consolidação de tráfego orgânico. O sistema coleta, analisa e armazena dados provenientes do **Google Search Console**, **Google Analytics 4** e **Bing Webmaster Tools**, centralizando o desempenho de múltiplos clientes em uma única interface moderna e de alta performance.

---

## 🏛️ Arquitetura

Este projeto foi construído sob um padrão pragmático de **Monolito Modular (Single Binary Deployment)**:
- **Backend:** Escrito em Go (chi router). Gerencia o loop de background (Worker), a orquestração de APIs externas (Google/Bing) e as rotas HTTP locais de forma assíncrona.
- **Frontend:** Um SPA construído em SvelteKit + Svelte 5, que é compilado e embutido diretamente no executável do Go utilizando a diretiva `//go:embed`.
- **Banco de Dados:** SQLite em modo WAL (Write-Ahead Logging). Extremamente rápido e dispensa a necessidade de orquestrar bancos de dados pesados, simplificando o deploy.

---

## 🔒 Segurança & Privacidade

Por lidar com dados analíticos confidenciais, o projeto conta com:
- **Zero Trust Local:** Todas as requisições API e o acesso à interface são protegidos pelo protocolo de **Basic Auth**.
- **Segregação de Secrets:** As credenciais e chaves do OAuth2 são repassadas via JSON injetado nas variáveis de ambiente.
- **Graceful Shutdown:** O sistema captura sinais do sistema operacional (SIGTERM/SIGINT) para concluir tarefas de inserção em background de forma segura, prevenindo corrupção no SQLite.

---

## ⚙️ Configuração do Ambiente (.env)

Copie o `.env.example` para `.env` e preencha as variáveis de acordo.

| Variável | Descrição |
|---|---|
| `PORT` | Porta de acesso do servidor HTTP (Ex: `8080`) |
| `DB_PATH` | Caminho do arquivo SQLite (Ex: `./data/analise-trafego.db`) |
| `API_USER` | Usuário do Basic Auth para acessar a interface e API |
| `API_PASS` | Senha do Basic Auth para acessar a interface e API |
| `SHEETS_CSV_URL` | URL pública da planilha do Google Sheets que contém os clientes |
| `GOOGLE_CREDENTIALS_JSON` | O JSON inteiro da Conta de Serviço (Service Account) gerada no Google Cloud Console |
| `BING_API_KEY` | Chave de API do Bing Webmaster Tools |

---

## 🐳 Deploy com Docker (Recomendado)

A forma mais fácil de rodar o projeto em produção é através do Docker Compose, pois o Dockerfile constrói as duas etapas (Bun/Svelte e Go) em uma imagem enxuta baseada em Alpine.

1. Clone o repositório:
   ```bash
   git clone https://github.com/SEU_USUARIO/analise-de-trafego-seo.git
   cd analise-de-trafego-seo
   ```
2. Crie e configure seu arquivo `.env`.
3. Suba o container (a persistência ocorre na pasta `/data` montada no host):
   ```bash
   docker compose up --build -d
   ```
4. **Acesse** em `http://localhost:8080` (Aguarde os prompts do Basic Auth).

> **Aviso GSC/GA4:** Como o sistema utiliza Conta de Serviço (Service Account), não se esqueça de adicionar o e-mail gerado (ex: `bot-seo@seu-projeto.iam.gserviceaccount.com`) como "Usuário com permissão de leitura" no Google Search Console e GA4 de cada cliente.

---

## 💻 Desenvolvimento Local

Caso queira modificar o projeto:

**1. Frontend (Hot Reload):**
```bash
cd frontend
bun install
bun run dev
```
*(O SvelteKit rodará na porta 5173. Configure a variável `CORS_ORIGIN=http://localhost:5173` no `.env` para o backend permitir a conexão)*

**2. Backend (API):**
```bash
cd backend
go run cmd/server/main.go
```

---

## ⚖️ Licença

Este projeto é disponibilizado sob a licença **MIT com Commons Clause (Restrição de Revenda/SaaS)**.

**O que isso significa de forma prática?**
- ✅ **Você (ou sua Agência) PODE:** Fazer o download, instalar, modificar o código e usar a ferramenta livremente em sua operação interna para gerar relatórios e prestar consultoria de SEO para seus clientes pagantes.
- ❌ **Você NÃO PODE:** Vender cópias deste software, vendê-lo como um produto (SaaS) ou criar plataformas *white-label* onde os clientes pagam pelo acesso direito à aplicação. O código é aberto para empoderar a comunidade técnica, mas a comercialização direta do código ou sistema é proibida.

Para mais detalhes jurídicos, consulte o arquivo `LICENSE` na raiz do projeto.
