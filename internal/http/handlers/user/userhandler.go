package userhandler

import (
	"context"
	"net/http"

	"github.com/alexgul25/gateway-svc/internal/dto"
	"github.com/alexgul25/gateway-svc/internal/http/middleware"
	"github.com/alexgul25/gateway-svc/internal/models/user"
)

type UserClient interface {
	Register(ctx context.Context, email string, password string, displayName string) (registerInfo *user.RegisterInfo, err error)
	Login(ctx context.Context, email string, password string) (accessToken string, err error)
	GetMyProfile(ctx context.Context) (getMyProfileInfo *user.GetMyProfileInfo, err error)
	Subscribe(ctx context.Context, followeeID string) error
	Unsubscribe(ctx context.Context, followeeID string) error
	GetFollowers(ctx context.Context, userID string) ([]user.FollowerInfo, error)
}

type Handler struct {
	client UserClient
}

func New(client UserClient) *Handler {
	return &Handler{client: client}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	const op = "userhandler.Register"

	ctx := r.Context()
	log := middleware.LoggerFromContext(ctx)

	var registerReq dto.RegisterRequest
	if ok := decodeJSON(w, r, log, op, &registerReq); !ok {
		return
	}

	registerInfo, err := h.client.Register(ctx, registerReq.Email, registerReq.Password, registerReq.DisplayName)
	if err != nil {
		writeGRPCError(w, ctx, log, op, err)
		return
	}

	registerResp := dto.RegisterResponse{
		UserID:      registerInfo.UserID,
		Email:       registerInfo.Email,
		DisplayName: registerInfo.DisplayName,
		CreatedAt:   registerInfo.CreatedAt,
	}

	writeJSON(w, ctx, log, op, http.StatusCreated, registerResp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	const op = "userhandler.Login"

	ctx := r.Context()
	log := middleware.LoggerFromContext(ctx)

	var loginReq dto.LoginRequest
	if ok := decodeJSON(w, r, log, op, &loginReq); !ok {
		return
	}

	accessToken, err := h.client.Login(ctx, loginReq.Email, loginReq.Password)
	if err != nil {
		writeGRPCError(w, ctx, log, op, err)
		return
	}

	loginResp := dto.LoginResponse{AccessToken: accessToken}

	writeJSON(w, ctx, log, op, http.StatusOK, loginResp)
}

func (h *Handler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	const op = "userhandler.GetMyProfile"

	ctx := r.Context()
	log := middleware.LoggerFromContext(ctx)

	userID, ok := getUserIDFromContext(w, ctx, log, op)
	if !ok {
		return
	}
	grpcCtx := enrichContextWithUserID(ctx, userID)

	getMyProfileInfo, err := h.client.GetMyProfile(grpcCtx)
	if err != nil {
		writeGRPCError(w, ctx, log, op, err)
		return
	}

	getMyProfileResp := dto.GetMyProfileResponse{
		UserID:      getMyProfileInfo.UserID,
		Email:       getMyProfileInfo.Email,
		DisplayName: getMyProfileInfo.DisplayName,
		CreatedAt:   getMyProfileInfo.CreatedAt,
	}

	writeJSON(w, ctx, log, op, http.StatusOK, getMyProfileResp)
}

func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	const op = "userhandler.Subscribe"

	ctx := r.Context()
	log := middleware.LoggerFromContext(ctx)

	userID, ok := getUserIDFromContext(w, ctx, log, op)
	if !ok {
		return
	}
	grpcCtx := enrichContextWithUserID(ctx, userID)

	var subscribeReq dto.SubscribeRequest
	if ok := decodeJSON(w, r, log, op, &subscribeReq); !ok {
		return
	}

	err := h.client.Subscribe(grpcCtx, subscribeReq.FolloweeID)
	if err != nil {
		writeGRPCError(w, ctx, log, op, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	const op = "userhandler.Unsubscribe"

	ctx := r.Context()
	log := middleware.LoggerFromContext(ctx)

	userID, ok := getUserIDFromContext(w, ctx, log, op)
	if !ok {
		return
	}
	grpcCtx := enrichContextWithUserID(ctx, userID)

	var unsubscribeReq dto.UnsubscribeRequest
	if ok := decodeJSON(w, r, log, op, &unsubscribeReq); !ok {
		return
	}

	err := h.client.Unsubscribe(grpcCtx, unsubscribeReq.FolloweeID)
	if err != nil {
		writeGRPCError(w, ctx, log, op, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetFollowers(w http.ResponseWriter, r *http.Request) {
	const op = "userhandler.GetFollowers"

	ctx := r.Context()
	log := middleware.LoggerFromContext(ctx)

	userID, ok := getUserIDFromContext(w, ctx, log, op)
	if !ok {
		return
	}
	grpcCtx := enrichContextWithUserID(ctx, userID)

	var getFollowersReq dto.GetFollowersRequest
	if ok := decodeJSON(w, r, log, op, &getFollowersReq); !ok {
		return
	}

	followers, err := h.client.GetFollowers(grpcCtx, getFollowersReq.UserID)
	if err != nil {
		writeGRPCError(w, ctx, log, op, err)
		return
	}

	dtoFollowers := make([]dto.Follower, len(followers))
	for i, f := range followers {
		dtoFollowers[i] = dto.Follower{UserID: f.UserID, DisplayName: f.DisplayName, Email: f.Email}
	}

	getFollowersResp := dto.GetFollowersResponse{Followers: dtoFollowers}

	writeJSON(w, ctx, log, op, http.StatusOK, getFollowersResp)
}
