package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/zainokta/item-sync/config"
	"github.com/zainokta/item-sync/internal/infrastructure/middleware"
	"github.com/zainokta/item-sync/pkg/logger"
)

type Server interface {
	Start() error
	Stop() error
}

type EchoServer struct {
	echo   *echo.Echo
	config *config.Config
	logger logger.Logger
}

func NewEchoServer(cfg *config.Config, appLogger logger.Logger) (*EchoServer, error) {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.TimeoutWithConfig(echoMiddleware.TimeoutConfig{
		Timeout: cfg.Server.ReadTimeout,
	}))
	e.Use(echoMiddleware.CORSWithConfig(cfg.CORS.ToEchoCORSConfig()))
	e.Use(echoMiddleware.RequestID())
	e.Use(middleware.LogMiddleware(appLogger))

	e.Server.ReadTimeout = cfg.Server.ReadTimeout
	e.Server.WriteTimeout = cfg.Server.WriteTimeout
	e.Server.IdleTimeout = cfg.Server.IdleTimeout

	return &EchoServer{
		echo:   e,
		config: cfg,
		logger: appLogger,
	}, nil
}

func (s *EchoServer) Start() error {
	s.logger.Info("Starting server",
		"port", s.config.Server.Port,
		"host", s.config.Server.Host,
		"environment", s.config.Environment,
	)

	if err := s.echo.Start(fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (s *EchoServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.Server.GracefulTimeout)
	defer cancel()

	s.logger.Info("Shutting down server gracefully", "timeout", s.config.Server.GracefulTimeout)

	if err := s.echo.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}

func (s *EchoServer) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
}
