package database

import "database/sql"

// Migrate aplica o schema DDL de forma idempotente.
// Usa IF NOT EXISTS para ser seguro em execuções repetidas.
func Migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS clients (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		name        TEXT    NOT NULL UNIQUE,
		gsc_url     TEXT,
		gsc_type    TEXT,
		ga4_id      TEXT,
		brand_regex TEXT,
		synced_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT chk_type CHECK (gsc_type IN ('domain', 'url') OR gsc_type IS NULL)
	);

	CREATE TABLE IF NOT EXISTS gsc_data (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		client_id    INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
		date         TEXT    NOT NULL,
		dimension    TEXT    NOT NULL,
		key          TEXT    NOT NULL,
		clicks       INTEGER DEFAULT 0,
		impressions  INTEGER DEFAULT 0,
		ctr          REAL    DEFAULT 0,
		position     REAL    DEFAULT 0,
		fetched_at   DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Índice principal para buscas por cliente + data (range queries no gráfico)
	CREATE INDEX IF NOT EXISTS idx_gsc_client_date ON gsc_data(client_id, date);

	-- Índice UNIQUE para upsert (evita duplicatas)
	CREATE UNIQUE INDEX IF NOT EXISTS idx_gsc_unique ON gsc_data(client_id, date, dimension, key);

	-- Índice de cobertura para agregações por dimensão (tabelas de queries/pages).
	-- Otimiza GROUP BY key com filtro client_id + dimension + date range.
	-- Inclui clicks, impressions, position para evitar table lookups (covering index).
	CREATE INDEX IF NOT EXISTS idx_gsc_agg ON gsc_data(client_id, dimension, date, key, clicks, impressions, position);

	CREATE TABLE IF NOT EXISTS ga4_data (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		client_id       INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
		date            TEXT    NOT NULL,
		item_name       TEXT,
		item_id         TEXT,
		sessions        INTEGER DEFAULT 0,
		engaged_sessions INTEGER DEFAULT 0,
		conversions     INTEGER DEFAULT 0,
		revenue         REAL    DEFAULT 0,
		items_purchased INTEGER DEFAULT 0,
		fetched_at      DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_ga4_client_date ON ga4_data(client_id, date);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_ga4_unique ON ga4_data(client_id, date, item_name);

	CREATE TABLE IF NOT EXISTS bing_data (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		client_id    INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
		date         TEXT    NOT NULL,
		dimension    TEXT    NOT NULL,
		key          TEXT    NOT NULL,
		clicks       INTEGER DEFAULT 0,
		impressions  INTEGER DEFAULT 0,
		position     REAL    DEFAULT 0,
		fetched_at   DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_bing_client_date ON bing_data(client_id, date);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_bing_unique ON bing_data(client_id, date, dimension, key);

	CREATE TABLE IF NOT EXISTS sync_log (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		client_id  INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
		source     TEXT    NOT NULL,
		status     TEXT    NOT NULL,
		message    TEXT,
		started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		ended_at   DATETIME
	);

	CREATE TABLE IF NOT EXISTS client_sync_status (
		client_id  INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
		source     TEXT    NOT NULL,
		status     TEXT    NOT NULL,
		error_message TEXT,
		last_sync_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY(client_id, source)
	);
	`

	_, err := db.Exec(schema)
	return err
}
