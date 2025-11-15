package server

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	"txn-processor/config"
	"txn-processor/internal/adapter/inbound/fiber/router"
	"txn-processor/internal/adapter/outbound/gorm/dao"
	"txn-processor/internal/core/service"
	"txn-processor/pkg/tracing"
)

type Server interface {
	Listen(string) error
	Shutdown() error
}

type App struct {
	config *config.App
	server Server
	tracer tracing.Tracer
}

func New() *App {

	var ctx = context.Background()

	var app = new(App)
	app.config = config.New()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.Level(app.config.LogLevel),
	}))
	slog.SetDefault(logger)

	var tracer tracing.Tracer
	var err error
	if app.config.Otel.Tracer.IsEnabled {
		tracer, err = tracing.NewOtelTracer(ctx, app.config.Name, app.config.Otel.Tracer.Endpoint)
		if err != nil {
			slog.Error("Failed to initialize tracer", "error", err)
			os.Exit(1)
		}
		slog.Info("Tracing enabled", "endpoint", app.config.Otel.Tracer.Endpoint)
	} else {
		tracer = tracing.NewBlankTracer()
		slog.Info("Tracing disabled")
	}
	app.tracer = tracer

	services, err := app.build(ctx)
	if err != nil {
		slog.Error("Failed to build application", "error", err)
		os.Exit(1)
	}
	app.server = router.New(services, tracer)

	app.seed(ctx)

	return app
}

func (a *App) Start() {

	slog.Info("Starting service", "name", a.config.Name, "port", a.config.Port, "env", a.config.Env)

	go func() {
		if err := a.server.Listen(":" + a.config.Port); err != nil {
			slog.Error("Failed to start server", "error", err)
			return
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	slog.Warn("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.Shutdown(ctx); err != nil {
		slog.ErrorContext(ctx, "Graceful shutdown failed", "error", err)
	} else {
		slog.InfoContext(ctx, "Graceful shutdown completed successfully")
	}

}

func (a *App) Shutdown(ctx context.Context) error {
	slog.InfoContext(ctx, "Graceful shutdown initiated")

	if err := a.server.Shutdown(); err != nil {
		slog.ErrorContext(ctx, "Error shutting down server", "error", err)
		return err
	}

	if conn, err := dao.GetConnections(); err == nil {
		if err := conn.Close(ctx); err != nil {
			slog.ErrorContext(ctx, "Connection close failed", "error", err)
		}
	} else {
		slog.WarnContext(ctx, "No connection instance found", "error", err)
	}

	if err := a.tracer.Shutdown(ctx); err != nil {
		slog.ErrorContext(ctx, "Failed to shutdown tracer", "error", err)
		return err
	}

	slog.InfoContext(ctx, "Shutdown completed successfully")
	return nil
}

func (a *App) build(ctx context.Context) (*service.Service, error) {

	dao, err := dao.New(ctx, a.config.DB, a.config.Cache, a.tracer)
	if err != nil {
		slog.ErrorContext(ctx, "error while creating dao", "err", err)
		os.Exit(1)
	}

	return service.New(dao, a.tracer), nil
}

func (a *App) seed(ctx context.Context) {

	if a.config.DB.IsAutoMigrate {
		err := dao.AutoMigrate(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "error while running migrator", "err", err)
			os.Exit(1)
		}
	}

}
