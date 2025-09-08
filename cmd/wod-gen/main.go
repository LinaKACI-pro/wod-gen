//nolint:gocrtic // main function
package main

import (
	"context"
	"errors"
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
	"github.com/LinaKACI-pro/wod-gen/pkg"
	"github.com/caarlos0/env/v11"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	ginvalidator "github.com/oapi-codegen/gin-middleware"
)

func main() {
	// ---- Config ----
	var cfg config.Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("env.Parse: %v", err)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	// ---- Logger ----
	logger := pkg.NewLogger()

	logger.Info("boot",
		slog.Int("port", cfg.HTTP.Port),
		slog.Bool("obs", cfg.Obs.Enabled),
		slog.String("rate_strategy", cfg.RateLimit.Strategy),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	err := r.SetTrustedProxies(nil)
	if err != nil {
		log.Printf("r.SetTrustedProxies: %v\n", err)
	}

	r.Use(pkg.RecoveryMiddleware(logger))
	r.Use(pkg.RequestID())
	r.Use(pkg.Logging(logger))

	isProd := os.Getenv("ENV") == "prod"
	r.Use(pkg.SecurityHeaders(isProd))
	r.Use(pkg.TimeoutMiddleware(10 * time.Second))
	r.Use(pkg.BodyLimit(cfg.HTTP))

	// Health / Ready
	r.GET("/healthz", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/readyz", func(c *gin.Context) { c.Status(http.StatusOK) })

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
		logger.Error("handlers.GetSwagger: ", "err", err)
		return
	}

	swagger.Servers = nil
	r.Use(ginvalidator.OapiRequestValidator(swagger))

	c, err := core.NewCatalog(catalog.Raw)
	if err != nil {
		logger.Error("catalog.NewCatalog: ", "err", err)
		return
	}

	server := handlers.NewServer(c)
	handlers.RegisterHandlersWithOptions(api, handlers.NewStrictHandler(server, nil), handlers.GinServerOptions{
		BaseURL: "",
	})

	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.HTTP.Port),
		Handler:           r,
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
