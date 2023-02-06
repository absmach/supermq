package testsutil

import (
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	"github.com/stretchr/testify/require"
)

func GenerateUUID(t *testing.T, idProvider mainflux.IDProvider) string {
	ulid, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	return ulid
}

func CleanUpDB(t *testing.T, db *sqlx.DB) {
	_, err := db.Exec("DELETE FROM policies")
	require.Nil(t, err, fmt.Sprintf("clean policies unexpected error: %s", err))
	_, err = db.Exec("DELETE FROM groups")
	require.Nil(t, err, fmt.Sprintf("clean groups unexpected error: %s", err))
	_, err = db.Exec("DELETE FROM clients")
	require.Nil(t, err, fmt.Sprintf("clean clients unexpected error: %s", err))
}
