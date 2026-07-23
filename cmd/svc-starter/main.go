package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	httpapp "github.com/alexgul25/gateway-svc/internal/app/http"
	"github.com/alexgul25/gateway-svc/internal/clients/grpc/placesvc"
	"github.com/alexgul25/gateway-svc/internal/clients/grpc/usersvc"
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

	userClient, err := usersvc.NewClient(
		log,
		cfg.GRPCClient.UserServiceAddr,
		cfg.GRPCClient.UserServiceTimeout,
		cfg.GRPCClient.UserServiceRetriesCount,
		cfg.ServiceName,
	)
	if err != nil {
		log.Error("failed to create usersvc client", slog.Any("error", err))
		os.Exit(1)
	}
	defer userClient.Close()

	placeClient, err := placesvc.NewClient(
		log,
		cfg.GRPCClient.PlaceServiceAddr,
		cfg.GRPCClient.PlaceServiceTimeout,
		cfg.GRPCClient.PlaceServiceRetriesCount,
		cfg.ServiceName,
	)
	if err != nil {
		log.Error("failed to create placesvc client", slog.Any("error", err))
		os.Exit(1)
	}
	defer placeClient.Close()

	jwtSecret := []byte(cfg.JWT.Secret)

	application := httpapp.New(
		log,
		userClient,
		placeClient,
		jwtSecret,
		cfg.HTTPServer.Addr,
		cfg.HTTPServer.ReadTimeout,
		cfg.HTTPServer.WriteTimeout,
		cfg.HTTPServer.IdleTimeout,
		cfg.HTTPServer.GracefulTimeout,
	)

	go func() {
		application.MustRun()
	}()

	<-appCtx.Done()

	application.GracefulStop()
}
