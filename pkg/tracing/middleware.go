package tracing

import (
	otelfiber "github.com/gofiber/contrib/otelfiber/v2"
	"github.com/gofiber/fiber/v2"
)

func Middleware() fiber.Handler {
	return otelfiber.Middleware()
}
