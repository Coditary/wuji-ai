package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	wujiv1 "github.com/coditary/wuji/api/proto/v1"
	llamacfg "github.com/coditary/wuji/drivers/llama/internal/config"
	llamadrv "github.com/coditary/wuji/drivers/llama/internal/driver"
	"github.com/coditary/wuji/internal/driver"
)

type grpcServer struct {
	wujiv1.UnimplementedDriverServiceServer
	drv *llamadrv.Driver
}

func (s *grpcServer) GetInfo(_ context.Context, _ *wujiv1.GetInfoRequest) (*wujiv1.GetInfoResponse, error) {
	info := s.drv.Info()
	caps := make([]string, 0, len(info.Capabilities))
	for _, c := range info.Capabilities {
		caps = append(caps, c.String())
	}
	return &wujiv1.GetInfoResponse{
		Metadata: &wujiv1.DriverMetadata{
			Id: info.ID, Name: info.Name, Version: info.Version,
			Description: info.Description, Capabilities: caps,
		},
	}, nil
}

func (s *grpcServer) GenerateText(ctx context.Context, req *wujiv1.GenerateTextRequest) (*wujiv1.GenerateTextResponse, error) {
	resp, err := s.drv.GenerateText(ctx, driver.TextRequest{
		Prompt: req.GetPrompt(), Model: req.GetModel(),
		MaxTokens: int(req.GetMaxTokens()), Temperature: req.GetTemperature(),
	})
	if err != nil {
		return nil, err
	}
	return &wujiv1.GenerateTextResponse{
		Text: resp.Text, TokensUsed: int32(resp.TokensUsed), FinishReason: resp.FinishReason,
	}, nil
}

func main() {
	var coreAddr = flag.String("core", "", "Wuji core address to register with (optional)")
	flag.Parse()

	cfg, err := llamacfg.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	drv, err := llamadrv.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "driver: %v\n", err)
		os.Exit(1)
	}
	defer drv.Close()

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen: %v\n", err)
		os.Exit(1)
	}

	srv := grpc.NewServer()
	wujiv1.RegisterDriverServiceServer(srv, &grpcServer{drv: drv})

	if *coreAddr != "" {
		conn, err := grpc.NewClient(*coreAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "connect to core: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		info := drv.Info()
		caps := make([]string, 0, len(info.Capabilities))
		for _, c := range info.Capabilities {
			caps = append(caps, c.String())
		}

		client := wujiv1.NewDriverRegistryClient(conn)
		resp, err := client.Register(context.Background(), &wujiv1.RegisterRequest{
			Metadata: &wujiv1.DriverMetadata{
				Id: info.ID, Name: info.Name, Version: info.Version,
				Description: info.Description, Capabilities: caps,
			},
			Endpoint: cfg.GRPCAddr,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "register with core: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Registered with core: %s\n", resp.GetMessage())
	}

	fmt.Fprintf(os.Stderr, "Driver %q listening on %s\n", llamadrv.ID, lis.Addr().String())
	fmt.Fprintf(os.Stderr, "Models directory: %s\n", cfg.ModelsDir)

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(lis) }()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	case <-sigCh:
		srv.GracefulStop()
	}
}
