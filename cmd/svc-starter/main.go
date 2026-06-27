package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexgul25/gateway-svc/internal/config"
	"github.com/alexgul25/gateway-svc/internal/lib/logger"
)

func main() {
	appCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cfg, err := config.LoadGatewayService()
	if err != nil {
		slog.Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	log := logger.New(cfg.Env)

	// 4. Инициализировать хендлеры

	// 5. Организовать chi роутер

	// 6. Инициализировать сервер

	// 7. Запустить сервер

	// 8. Gracefull shotdown
	<-appCtx.Done()

	log.Info("technical work, stay tuned")
}
