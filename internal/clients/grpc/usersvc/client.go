package usersvc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	userv1 "github.com/alexgul25/protos/gen/go/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	grpcclient "github.com/alexgul25/gateway-svc/internal/clients/grpc"
	"github.com/alexgul25/gateway-svc/internal/models/user"
)

type Client struct {
	api  userv1.UserServiceClient
	conn *grpc.ClientConn
}

func New(log *slog.Logger, addr string, timeout time.Duration, retriesCount int, serviceName string) (*Client, error) {
	const op = "grpc.New"

	kvToAdd := []string{grpcclient.HeaderServiceName, serviceName}
	headersToLog := []string{grpcclient.HeaderServiceName, grpcclient.HeaderUserID}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpcclient.NewAddingHeadersInterceptor(kvToAdd),
			grpcclient.NewLoggingInterceptor(log, headersToLog),
			grpcclient.NewRetryInterceptor(retriesCount, timeout),
		),
	}

	conn, err := grpc.NewClient(addr, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	grpcClient := userv1.NewUserServiceClient(conn)

	return &Client{
		api:  grpcClient,
		conn: conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Register(ctx context.Context, email string, password string, displayName string) (*user.RegisterInfo, error) {
	const op = "grpc.Client.Register"

	resp, err := c.api.Register(ctx, &userv1.RegisterRequest{
		Email:       email,
		Password:    password,
		DisplayName: displayName,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user.RegisterInfo{
		UserID:      resp.UserId,
		Email:       resp.Email,
		DisplayName: resp.DisplayName,
		CreatedAt:   resp.CreatedAt.AsTime(),
	}, nil
}

func (c *Client) Login(ctx context.Context, email, password string) (string, error) {
	const op = "grpc.Client.Login"

	resp, err := c.api.Login(ctx, &userv1.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resp.AccessToken, nil
}

func (c *Client) GetMyProfile(ctx context.Context) (*user.GetMyProfileInfo, error) {
	const op = "grpc.Client.GetMyProfile"

	resp, err := c.api.GetMyProfile(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user.GetMyProfileInfo{
		UserID:      resp.UserId,
		Email:       resp.Email,
		DisplayName: resp.DisplayName,
		CreatedAt:   resp.CreatedAt.AsTime(),
	}, nil
}

func (c *Client) Subscribe(ctx context.Context, followeeID string) error {
	const op = "grpc.Client.Subscribe"

	_, err := c.api.Subscribe(ctx, &userv1.SubscribeRequest{
		FolloweeId: followeeID,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) Unsubscribe(ctx context.Context, followeeID string) error {
	const op = "grpc.Client.Unsubscribe"

	_, err := c.api.Unsubscribe(ctx, &userv1.UnsubscribeRequest{
		FolloweeId: followeeID,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) GetFollowers(ctx context.Context, userID string) ([]user.FollowerInfo, error) {
	const op = "grpc.Client.GetFollowers"

	resp, err := c.api.GetFollowers(ctx, &userv1.GetFollowersRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	followers := make([]user.FollowerInfo, len(resp.Followers))
	for i, follower := range resp.Followers {
		followers[i] = user.FollowerInfo{UserID: follower.UserId, Email: follower.Email, DisplayName: follower.DisplayName}
	}

	return followers, nil
}
