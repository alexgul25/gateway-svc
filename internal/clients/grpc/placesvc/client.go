package placesvc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	grpcclient "github.com/alexgul25/gateway-svc/internal/clients/grpc"
	"github.com/alexgul25/gateway-svc/internal/models/place"
	placev1 "github.com/alexgul25/protos/gen/go/place/v1"
)

type Client struct {
	api  placev1.PlaceServiceClient
	conn *grpc.ClientConn
}

func NewClient(log *slog.Logger, addr string, timeout time.Duration, retriesCount int, serviceName string) (*Client, error) {
	const op = "placesvc.NewClient"

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

	grpcClient := placev1.NewPlaceServiceClient(conn)

	return &Client{
		api:  grpcClient,
		conn: conn,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) AddPlace(ctx context.Context, name string, info string) (addPlaceInfo *place.AddPlaceInfo, err error) {
	const op = "placesvc.Client.AddPlace"

	resp, err := c.api.AddPlace(ctx, &placev1.AddPlaceRequest{
		Name: name,
		Info: info,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &place.AddPlaceInfo{
		PlaceID:        resp.Id,
		UserID:         resp.UserId,
		PlaceName:      resp.Name,
		PlaceInfo:      resp.Info,
		PlaceCreatedAt: resp.CreatedAt.AsTime(),
	}, nil
}

func (c *Client) GetUserPlaces(ctx context.Context, userID string) ([]place.PlaceInfo, error) {
	const op = "placesvc.Client.GetUserPlaces"

	resp, err := c.api.GetUserPlaces(ctx, &placev1.GetUserPlacesRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	places := make([]place.PlaceInfo, len(resp.Places))
	for i, p := range resp.Places {
		places[i] = place.PlaceInfo{Name: p.Name, Info: p.Info, CreatedAt: p.CreatedAt.AsTime()}
	}

	return places, nil
}
