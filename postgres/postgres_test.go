package postgres_test

import (
	"context"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	pgContainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	dbname   = "testdb"
	dbuser   = "testuser"
	password = "testsuperpassword"

	migrationFolder = "../migrations"
)

func TestContainerConnection(t *testing.T) {
	gormDB, err := setupTestDatabase(t)
	require.NotNil(t, gormDB)
	require.NoError(t, err)
}

func setupTestDatabase(t *testing.T) (*gorm.DB, error) {
	ctx := context.Background()
	ctr, err := pgContainer.Run(
		ctx,
		"postgres:17-alpine",
		pgContainer.WithDatabase(dbname),
		pgContainer.WithUsername(dbuser),
		pgContainer.WithPassword(password),
		pgContainer.BasicWaitStrategies(),
	)
	testcontainers.CleanupContainer(t, ctr)
	if err != nil {
		return nil, err
	}
	dbURL, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		TranslateError: true,
	})
	db = db.Debug()
	if err != nil {
		return nil, err
	}
	dbConn, err := db.DB()
	if err != nil {
		return nil, err
	}
	return db, goose.Up(dbConn, migrationFolder)
}
