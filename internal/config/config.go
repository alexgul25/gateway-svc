package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Env        string `envconfig:"ENV"`
	HTTPServer HTTPServerConfig
	JWT        JWTConfig
	GRPCClient GRPCClientConfig
}

type HTTPServerConfig struct {
	Addr            string        `envconfig:"SERVER_ADDR"`
	Timeout         time.Duration `envconfig:"SERVER_TIMEOUT" env-default:"4s"`
	IdleTimeout     time.Duration `envconfig:"SERVER_IDLE_TIMEOUT" env-default:"60s"`
	GracefulTimeout time.Duration `envconfig:"GRACEFUL_TIMEOUT" env-default:"10s"`
}

type JWTConfig struct {
	Secret string `envconfig:"JWT_SECRET"`
}

type GRPCClientConfig struct {
	UserServiceAddr string `envconfig:"USER_SERVICE_ADDR"`
	APIKey          string `envconfig:"API_KEY"`
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
		return nil, fmt.Errorf("%s env variable not set: ENV", op)
	}
	if cfg.HTTPServer.Addr == "" {
		return nil, fmt.Errorf("%s env variable not set: SERVER_ADDR", op)
	}
	if cfg.GRPCClient.UserServiceAddr == "" {
		return nil, fmt.Errorf("%s env variable not set: USER_SERVICE_ADDR", op)
	}
	if cfg.GRPCClient.APIKey == "" {
		return nil, fmt.Errorf("%s env variable not set: API_KEY", op)
	}
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("%s env variable not set: JWT_SECRET", op)
	}

	return cfg, nil
}
