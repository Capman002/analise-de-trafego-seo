package repository

import (
	"database/sql"

	"github.com/wicomm/analise-trafego/internal/model"
)

// ClientRepo gerencia operações CRUD sobre a tabela clients.
type ClientRepo struct {
	db *sql.DB
}

func NewClientRepo(db *sql.DB) *ClientRepo {
	return &ClientRepo{db: db}
}

// ListAll retorna todos os clientes ordenados por nome.
func (r *ClientRepo) ListAll() ([]model.Client, error) {
	rows, err := r.db.Query(`
		SELECT id, name, COALESCE(gsc_url,''), COALESCE(gsc_type,''),
		       COALESCE(ga4_id,''), COALESCE(brand_regex,''), synced_at
		FROM clients ORDER BY name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []model.Client
	for rows.Next() {
		var c model.Client
		if err := rows.Scan(&c.ID, &c.Name, &c.GscURL, &c.GscType, &c.Ga4ID, &c.BrandRegex, &c.SyncedAt); err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return clients, rows.Err()
}

// GetByID retorna um cliente pelo ID.
func (r *ClientRepo) GetByID(id int64) (*model.Client, error) {
	var c model.Client
	err := r.db.QueryRow(`
		SELECT id, name, COALESCE(gsc_url,''), COALESCE(gsc_type,''),
		       COALESCE(ga4_id,''), COALESCE(brand_regex,''), synced_at
		FROM clients WHERE id = ?
	`, id).Scan(&c.ID, &c.Name, &c.GscURL, &c.GscType, &c.Ga4ID, &c.BrandRegex, &c.SyncedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Upsert insere ou atualiza um cliente pelo nome (chave natural).
// Usado no sync da planilha — preserva o ID se já existir.
func (r *ClientRepo) Upsert(c model.Client) error {
	_, err := r.db.Exec(`
		INSERT INTO clients (name, gsc_url, gsc_type, ga4_id, brand_regex, synced_at)
		VALUES (?, ?, NULLIF(?,''), NULLIF(?,''), NULLIF(?,''), CURRENT_TIMESTAMP)
		ON CONFLICT(name) DO UPDATE SET
			gsc_url     = EXCLUDED.gsc_url,
			gsc_type    = EXCLUDED.gsc_type,
			ga4_id      = EXCLUDED.ga4_id,
			brand_regex = EXCLUDED.brand_regex,
			synced_at   = CURRENT_TIMESTAMP
	`, c.Name, c.GscURL, c.GscType, c.Ga4ID, c.BrandRegex)
	return err
}
