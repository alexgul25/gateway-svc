package httpapp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	placehandler "github.com/alexgul25/gateway-svc/internal/http/handlers/place"
	userhandler "github.com/alexgul25/gateway-svc/internal/http/handlers/user"
	"github.com/alexgul25/gateway-svc/internal/http/middleware"
	"github.com/alexgul25/gateway-svc/internal/http/routing"
)

type ServerApp struct {
	log             *slog.Logger
	httpServer      *http.Server
	gracefulTimeout time.Duration
}

func New(
	log *slog.Logger,
	userClient userhandler.UserClient,
	placeClient placehandler.PlaceClient,
	jwtSecret []byte,
	serverAddr string,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	idleTimeout time.Duration,
	gracefulTimeout time.Duration,
) *ServerApp {
	userHandler := userhandler.New(userClient)
	placeHandler := placehandler.New(placeClient)

	router := chi.NewRouter()
	router.Use(chimw.RequestID)
	router.Use(middleware.NewLoggerMiddleware(log))
	router.Use(chimw.Recoverer)

	router.Route("/api", func(rtr chi.Router) {
		rtr.Post(routing.PathUsers, userHandler.Register)
		rtr.Post(routing.PathAuthLogin, userHandler.Login)

		rtr.Group(func(r chi.Router) {
			r.Use(middleware.NewAuthMiddleware(jwtSecret))

			r.Get(routing.PathUsersMe, userHandler.GetMyProfile)
			r.Post(routing.PathSubscriptions, userHandler.Subscribe)
			r.Delete(routing.PathSubscriptionByID, userHandler.Unsubscribe)
			r.Get(routing.PathUsersMeFollowers, userHandler.GetFollowers)
			r.Get(routing.PathUserFollowers, userHandler.GetFollowers)

			r.Post(routing.PathPlaces, placeHandler.AddPlace)
			r.Get(routing.PathMyPlaces, placeHandler.GetUserPlaces)
			r.Get(routing.PathUserPlaces, placeHandler.GetUserPlaces)
		})
	})

	httpServer := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	return &ServerApp{
		log:             log,
		httpServer:      httpServer,
		gracefulTimeout: gracefulTimeout,
	}
}

func (sa *ServerApp) Run() error {
	const op = "ServerApp.Run"

	sa.log.Info("http server started", slog.String("addr", sa.httpServer.Addr))

	if err := sa.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (sa *ServerApp) MustRun() {
	if err := sa.Run(); err != nil {
		panic(err)
	}
}

func (sa *ServerApp) GracefulStop() {
	const op = "ServerApp.GracefulStop"

	sa.log.With(slog.String("source", op)).Info("http server shutdown", slog.String("addr", sa.httpServer.Addr))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), sa.gracefulTimeout)
	defer cancel()
	if err := sa.httpServer.Shutdown(shutdownCtx); err != nil {
		sa.log.Error("trouble with shutdown", slog.Any("error", err))
		sa.httpServer.Close()
		return
	}

	sa.log.Info("http server gracefully stopped")
}
