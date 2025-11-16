//go:build integration

package integration_tests

import (
	"PrService/src/internal/domain"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testPool *pgxpool.Pool

const migrationFile = "../migrations/0001_create_schema.up.sql"

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start postgres container: %v\n", err)
		os.Exit(1)
	}

	cleanup := func() {
		if testPool != nil {
			testPool.Close()
		}
		_ = pgContainer.Terminate(context.Background())
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get postgres connection string: %v\n", err)
		cleanup()
		os.Exit(1)
	}

	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create pgxpool: %v\n", err)
		cleanup()
		os.Exit(1)
	}

	if err := applyMigration(ctx, testPool, migrationFile); err != nil {
		fmt.Fprintf(os.Stderr, "failed to apply migration: %v\n", err)
		cleanup()
		os.Exit(1)
	}

	code := m.Run()

	cleanup()
	os.Exit(code)
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, path string) error {
	migrationPath := filepath.FromSlash(path)

	sqlBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("read migration file %s: %w", migrationPath, err)
	}

	if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("execute migration %s: %w", migrationPath, err)
	}

	return nil
}

func truncateAll(t *testing.T, ctx context.Context) {
	t.Helper()

	_, err := testPool.Exec(ctx, `
		TRUNCATE TABLE pull_request_reviewers, pull_requests, users, teams
		RESTART IDENTITY CASCADE;
	`)
	if err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}
}

func insertTeam(t *testing.T, ctx context.Context, name domain.TeamName) {
	t.Helper()

	_, err := testPool.Exec(ctx, `INSERT INTO teams (name) VALUES ($1)`, name)
	if err != nil {
		t.Fatalf("failed to insert team %s: %v", name, err)
	}
}

func insertUser(t *testing.T, ctx context.Context, u domain.User) {
	t.Helper()

	_, err := testPool.Exec(ctx, `
		INSERT INTO users (id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
	`, u.ID, u.Username, u.TeamName, u.IsActive)
	if err != nil {
		t.Fatalf("failed to insert user %s: %v", u.ID, err)
	}
}
