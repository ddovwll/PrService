package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host              string
	Port              int
	User              string
	Password          string
	Name              string
	SSLMode           string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

type Config struct {
	HTTPPort      string
	LogLevel      string
	LogFormat     string
	DB            DBConfig
	MigrationsDir string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}

	cfg := &Config{
		HTTPPort:  getEnv("HTTP_PORT", "8080"),
		LogLevel:  getEnv("LOG_LEVEL", "INFO"),
		LogFormat: getEnv("LOG_FORMAT", "text"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "postgres"),
			User:     getEnv("DB_USER", "user"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		MigrationsDir: getEnv("MIGRATIONS_DIR", "src/internal/infrastructure/data/migrations"),
	}

	if cfg.DB.Port, err = getEnvInt("DB_PORT", 5432); err != nil {
		return nil, fmt.Errorf("parse DB_PORT: %w", err)
	}

	var maxConns int
	if maxConns, err = getEnvInt("DB_MAX_CONNS", 10); err != nil {
		return nil, fmt.Errorf("parse DB_MAX_CONNS: %w", err)
	}
	cfg.DB.MaxConns = int32(maxConns)

	var minConns int
	if minConns, err = getEnvInt("DB_MIN_CONNS", 2); err != nil {
		return nil, fmt.Errorf("parse DB_MIN_CONNS: %w", err)
	}
	cfg.DB.MinConns = int32(minConns)

	if cfg.DB.MaxConnLifetime, err = getEnvDurationSeconds("MAX_CONN_LIFETIME", 1800); err != nil {
		return nil, fmt.Errorf("parse MAX_CONN_LIFETIME: %w", err)
	}
	if cfg.DB.MaxConnIdleTime, err = getEnvDurationSeconds("MAX_CONN_IDLE_TIME", 300); err != nil {
		return nil, fmt.Errorf("parse MAX_CONN_IDLE_TIME: %w", err)
	}
	if cfg.DB.HealthCheckPeriod, err = getEnvDurationSeconds("HEALTH_CHECK_PERIOD", 60); err != nil {
		return nil, fmt.Errorf("parse HEALTH_CHECK_PERIOD: %w", err)
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return def
}

func getEnvInt(key string, def int) (int, error) {
	valStr, ok := os.LookupEnv(key)
	if !ok || valStr == "" {
		return def, nil
	}
	v, err := strconv.Atoi(valStr)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func getEnvDurationSeconds(key string, defSeconds int) (time.Duration, error) {
	valStr, ok := os.LookupEnv(key)
	if !ok || valStr == "" {
		return time.Duration(defSeconds) * time.Second, nil
	}
	secs, err := strconv.Atoi(valStr)
	if err != nil {
		return 0, err
	}
	return time.Duration(secs) * time.Second, nil
}
