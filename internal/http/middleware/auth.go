package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	jwtpt "github.com/alexgul25/gateway-svc/internal/lib/jwt"
)

var (
	ErrMissedBearerToken = errors.New("bearer token not found")
	ErrInvalidToken      = errors.New("invalid token")
)

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	splitToken := strings.Split(authHeader, "Bearer ")
	if len(splitToken) != 2 {
		return ""
	}

	return splitToken[1]
}

func NewAuthMiddleware(jwtSecret []byte) func(next http.Handler) http.Handler {
	const op = "NewAuthMiddleware"

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			log := LoggerFromContext(r.Context())

			tokenStr := extractBearerToken(r)
			if tokenStr == "" {
				log.Warn("missing bearer token", slog.String("source", op), slog.Any("error", ErrMissedBearerToken))
				w.Header().Set("WWW-Authenticate", "Bearer")
				http.Error(w, "Bearer token is required", http.StatusUnauthorized)

				return
			}

			claims, err := jwtpt.ParseToken(tokenStr, jwtSecret)
			if err != nil {
				log.Warn("failed to parse bearer token", slog.String("source", op), slog.Any("error", err))
				w.Header().Set("WWW-Authenticate", "Bearer")
				http.Error(w, "Invalid bearer token", http.StatusUnauthorized)

				return
			}

			log.Info("user authorized", slog.String("user_id", claims.UserID))

			ctx := context.WithValue(r.Context(), userIDKey{}, claims.UserID)

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

type userIDKey struct{}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey{}).(string)
	return id, ok
}
