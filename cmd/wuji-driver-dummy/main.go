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
	"github.com/coditary/wuji/internal/driver"
	"github.com/coditary/wuji/internal/driver/dummy"
)

type driverServer struct {
	wujiv1.UnimplementedDriverServiceServer
	dummy *dummy.Driver
}

func (s *driverServer) GetInfo(_ context.Context, _ *wujiv1.GetInfoRequest) (*wujiv1.GetInfoResponse, error) {
	info := s.dummy.Info()
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

func (s *driverServer) GenerateText(ctx context.Context, req *wujiv1.GenerateTextRequest) (*wujiv1.GenerateTextResponse, error) {
	resp, err := s.dummy.GenerateText(ctx, driver.TextRequestFromProto(req))
	if err != nil {
		return nil, err
	}
	return &wujiv1.GenerateTextResponse{
		Text: resp.Text, TokensUsed: int32(resp.TokensUsed), FinishReason: resp.FinishReason,
	}, nil
}

func (s *driverServer) GenerateImage(ctx context.Context, req *wujiv1.GenerateImageRequest) (*wujiv1.GenerateImageResponse, error) {
	resp, err := s.dummy.GenerateImage(ctx, driver.ImageRequestFromProto(req))
	if err != nil {
		return nil, err
	}
	return driver.ImageResponseToProto(resp), nil
}

func (s *driverServer) GenerateVideo(ctx context.Context, req *wujiv1.GenerateVideoRequest) (*wujiv1.GenerateVideoResponse, error) {
	resp, err := s.dummy.GenerateVideo(ctx, driver.VideoRequestFromProto(req))
	if err != nil {
		return nil, err
	}
	return driver.VideoResponseToProto(resp), nil
}

func (s *driverServer) GenerateAudio(ctx context.Context, req *wujiv1.GenerateAudioRequest) (*wujiv1.GenerateAudioResponse, error) {
	resp, err := s.dummy.GenerateAudio(ctx, driver.AudioRequestFromProto(req))
	if err != nil {
		return nil, err
	}
	return driver.AudioResponseToProto(resp), nil
}

func (s *driverServer) Generate3D(ctx context.Context, req *wujiv1.Generate3DRequest) (*wujiv1.Generate3DResponse, error) {
	resp, err := s.dummy.Generate3D(ctx, driver.Asset3DRequest{
		Prompt: req.GetPrompt(), Format: req.GetFormat(),
	})
	if err != nil {
		return nil, err
	}
	return &wujiv1.Generate3DResponse{Path: resp.Path, Format: resp.Format}, nil
}

func (s *driverServer) Synthesize(ctx context.Context, req *wujiv1.SynthesizeRequest) (*wujiv1.SynthesizeResponse, error) {
	resp, err := s.dummy.Synthesize(ctx, driver.TTSRequestFromProto(req))
	if err != nil {
		return nil, err
	}
	return driver.TTSResponseToProto(resp), nil
}

func (s *driverServer) Transcribe(ctx context.Context, req *wujiv1.TranscribeRequest) (*wujiv1.TranscribeResponse, error) {
	resp, err := s.dummy.Transcribe(ctx, driver.STTRequestFromProto(req))
	if err != nil {
		return nil, err
	}
	return driver.STTResponseToProto(resp), nil
}

func (s *driverServer) CloneVoice(ctx context.Context, req *wujiv1.CloneVoiceRequest) (*wujiv1.CloneVoiceResponse, error) {
	resp, err := s.dummy.CloneVoice(ctx, driver.VoiceRequestFromProto(req))
	if err != nil {
		return nil, err
	}
	return driver.VoiceResponseToProto(resp), nil
}

func (s *driverServer) TrainText(ctx context.Context, req *wujiv1.TrainTextRequest) (*wujiv1.TrainResponse, error) {
	resp, err := s.dummy.TrainText(ctx, driver.TextTrainRequestFromProto(req))
	if err != nil {
		return nil, err
	}
	return driver.TrainResponseToProto(resp), nil
}

func (s *driverServer) TrainImage(ctx context.Context, req *wujiv1.TrainImageRequest) (*wujiv1.TrainResponse, error) {
	resp, err := s.dummy.TrainImage(ctx, driver.ImageTrainRequestFromProto(req))
	if err != nil {
		return nil, err
	}
	return driver.TrainResponseToProto(resp), nil
}

func (s *driverServer) TrainVideo(ctx context.Context, req *wujiv1.TrainVideoRequest) (*wujiv1.TrainResponse, error) {
	resp, err := s.dummy.TrainVideo(ctx, driver.VideoTrainRequestFromProto(req))
	if err != nil {
		return nil, err
	}
	return driver.TrainResponseToProto(resp), nil
}

func (s *driverServer) TrainAudio(ctx context.Context, req *wujiv1.TrainAudioRequest) (*wujiv1.TrainResponse, error) {
	resp, err := s.dummy.TrainAudio(ctx, driver.AudioTrainRequestFromProto(req))
	if err != nil {
		return nil, err
	}
	return driver.TrainResponseToProto(resp), nil
}

func (s *driverServer) Train3D(ctx context.Context, req *wujiv1.Train3DRequest) (*wujiv1.TrainResponse, error) {
	resp, err := s.dummy.Train3D(ctx, driver.Asset3DTrainRequestFromProto(req))
	if err != nil {
		return nil, err
	}
	return driver.TrainResponseToProto(resp), nil
}

func (s *driverServer) TrainTTS(ctx context.Context, req *wujiv1.TrainTTSRequest) (*wujiv1.TrainResponse, error) {
	resp, err := s.dummy.TrainTTS(ctx, driver.TTSTrainRequestFromProto(req))
	if err != nil {
		return nil, err
	}
	return driver.TrainResponseToProto(resp), nil
}

func (s *driverServer) TrainSTT(ctx context.Context, req *wujiv1.TrainSTTRequest) (*wujiv1.TrainResponse, error) {
	resp, err := s.dummy.TrainSTT(ctx, driver.STTTrainRequestFromProto(req))
	if err != nil {
		return nil, err
	}
	return driver.TrainResponseToProto(resp), nil
}

func (s *driverServer) TrainVoice(ctx context.Context, req *wujiv1.TrainVoiceRequest) (*wujiv1.TrainResponse, error) {
	resp, err := s.dummy.TrainVoice(ctx, driver.VoiceTrainRequestFromProto(req))
	if err != nil {
		return nil, err
	}
	return driver.TrainResponseToProto(resp), nil
}

func (s *driverServer) ManageDataset(ctx context.Context, req *wujiv1.ManageDatasetRequest) (*wujiv1.ManageDatasetResponse, error) {
	resp, err := s.dummy.ManageDataset(ctx, driver.DatasetRequest{
		Action: driver.DatasetAction(req.GetAction()), Name: req.GetName(), Path: req.GetPath(),
	})
	if err != nil {
		return nil, err
	}

	entries := make([]*wujiv1.DatasetEntry, 0, len(resp.Datasets))
	for _, ds := range resp.Datasets {
		entries = append(entries, &wujiv1.DatasetEntry{
			Id: ds.ID, Name: ds.Name, Path: ds.Path, Size: ds.Size,
		})
	}
	return &wujiv1.ManageDatasetResponse{Datasets: entries, Message: resp.Message}, nil
}

func main() {
	var (
		addr     = flag.String("addr", "127.0.0.1:50052", "gRPC listen address for this driver")
		coreAddr = flag.String("core", "", "Wuji core address to register with (optional)")
	)
	flag.Parse()

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen: %v\n", err)
		os.Exit(1)
	}

	srv := grpc.NewServer()
	wujiv1.RegisterDriverServiceServer(srv, &driverServer{dummy: dummy.New()})

	if *coreAddr != "" {
		conn, err := grpc.NewClient(*coreAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "connect to core: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		client := wujiv1.NewDriverRegistryClient(conn)
		info := dummy.New().Info()
		caps := make([]string, 0, len(info.Capabilities))
		for _, c := range info.Capabilities {
			caps = append(caps, c.String())
		}

		resp, err := client.Register(context.Background(), &wujiv1.RegisterRequest{
			Metadata: &wujiv1.DriverMetadata{
				Id: info.ID, Name: info.Name, Version: info.Version,
				Description: info.Description, Capabilities: caps,
			},
			Endpoint: *addr,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "register with core: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Registered with core: %s\n", resp.GetMessage())
	}

	fmt.Fprintf(os.Stderr, "Dummy driver listening on %s\n", lis.Addr().String())

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
