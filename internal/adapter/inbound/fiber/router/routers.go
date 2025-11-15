package router

import (
	"txn-processor/internal/adapter/inbound/fiber/handler"
	"txn-processor/internal/adapter/inbound/fiber/middleware"
	"txn-processor/internal/core/service"
	"txn-processor/internal/port"
	"txn-processor/pkg/tracing"

	"github.com/gofiber/fiber/v2"
)

func New(svc *service.Service, tracer tracing.Tracer) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})
	SetupRoutes(app, svc, tracer)
	return app
}

func SetupRoutes(app *fiber.App, inbound port.Inbound, tracer tracing.Tracer) {
	app.Use(middleware.RequestLogger())

	if tracer.IsEnabled() {
		app.Use(tracing.Middleware())
	}

	v1 := app.Group("/v1")
	HealthRoutes(v1, inbound)

	// slog.Debug("Fiber server initialized", "routes", app.GetRoutes())
}

func HealthRoutes(router fiber.Router, svc port.HealthService) {
	h := handler.NewHealthHandler(svc)
	router.Get("/health", h.HealthCheck)
}
