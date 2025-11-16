package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"PrService/src/internal/infrastructure/data/repositories"

	"PrService/src/internal/application/contracts"
	"PrService/src/internal/application/services"
	"PrService/src/internal/domain"
	"PrService/src/internal/http_api/controllers"
	"PrService/src/internal/http_api/middlewares"

	"PrService/src/cmd/config"
	"PrService/src/internal/infrastructure/data"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "PrService/src/internal/http_api/swagger"
)

type App struct {
	cfg    *config.Config
	logger *slog.Logger
	pool   *pgxpool.Pool
	server *http.Server
}

func NewApp(cfg *config.Config) (*App, error) {
	logger, err := initLogger(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	pool, err := initPgPool(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("init pg pool: %w", err)
	}

	prRepo, teamRepo, userRepo := initRepositories(pool)
	txManager := data.NewTxManager(pool)
	prService, teamService, userService := initServices(prRepo, teamRepo, userRepo, txManager)
	validate := validator.New()
	prController, teamController, userController, healthController := initControllers(
		prService,
		teamService,
		userService,
		validate,
		logger,
	)
	server := initServer(prController, teamController, userController, healthController, logger, cfg.HTTPPort)

	return &App{
		cfg:    cfg,
		logger: logger,
		pool:   pool,
		server: server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.logger.Info("starting server", "addr", "http://localhost:"+a.cfg.HTTPPort)

	errCh := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")
	case err := <-errCh:
		a.logger.Error("server error", "err", err)
		return err
	}

	withOutCancelCtx := context.WithoutCancel(ctx)
	shutdownCtx, cancel := context.WithTimeout(withOutCancelCtx, 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("server shutdown error", "err", err)
	}
	a.logger.Info("server gracefully stopped")

	a.pool.Close()
	a.logger.Info("pgx pool closed")

	return nil
}

func initServer(
	prController *controllers.PullRequestController,
	teamController *controllers.TeamController,
	userController *controllers.UserController,
	healthController *controllers.HealthController,
	logger *slog.Logger,
	port string,
) *http.Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	requestLogMiddleware := middlewares.NewRequestLogger(logger)
	r.Use(requestLogMiddleware.LogRequest)
	prController.UseHandlers(r)
	teamController.UseHandlers(r)
	userController.UseHandlers(r)
	healthController.UseHandlers(r)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	server := &http.Server{
		Addr: ":" + port,
	}
	server.Handler = r

	return server
}

func initControllers(
	prService domain.PullRequestService,
	teamService domain.TeamService,
	userService domain.UserService,
	validate *validator.Validate,
	logger *slog.Logger,
) (
	*controllers.PullRequestController,
	*controllers.TeamController,
	*controllers.UserController,
	*controllers.HealthController,
) {
	prController := controllers.NewPullRequestController(prService, validate, logger)
	teamController := controllers.NewTeamController(teamService, validate, logger)
	userController := controllers.NewUserController(userService, validate, logger)
	healthController := controllers.NewHealthController(validate, logger)
	return prController, teamController, userController, healthController
}

func initServices(
	prRepo domain.PullRequestRepository,
	teamRepo domain.TeamRepository,
	userRepo domain.UserRepository,
	txManager contracts.TxManager,
) (domain.PullRequestService, domain.TeamService, domain.UserService) {
	prService := services.NewPullRequestService(prRepo, teamRepo, txManager)
	teamService := services.NewTeamService(teamRepo, userRepo, txManager)
	userService := services.NewUserService(userRepo, prRepo)

	return prService, teamService, userService
}

func initRepositories(pool *pgxpool.Pool) (
	domain.PullRequestRepository,
	domain.TeamRepository,
	domain.UserRepository,
) {
	pr := repositories.NewPullRequestRepository(pool)
	team := repositories.NewTeamRepository(pool)
	user := repositories.NewUserRepository(pool)

	return pr, team, user
}

func initPgPool(cfg config.DBConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)

	pgCfg := data.PostgresConfig{
		DSN:               dsn,
		MaxConns:          cfg.MaxConns,
		MinConns:          cfg.MinConns,
		MaxConnLifetime:   cfg.MaxConnLifetime,
		MaxConnIdleTime:   cfg.MaxConnIdleTime,
		HealthCheckPeriod: cfg.HealthCheckPeriod,
	}

	pool, err := data.NewPgxPool(context.Background(), pgCfg)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func initLogger(levelStr, format string) (*slog.Logger, error) {
	level, err := parseLevel(levelStr)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	switch strings.ToLower(format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text", "console", "":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		return nil, fmt.Errorf("unsupported log format: %s", format)
	}

	logger := slog.New(handler)

	return logger, nil
}

func parseLevel(levelStr string) (slog.Level, error) {
	switch strings.ToUpper(strings.TrimSpace(levelStr)) {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO", "":
		return slog.LevelInfo, nil
	case "WARN", "WARNING":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log level: %s", levelStr)
	}
}
