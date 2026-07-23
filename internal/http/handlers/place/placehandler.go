package placehandler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/alexgul25/gateway-svc/internal/dto"
	"github.com/alexgul25/gateway-svc/internal/http/handlerutil"
	"github.com/alexgul25/gateway-svc/internal/http/middleware"
	"github.com/alexgul25/gateway-svc/internal/http/routing"
	"github.com/alexgul25/gateway-svc/internal/models/place"
)

type PlaceClient interface {
	AddPlace(ctx context.Context, name string, info string) (addPlaceInfo *place.AddPlaceInfo, err error)
	GetUserPlaces(ctx context.Context, userID string) ([]place.PlaceInfo, error)
}

type Handler struct {
	client PlaceClient
}

func New(client PlaceClient) *Handler {
	return &Handler{client: client}
}

func (h *Handler) AddPlace(w http.ResponseWriter, r *http.Request) {
	const op = "placehandler.AddPlace"

	ctx := r.Context()
	log := middleware.LoggerFromContext(ctx)

	userID, ok := handlerutil.GetUserIDFromContext(w, ctx, log, op)
	if !ok {
		return
	}
	grpcCtx := handlerutil.EnrichGRPCContextWithUserID(ctx, userID)

	var addPlaceReq dto.AddPlaceRequest
	if ok := handlerutil.DecodeJSON(w, r, log, op, &addPlaceReq); !ok {
		return
	}

	addPlaceInfo, err := h.client.AddPlace(grpcCtx, addPlaceReq.Name, addPlaceReq.Info)
	if err != nil {
		handlerutil.WriteGRPCError(w, ctx, log, op, err)
		return
	}

	addPlaceResp := dto.AddPlaceResponse{
		PlaceID:        addPlaceInfo.PlaceID,
		UserID:         addPlaceInfo.UserID,
		PlaceName:      addPlaceInfo.PlaceName,
		PlaceInfo:      addPlaceInfo.PlaceInfo,
		PlaceCreatedAt: addPlaceInfo.PlaceCreatedAt,
	}

	handlerutil.WriteJSON(w, ctx, log, op, http.StatusCreated, addPlaceResp)
}

func (h *Handler) GetUserPlaces(w http.ResponseWriter, r *http.Request) {
	const op = "placehandler.GetUserPlaces"

	ctx := r.Context()
	log := middleware.LoggerFromContext(ctx)

	authUserID, ok := handlerutil.GetUserIDFromContext(w, ctx, log, op)
	if !ok {
		return
	}
	grpcCtx := handlerutil.EnrichGRPCContextWithUserID(ctx, authUserID)

	targetUserID := chi.URLParam(r, routing.ParamUserID)
	if targetUserID == "" {
		targetUserID = authUserID
	}

	places, err := h.client.GetUserPlaces(grpcCtx, targetUserID)
	if err != nil {
		handlerutil.WriteGRPCError(w, ctx, log, op, err)
		return
	}

	dtoPlaces := make([]dto.Place, len(places))
	for i, p := range places {
		dtoPlaces[i] = dto.Place{Name: p.Name, Info: p.Info, CreatedAt: p.CreatedAt}
	}

	getUserPlacesResp := dto.GetUserPlacesResponse{Places: dtoPlaces}

	handlerutil.WriteJSON(w, ctx, log, op, http.StatusOK, getUserPlacesResp)
}
