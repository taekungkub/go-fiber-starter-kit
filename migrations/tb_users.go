package migrations

import "github.com/jmoiron/sqlx"

func CreateTableUsers() string {
	return `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email VARCHAR(255) NOT NULL UNIQUE,
		username VARCHAR(100) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL,
		first_name VARCHAR(100) NOT NULL DEFAULT '',
		last_name VARCHAR(100) NOT NULL DEFAULT '',
		name VARCHAR(200) GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED,
		role user_role NOT NULL DEFAULT 'CUSTOMER',
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		refresh_token_hash VARCHAR(255),
		telephone VARCHAR(20),
		deleted_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`
}

func MigrateTableUsers(db *sqlx.DB) error {
	_, err := db.Exec(CreateTableUsers())
	if err != nil {
		return err
	}

	return nil
}
