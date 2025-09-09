package config

import (
	"time"
)

type AuthConfig struct {
	JWTSecret string `env:"AUTH_JWT_SECRET" required:"true"`
}

type RateLimiterConfig struct {
	Enabled      bool          `env:"RATE_LIMITER_ENABLED" envDefault:"true"`
	RequestsPerS float64       `env:"RATE_LIMITER_REQUESTS_PER_SECOND" envDefault:"20"`
	Burst        int           `env:"RATE_LIMITER_BURST" envDefault:"10"`
	TTL          time.Duration `env:"RATE_LIMITER_TTL" envDefault:"1m"`
	// Strategy: "ip", "token", "global"
	Strategy string `env:"RATE_LIMITER_STRATEGY" envDefault:"token" validate:"oneof=ip token global"`
}

type HTTPConfig struct {
	Port           int           `env:"PORT" envDefault:"8080"`
	BodyMaxBytes   int64         `env:"BODY_MAX_BYTES" envDefault:"1048576"` // 1 MiB
	ReadTimeout    time.Duration `env:"READ_TIMEOUT" envDefault:"5s"`
	WriteTimeout   time.Duration `env:"WRITE_TIMEOUT" envDefault:"10s"`
	IdleTimeout    time.Duration `env:"IDLE_TIMEOUT" envDefault:"60s"`
	MaxHeaderBytes int           `env:"MAX_HEADER_BYTES" envDefault:"1048576"`
	InFlightLimit  int           `env:"INFLIGHT_LIMIT" envDefault:"50"`
}

type DBConfig struct {
	NAME     string `env:"DATABASE_NAME" envDefault:"wodgen"`
	USER     string `env:"DATABASE_USER" envDefault:"postgres"`
	PASSWORD string `env:"DATABASE_PASSWORD" envDefault:"postgres"`
	HOST     string `env:"DATABASE_HOST" envDefault:"pgsql"`
	PORT     string `env:"DATABASE_PORT" envDefault:"5432"`
	SSLMODE  string `env:"DATABASE_SSLMODE" envDefault:"disable"`
}

type ObsConfig struct {
	Enabled      bool   `env:"OBS_ENABLED" envDefault:"false"`
	OTLPEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	GrafanaToken string `env:"GRAFANA_CLOUD_TOKEN" envDefault:""`
	OTLPProtocol string `env:"OTEL_EXPORTER_OTLP_PROTOCOL" envDefault:"http/protobuf"`
}

type DebugConfig struct {
	PprofEnabled bool `env:"PPROF_ENABLED" envDefault:"false"`
}

type Config struct {
	Auth      AuthConfig
	RateLimit RateLimiterConfig
	HTTP      HTTPConfig
	DB        DBConfig
	Obs       ObsConfig
	Debug     DebugConfig
}
