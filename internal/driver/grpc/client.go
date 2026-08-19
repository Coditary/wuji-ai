package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	wujiv1 "github.com/coditary/wuji/api/proto/v1"
	"github.com/coditary/wuji/internal/capability"
	"github.com/coditary/wuji/internal/driver"
)

// RemoteDriver wraps a driver exposed over gRPC.
type RemoteDriver struct {
	info   driver.Info
	client wujiv1.DriverServiceClient
	conn   *grpc.ClientConn
}

// Connect creates a remote driver client for the given endpoint.
func Connect(ctx context.Context, endpoint string) (*RemoteDriver, error) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect to driver at %s: %w", endpoint, err)
	}

	client := wujiv1.NewDriverServiceClient(conn)

	resp, err := client.GetInfo(ctx, &wujiv1.GetInfoRequest{})
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("get driver info: %w", err)
	}

	meta := resp.GetMetadata()
	caps := make([]capability.Type, 0, len(meta.GetCapabilities()))
	for _, c := range meta.GetCapabilities() {
		caps = append(caps, capability.Type(c))
	}

	return &RemoteDriver{
		info: driver.Info{
			ID:           meta.GetId(),
			Name:         meta.GetName(),
			Version:      meta.GetVersion(),
			Description:  meta.GetDescription(),
			Capabilities: caps,
			Remote:       true,
			Endpoint:     endpoint,
		},
		client: client,
		conn:   conn,
	}, nil
}

func (d *RemoteDriver) Info() driver.Info {
	return d.info
}

func (d *RemoteDriver) Capabilities() []capability.Type {
	return d.info.Capabilities
}

func (d *RemoteDriver) HasCapability(c capability.Type) bool {
	for _, cap := range d.Capabilities() {
		if cap == c {
			return true
		}
	}
	return false
}

func (d *RemoteDriver) GenerateText(ctx context.Context, req driver.TextRequest) (*driver.TextResponse, error) {
	resp, err := d.client.GenerateText(ctx, &wujiv1.GenerateTextRequest{
		Prompt:      req.Prompt,
		MaxTokens:   int32(req.MaxTokens),
		Temperature: req.Temperature,
	})
	if err != nil {
		return nil, fmt.Errorf("remote generate text: %w", err)
	}

	return &driver.TextResponse{
		Text:         resp.GetText(),
		TokensUsed:   int(resp.GetTokensUsed()),
		FinishReason: resp.GetFinishReason(),
	}, nil
}

func (d *RemoteDriver) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}
