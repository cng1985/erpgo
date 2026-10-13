package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/cng1985/erpgo/backend/internal/api"
	"github.com/cng1985/erpgo/backend/internal/config"
	"github.com/cng1985/erpgo/backend/internal/model"
	"github.com/cng1985/erpgo/backend/internal/store"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func main() {
	fx.New(
		fx.Provide(config.Load, store.Open, api.New),
		fx.Invoke(startServer),
	).Run()
}

func startServer(lifecycle fx.Lifecycle, cfg config.Config, db *gorm.DB, handler *gin.Engine) {
	server := &http.Server{
		Addr:              cfg.Address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			if err := model.Migrate(db); err != nil {
				return err
			}
			if err := store.Seed(db); err != nil {
				return err
			}
			listener, err := net.Listen("tcp", cfg.Address)
			if err != nil {
				return err
			}
			go func() {
				if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
					slog.Error("HTTP server stopped", "error", err)
					os.Exit(1)
				}
			}()
			slog.Info("ERP API listening", "address", cfg.Address, "database", cfg.DatabasePath)
			return nil
		},
		OnStop: func(ctx context.Context) error { return server.Shutdown(ctx) },
	})
}
