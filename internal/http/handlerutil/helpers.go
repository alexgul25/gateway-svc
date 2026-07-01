package handlerutil

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	grpcclient "github.com/alexgul25/gateway-svc/internal/clients/grpc"
	"github.com/alexgul25/gateway-svc/internal/http/middleware"
)

func DecodeJSON(w http.ResponseWriter, r *http.Request, log *slog.Logger, op string, v any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		log.WarnContext(r.Context(), "request body decode failed", slog.String("source", op), slog.Any("error", err))
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return false
	}

	return true
}

func WriteGRPCError(w http.ResponseWriter, ctx context.Context, log *slog.Logger, op string, err error) {
	st := status.Convert(err)
	switch st.Code() {
	case codes.InvalidArgument:
		log.WarnContext(ctx, "invalid arguments in request", slog.String("source", op), slog.Any("error", err))
		http.Error(w, st.Message(), http.StatusBadRequest)
	case codes.AlreadyExists:
		log.WarnContext(ctx, "already exists", slog.String("source", op), slog.Any("error", err))
		http.Error(w, st.Message(), http.StatusConflict)
	case codes.Unauthenticated:
		log.WarnContext(ctx, "unauthenticated", slog.String("source", op), slog.Any("error", err))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	default:
		log.ErrorContext(ctx, "operation failed", slog.String("source", op), slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func WriteJSON(w http.ResponseWriter, ctx context.Context, log *slog.Logger, op string, status int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		log.ErrorContext(ctx, "response marshal failed", slog.String("source", op), slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		log.ErrorContext(ctx, "response write failed", slog.String("source", op), slog.Any("error", err))
	}
}

func GetUserIDFromContext(w http.ResponseWriter, ctx context.Context, log *slog.Logger, op string) (string, bool) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		log.WarnContext(ctx, "failed to get user id from context", slog.String("source", op))
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}
	return userID, ok
}

func EnrichGRPCContextWithUserID(ctx context.Context, userID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, grpcclient.HeaderUserID, userID)
}
