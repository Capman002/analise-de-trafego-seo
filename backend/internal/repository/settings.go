package repository

import (
	"database/sql"
	"fmt"
)

type SettingsRepo struct {
	db *sql.DB
}

func NewSettingsRepo(db *sql.DB) *SettingsRepo {
	return &SettingsRepo{db: db}
}

func (r *SettingsRepo) Get(key string) (string, error) {
	var value string
	err := r.db.QueryRow("SELECT value FROM system_settings WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil // Chave não existe, retorna string vazia sem erro
	}
	if err != nil {
		return "", fmt.Errorf("erro ao buscar configuração %q: %w", key, err)
	}
	return value, nil
}

func (r *SettingsRepo) Set(key, value string) error {
	_, err := r.db.Exec(`
		INSERT INTO system_settings (key, value) 
		VALUES (?, ?) 
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	if err != nil {
		return fmt.Errorf("erro ao salvar configuração %q: %w", key, err)
	}
	return nil
}
