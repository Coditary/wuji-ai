package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	wujiv1 "github.com/coditary/wuji/api/proto/v1"
	"github.com/coditary/wuji/internal/core"
)

// Server hosts the Wuji gRPC API for remote driver registration.
type Server struct {
	wujiv1.UnimplementedDriverRegistryServer

	core     *core.Core
	grpcSrv  *grpc.Server
	listener net.Listener
}

// NewServer creates a gRPC server bound to the given address.
func NewServer(c *core.Core, addr string) (*Server, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", addr, err)
	}

	s := &Server{
		core:     c,
		grpcSrv:  grpc.NewServer(),
		listener: lis,
	}

	wujiv1.RegisterDriverRegistryServer(s.grpcSrv, s)
	reflection.Register(s.grpcSrv)

	return s, nil
}

// Addr returns the bound network address.
func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

// Start begins serving gRPC requests.
func (s *Server) Start() error {
	return s.grpcSrv.Serve(s.listener)
}

// Stop gracefully shuts down the server.
func (s *Server) Stop() {
	s.grpcSrv.GracefulStop()
}

func (s *Server) Register(ctx context.Context, req *wujiv1.RegisterRequest) (*wujiv1.RegisterResponse, error) {
	if req.Metadata == nil || req.Metadata.Id == "" {
		return &wujiv1.RegisterResponse{Success: false, Message: "driver metadata with ID is required"}, nil
	}
	if req.Endpoint == "" {
		return &wujiv1.RegisterResponse{Success: false, Message: "driver endpoint is required"}, nil
	}

	if err := s.core.ConnectRemote(ctx, req.Endpoint, true); err != nil {
		return &wujiv1.RegisterResponse{Success: false, Message: err.Error()}, nil
	}

	return &wujiv1.RegisterResponse{
		Success: true,
		Message: fmt.Sprintf("driver %q registered from %s", req.Metadata.Id, req.Endpoint),
	}, nil
}

func (s *Server) Heartbeat(_ context.Context, req *wujiv1.HeartbeatRequest) (*wujiv1.HeartbeatResponse, error) {
	if req.DriverId == "" {
		return &wujiv1.HeartbeatResponse{Alive: false}, nil
	}
	_, err := s.core.Registry().Get(req.DriverId)
	return &wujiv1.HeartbeatResponse{Alive: err == nil}, nil
}

func (s *Server) Unregister(_ context.Context, req *wujiv1.UnregisterRequest) (*wujiv1.UnregisterResponse, error) {
	if req.DriverId == "" {
		return &wujiv1.UnregisterResponse{Success: false}, nil
	}
	if err := s.core.Registry().Unregister(req.DriverId); err != nil {
		return &wujiv1.UnregisterResponse{Success: false}, nil
	}
	return &wujiv1.UnregisterResponse{Success: true}, nil
}
