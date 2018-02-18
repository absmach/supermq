package postgres

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Connect creates a connection to the PostgreSQL instance. A non-nil error
// is returned to indicate failure.
func Connect(host, port, name, user, pass string) (*DB, error) {
	t := "host=%s port=%s user=%s dbname=%s password=%s"
	str := fmt.Sprintf(t, cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBName, cfg.DBPass)

	return gorm.Open("postgres", str)
}
