package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v8"
)

type Config struct {
	ServerPort     string `env:"SERVER_PORT"      envDefault:"8080"`
	DefaultTimeout int    `env:"DEFAULT_TIMEOUT"  envDefault:"60"`
	Status         string `env:"STATUS"           envDefault:"UP"`
	APIKey         string `env:"GATEWAY_API_KEY"  envDefault:"gateway-default-api-key-2025"`

	Database DatabaseConfig
	Tracing  TracingConfig
}

type DatabaseConfig struct {
	Name           string `env:"DATABASE"         envDefault:"gateway.db"`
	Path           string `env:"DATABASE_PATH"    envDefault:"db"`
	TimeoutSeconds int    `env:"DATABASE_TIMEOUT" envDefault:"10"`
}

type TracingConfig struct {
	Enabled        bool   `env:"JAEGER_ENABLED"          envDefault:"false"`
	ServiceName    string `env:"JAEGER_SERVICE_NAME"     envDefault:"gateway"`
	ServiceVersion string `env:"JAEGER_SERVICE_VERSION"  envDefault:"1.0.0"`
	Environment    string `env:"JAEGER_ENVIRONMENT"      envDefault:"development"`
	Endpoint       string `env:"JAEGER_ENDPOINT"         envDefault:"http://localhost:4318"`
}

func Load() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		panic(fmt.Sprintf("failed to parse config: %v", err))
	}
	return cfg
}

func (c *Config) Addr() string {
	return ":" + c.ServerPort
}

func (c *Config) Timeout() time.Duration {
	return time.Duration(c.DefaultTimeout) * time.Second
}

func (d *DatabaseConfig) FilePath() string {
	return fmt.Sprintf("%s/%s", d.Path, d.Name)
}

func (d *DatabaseConfig) Timeout() time.Duration {
	return time.Duration(d.TimeoutSeconds) * time.Second
}
