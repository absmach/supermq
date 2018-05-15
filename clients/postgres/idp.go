package postgres

import (
	"database/sql"

	"github.com/mainflux/mainflux/clients"
	"github.com/mainflux/mainflux/logger"
	uuid "github.com/satori/go.uuid"
)

var _ clients.IdentityProvider = (*identityProvider)(nil)

type identityProvider struct {
	db  *sql.DB
	log logger.Logger
}

// NewIdentityProvider instantiates a PostgreSQL implementation of identity
// provider.
func NewIdentityProvider(db *sql.DB, log logger.Logger) clients.IdentityProvider {
	return identityProvider{db, log}
}

func (ip identityProvider) Key() string {
	return uuid.NewV4().String()
}

func (ip identityProvider) Identity(key string) (string, error) {
	q := `SELECT id FROM clients WHERE key = $1`

	var id string
	if err := ip.db.QueryRow(q, key).Scan(&id); err != nil {
		return "", clients.ErrUnauthorizedAccess
	}

	return id, nil
}
