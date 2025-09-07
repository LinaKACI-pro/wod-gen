package config

import (
	"time"
)

type AuthConfig struct {
	APIKeys []string `env:"API_KEYS" required:"true"`
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
	URL string `env:"DATABASE_URL"`
}

type ObsConfig struct {
	Enabled      bool   `env:"OBS_ENABLED" envDefault:"false"`
	OTLPEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTLPProtocol string `env:"OTEL_EXPORTER_OTLP_PROTOCOL" envDefault:"http/protobuf"`
	OTLPHeaders  string `env:"OTEL_EXPORTER_OTLP_HEADERS"`
	OTELResource string `env:"OTEL_RESOURCE_ATTRIBUTES" envDefault:"service.name=hyrox-api,service.version=0.1.0,env=dev"`
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
