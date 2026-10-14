package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/cng1985/erpgo/backend/internal/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"http-server",
	fx.Provide(New),
	fx.Invoke(registerLifecycle),
)

func New(cfg config.Config, handler *gin.Engine) *http.Server {
	return &http.Server{
		Addr:              cfg.Address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func registerLifecycle(lifecycle fx.Lifecycle, server *http.Server, cfg config.Config, shutdowner fx.Shutdowner) {
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			listener, err := net.Listen("tcp", cfg.Address)
			if err != nil {
				return err
			}
			go func() {
				if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
					slog.Error("HTTP server stopped unexpectedly", "error", err)
					if shutdownErr := shutdowner.Shutdown(fx.ExitCode(1)); shutdownErr != nil {
						slog.Error("request application shutdown", "error", shutdownErr)
					}
				}
			}()
			slog.Info("ERP server listening", "address", cfg.Address, "database", cfg.DatabasePath)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}
