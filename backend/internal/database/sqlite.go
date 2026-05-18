package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Open abre (ou cria) o banco SQLite no caminho especificado.
// Aplica WAL mode para melhor concorrência de leitura e
// foreign keys para integridade referencial.
func Open(dbPath string) (*sql.DB, error) {
	// Garante que o diretório existe
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("falha ao criar diretório do banco: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir SQLite: %w", err)
	}

	// Pragmas de performance e segurança
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=-64000", // 64MB cache
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return nil, fmt.Errorf("falha ao aplicar pragma %q: %w", p, err)
		}
	}

	// Pool de conexões — SQLite é single-writer, mas multi-reader com WAL
	db.SetMaxOpenConns(1) // Evita SQLITE_BUSY em writes concorrentes
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0) // Sem expiração

	if err := Migrate(db); err != nil {
		return nil, fmt.Errorf("falha na migração: %w", err)
	}

	return db, nil
}
