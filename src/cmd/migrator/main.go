package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"PrService/src/cmd/config"
	"PrService/src/internal/infrastructure/data"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	if cfg.MigrationsDir == "" {
		fmt.Println("MIGRATIONS_DIR is empty in config/env")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)

	pgCfg := data.PostgresConfig{
		DSN:               dsn,
		MaxConns:          cfg.DB.MaxConns,
		MinConns:          cfg.DB.MinConns,
		MaxConnLifetime:   cfg.DB.MaxConnLifetime,
		MaxConnIdleTime:   cfg.DB.MaxConnIdleTime,
		HealthCheckPeriod: cfg.DB.HealthCheckPeriod,
	}

	pool, err := data.NewPgxPool(ctx, pgCfg)
	if err != nil {
		fmt.Println("failed to create pgx pool", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	fmt.Println("starting migrations", "dir", cfg.MigrationsDir)

	if err := runMigrations(ctx, pool, cfg.MigrationsDir); err != nil {
		fmt.Println("migrations failed", "err", err)
		os.Exit(1)
	}

	fmt.Println("migrations applied successfully")
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %s: %w", dir, err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".up.sql") {
			files = append(files, filepath.Join(dir, name))
		}
	}

	if len(files) == 0 {
		fmt.Println("no *.up.sql migrations found", "dir", dir)
		return nil
	}

	sort.Strings(files)

	for _, path := range files {
		fmt.Println("applying migration", "file", path)

		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", path, err)
		}

		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("execute migration %s: %w", path, err)
		}
	}

	return nil
}
