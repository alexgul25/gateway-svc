package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Env         string `envconfig:"ENV"`
	ServiceName string `envconfig:"SERVICE_NAME" env-default:"gateway-svc"`
	HTTPServer  HTTPServerConfig
	JWT         JWTConfig
	GRPCClient  GRPCClientConfig
}

type HTTPServerConfig struct {
	Addr            string        `envconfig:"SERVER_ADDR"`
	ReadTimeout     time.Duration `envconfig:"SERVER_READ_TIMEOUT" env-default:"4s"`
	WriteTimeout    time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" env-default:"8s"`
	IdleTimeout     time.Duration `envconfig:"SERVER_IDLE_TIMEOUT" env-default:"60s"`
	GracefulTimeout time.Duration `envconfig:"GRACEFUL_TIMEOUT" env-default:"10s"`
}

type JWTConfig struct {
	Secret string `envconfig:"JWT_SECRET"`
}

type GRPCClientConfig struct {
	UserServiceAddr         string        `envconfig:"USER_SERVICE_ADDR"`
	UserServiceTimeout      time.Duration `envconfig:"USER_SERVICE_TIMEOUT" env-default:"5s"`
	UserServiceRetriesCount int           `envconfig:"USER_SERVICE_RETRY_COUNT" env-default:"3"`
}

func load() (*Config, error) {
	const op = "load"

	err := godotenv.Load()
	if err != nil {
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
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("%s: env variable JWT_SECRET not set", op)
	}

	return cfg, nil
}
