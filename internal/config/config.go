package config

import (
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Env         string `envconfig:"ENV"`
	ServiceName string `envconfig:"SERVICE_NAME" default:"gateway-svc"`
	HTTPServer  HTTPServerConfig
	JWT         JWTConfig
	GRPCClient  GRPCClientConfig
}

type HTTPServerConfig struct {
	Addr            string        `envconfig:"SERVER_ADDR"`
	ReadTimeout     time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"4s"`
	WriteTimeout    time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"8s"`
	HandlerTimeout  time.Duration `envconfig:"SERVER_HANDLER_TIMEOUT" default:"7s"`
	IdleTimeout     time.Duration `envconfig:"SERVER_IDLE_TIMEOUT" default:"60s"`
	GracefulTimeout time.Duration `envconfig:"GRACEFUL_TIMEOUT" default:"10s"`
}

type JWTConfig struct {
	Secret string `envconfig:"JWT_SECRET"`
}

type GRPCClientConfig struct {
	UserServiceAddr         string        `envconfig:"USER_SERVICE_ADDR"`
	UserServiceTimeout      time.Duration `envconfig:"USER_SERVICE_TIMEOUT" default:"5s"`
	UserServiceRetriesCount int           `envconfig:"USER_SERVICE_RETRY_COUNT" default:"3"`

	PlaceServiceAddr         string        `envconfig:"PLACE_SERVICE_ADDR"`
	PlaceServiceTimeout      time.Duration `envconfig:"PLACE_SERVICE_TIMEOUT" default:"5s"`
	PlaceServiceRetriesCount int           `envconfig:"PLACE_SERVICE_RETRY_COUNT" default:"3"`
}

func load() (*Config, error) {
	const op = "load"

	err := godotenv.Load()
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var cfg Config
	err = envconfig.Process("", &cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &cfg, nil
}

func LoadGatewayService() (*Config, error) {
	const op = "LoadGatewayService"

	cfg, err := load()
	if err != nil {
		return nil, err
	}

	if cfg.Env == "" {
		return nil, fmt.Errorf("%s: env variable ENV not set", op)
	}
	if cfg.HTTPServer.Addr == "" {
		return nil, fmt.Errorf("%s: env variable SERVER_ADDR not set", op)
	}
	if cfg.GRPCClient.UserServiceAddr == "" {
		return nil, fmt.Errorf("%s: env variable USER_SERVICE_ADDR not set", op)
	}
	if cfg.GRPCClient.PlaceServiceAddr == "" {
		return nil, fmt.Errorf("%s: env variable PLACE_SERVICE_ADDR not set", op)
	}
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("%s: env variable JWT_SECRET not set", op)
	}
	if cfg.HTTPServer.HandlerTimeout <= 0 || cfg.HTTPServer.HandlerTimeout >= cfg.HTTPServer.WriteTimeout {
		return nil, fmt.Errorf("%s: SERVER_HANDLER_TIMEOUT (%s) must be positive and less than SERVER_WRITE_TIMEOUT (%s)",
			op, cfg.HTTPServer.HandlerTimeout, cfg.HTTPServer.WriteTimeout)
	}

	return cfg, nil
}
