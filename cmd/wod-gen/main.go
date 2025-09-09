//nolint:gocritic // main function
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/LinaKACI-pro/wod-gen/internal/config"
	"github.com/LinaKACI-pro/wod-gen/internal/core"
	"github.com/LinaKACI-pro/wod-gen/internal/core/catalog"
	"github.com/LinaKACI-pro/wod-gen/internal/handlers"
	"github.com/LinaKACI-pro/wod-gen/internal/repository"
	"github.com/LinaKACI-pro/wod-gen/pkg"
	"github.com/LinaKACI-pro/wod-gen/pkg/obs"
	"github.com/caarlos0/env/v11"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	_ "github.com/lib/pq"
	ginvalidator "github.com/oapi-codegen/gin-middleware"
)

func main() {
	// Init config
	var cfg config.Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("env.Parse: %v", err)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	// init logger
	logger := pkg.NewLogger()

	logger.Info("boot",
		slog.Int("port", cfg.HTTP.Port),
		slog.Bool("obs", cfg.Obs.Enabled),
		slog.String("rate_strategy", cfg.RateLimit.Strategy),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// init Observability
	otelSetup, err := obs.Init(ctx, cfg.Obs.OTLPEndpoint, cfg.Obs.GrafanaToken)
	if err != nil {
		logger.Warn("obs.Init()", slog.Any("err", err))
		otelSetup = &obs.Setup{Stop: func(context.Context) error { return nil }}
	}
	defer func() {
		if err = otelSetup.Stop(context.Background()); err != nil {
			logger.Error("otelSetup.Stop", slog.Any("err", err))
		}
	}()

	// init db
	database, err := initDB(cfg.DB)
	if err != nil {
		logger.Error("initDb: ", slog.Any("err", err))
		return
	}
	defer func() {
		if err = database.Close(); err != nil {
			logger.Warn("failed to close rows: ", slog.Any("err", err))
		}
	}()

	r := initRouter(cfg, logger)

	// API v1 (auth + rate-limit)
	api := r.Group("/api/v1")

	jwtManager := pkg.NewJWTManager(cfg.Auth.JWTSecret, 24*time.Hour)
	api.Use(pkg.AuthJWT(jwtManager, logger))

	if cfg.RateLimit.Enabled {
		rl := pkg.NewLimiter(&cfg.RateLimit)
		defer rl.Stop()
		api.Use(rl.Middleware(logger))
	}

	swagger, err := handlers.GetSwagger()
	if err != nil {
		logger.Error("handlers.GetSwagger: ", slog.Any("err", err))
		return
	}

	swagger.Servers = nil
	r.Use(ginvalidator.OapiRequestValidator(swagger))

	// load catalog of wod.
	c, err := catalog.NewCatalog(catalog.Raw)
	if err != nil {
		logger.Error("catalog.NewCatalog: ", slog.Any("err", err))
		return
	}

	// init repository
	wodRepo := repository.NewWodRepository(database)

	// init core
	wodGenerateCore := core.NewWodGenerator(c, wodRepo)
	wodListCore := core.NewWodList(c, wodRepo)

	server := handlers.NewServer(wodGenerateCore, wodListCore)
	handlers.RegisterHandlersWithOptions(api, handlers.NewStrictHandler(server, nil), handlers.GinServerOptions{
		BaseURL: "",
	})

	// run and shutdown server
	runServer(ctx, cfg, logger, r)
}

func initDB(dbCfg config.DBConfig) (*sql.DB, error) {
	// Construct the data source name (DSN)
	dataSourceName := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		dbCfg.USER, dbCfg.PASSWORD, dbCfg.HOST, dbCfg.PORT, dbCfg.NAME, dbCfg.SSLMODE)

	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("sql.Open(dbName, url): %w", err)
	}
	// test connection
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}

	return db, nil
}

func initRouter(cfg config.Config, logger *slog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	err := r.SetTrustedProxies(nil)
	if err != nil {
		logger.Warn("SetTrustedProxies: ", slog.Any("err", err))
	}

	r.Use(pkg.RecoveryMiddleware(logger))
	r.Use(pkg.RequestID())
	r.Use(pkg.Logging(logger))
	r.Use(obs.MetricsMiddleware())

	isProd := os.Getenv("ENV") == "prod"
	r.Use(pkg.SecurityHeaders(isProd))
	r.Use(pkg.TimeoutMiddleware(10 * time.Second))
	r.Use(pkg.BodyLimit(cfg.HTTP))

	// Health / Ready
	r.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/readyz", func(c *gin.Context) { c.Status(http.StatusOK) })

	return r
}

func runServer(ctx context.Context, cfg config.Config, logger *slog.Logger, handler http.Handler) {
	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.HTTP.Port),
		Handler:           handler,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", slog.String("addr", srv.Addr))
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server error", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}

	shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shCtx); err != nil {
		logger.Error("graceful shutdown failed", slog.String("err", err.Error()))
	} else {
		logger.Info("server stopped")
	}
}
