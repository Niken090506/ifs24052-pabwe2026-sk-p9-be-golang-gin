package database

import (
	"testing"

	"app/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseSQLiteAndSeed(t *testing.T) {
	cfg := config.LoadConfig()
	cfg.DBConnection = "sqlite"
	db, err := Connect(cfg)
	assert.NoError(t, err)
	assert.NoError(t, Migrate(db))
	assert.NoError(t, Seed(db, cfg))
	assert.NoError(t, Seed(db, cfg))

	// Error branch when tables dropped
	_ = db.Migrator().DropTable("users")
	assert.Error(t, Seed(db, cfg))

	// Postgres Connect branch (expected connection error on invalid port)
	cfg.DBConnection = "pgsql"
	cfg.DBPort = "1"
	_, _ = Connect(cfg)
}
