package server

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/coditary/wuji/drivers/llama/internal/client"
	"github.com/coditary/wuji/drivers/llama/internal/config"
)

type Server struct {
	cfg       *config.Config
	mu        sync.Mutex
	cmd       *exec.Cmd
	client    *client.Client
	modelPath string
}

func New(cfg *config.Config) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) BaseURL() string {
	return fmt.Sprintf("http://%s:%d", s.cfg.InferenceHost, s.cfg.InferencePort)
}

func (s *Server) EnsureRunning(ctx context.Context, modelPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd != nil && s.modelPath == modelPath {
		if s.client != nil && s.client.Healthy(ctx) {
			return nil
		}
		s.stopLocked()
	}

	if err := s.startLocked(ctx, modelPath); err != nil {
		return err
	}

	s.client = client.New(s.BaseURL())
	return s.client.WaitForHealthy(ctx, 2*time.Minute)
}

func (s *Server) Client() *client.Client {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.client
}

func (s *Server) startLocked(ctx context.Context, modelPath string) error {
	if !s.cfg.ServerAvailable() {
		return fmt.Errorf("inference server binary not found at %s (run: make -C drivers/llama setup)", s.cfg.ServerBin)
	}

	binDir := filepath.Dir(s.cfg.ServerBin)
	s.cmd = exec.CommandContext(ctx,
		s.cfg.ServerBin,
		"-m", modelPath,
		"--host", s.cfg.InferenceHost,
		"--port", fmt.Sprintf("%d", s.cfg.InferencePort),
	)
	s.cmd.Dir = binDir
	s.cmd.Env = append(os.Environ(), "LD_LIBRARY_PATH="+binDir)

	stderr, err := s.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}
	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("start inference server: %w", err)
	}

	s.modelPath = modelPath

	go func() {
		buf := make([]byte, 4096)
		for {
			n, readErr := stderr.Read(buf)
			if n > 0 {
				_, _ = os.Stderr.Write(buf[:n])
			}
			if readErr != nil {
				return
			}
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stopLocked()
}

func (s *Server) stopLocked() error {
	if s.cmd == nil || s.cmd.Process == nil {
		s.cmd = nil
		s.client = nil
		s.modelPath = ""
		return nil
	}

	_ = s.cmd.Process.Signal(syscall.SIGTERM)
	done := make(chan error, 1)
	go func() { done <- s.cmd.Wait() }()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		_ = s.cmd.Process.Kill()
		<-done
	}

	s.cmd = nil
	s.client = nil
	s.modelPath = ""
	return nil
}
