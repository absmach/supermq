package postgres

import (
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // required for SQL access
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
)

// Config defines the options that are used when connecting to a PostgreSQL instance
type Config struct {
	Host        string
	Port        string
	User        string
	Pass        string
	Name        string
	SSLMode     string
	SSLCert     string
	SSLKey      string
	SSLRootCert string
}

// Connect creates a connection to the PostgreSQL instance and applies any
// unapplied database migrations. A non-nil error is returned to indicate
// failure.
func Connect(cfg Config) (*sqlx.DB, error) {
	url := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s", cfg.Host, cfg.Port, cfg.User, cfg.Name, cfg.Pass, cfg.SSLMode, cfg.SSLCert, cfg.SSLKey, cfg.SSLRootCert)
	db, err := sqlx.Open("pgx", url)
	if err != nil {
		return nil, err
	}

	if err := migrateDB(db); err != nil {
		return nil, err
	}
	return db, nil
}

func migrateDB(db *sqlx.DB) error {
	migrations := &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "clients_01",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS clients (
						id          VARCHAR(254) PRIMARY KEY,
						name        VARCHAR(1024),
						owner       VARCHAR(254),
						identity    VARCHAR(254),
						secret      VARCHAR(4096) UNIQUE NOT NULL,
						tags        TEXT[],
						metadata    JSONB,
						created_at  TIMESTAMP,
						updated_at  TIMESTAMP,
						status      SMALLINT NOT NULL CHECK (status >= 0) DEFAULT 1
					)`,
					`CREATE TABLE IF NOT EXISTS groups (
						id          VARCHAR(254) PRIMARY KEY,
						parent_id   VARCHAR(254),
						owner_id    VARCHAR(254) NOT NULL,
						name        VARCHAR(1024) NOT NULL,
						description VARCHAR(1024),
						metadata    JSONB,
						created_at  TIMESTAMP,
						updated_at  TIMESTAMP,
						status      SMALLINT NOT NULL CHECK (status >= 0) DEFAULT 1,
						UNIQUE (owner_id, name),
						FOREIGN KEY (parent_id) REFERENCES groups (id) ON DELETE CASCADE
					)`,
					`CREATE TABLE IF NOT EXISTS policies (
						owner_id    VARCHAR(254) NOT NULL,
						subject     VARCHAR(254) NOT NULL,
						object      VARCHAR(254) NOT NULL,
						actions     TEXT[] NOT NULL,
						created_at  TIMESTAMP,
						updated_at  TIMESTAMP,
						FOREIGN KEY (subject) REFERENCES clients (id) ON DELETE CASCADE ON UPDATE CASCADE,
						FOREIGN KEY (object) REFERENCES groups (id) ON DELETE CASCADE ON UPDATE CASCADE,
						PRIMARY KEY (subject, object, actions)
					)`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS clients`,
					`DROP TABLE IF EXISTS groups`,
					`DROP TABLE IF EXISTS policies`,
				},
			},
		},
	}

	_, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	return err
}
