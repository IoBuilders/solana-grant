package router

import (
	"dlt-ingress/src/main/core"
	"dlt-ingress/src/main/dltingress/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/port/health"

	"dlt-ingress/src/main/config"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// ConfigRouter initializes the router and dependencies
func ConfigRouter(app *core.Application) (*gin.Engine, error) {
	/*
		Release mode implies following things:
			* No debug mode warning log.
			* No colors in logs.
			* No route registry log.
			* No complete stack trace for panics
			* Engine lighter and more efficient
	*/
	if config.AppConfig.LoggingConfig.Level != "DEBUG" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// Configure CORS
	router.Use(
		cors.New(cors.Config{
			AllowOrigins: config.AppConfig.Server.AllowedOrigins,
			AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
			ExposeHeaders: []string{
				"Content-Length", "Content-Type", "Authorization",
			},
			AllowCredentials: true,
		}),
		otelgin.Middleware(config.AppConfig.Application.Name),
		middleware.LoggerMiddleware(config.AppConfig.Server.Health.Check.Path, config.AppConfig.Server.Health.Status.Path),
		middleware.RecoveryMiddleware(),
		middleware.TimeOutMiddleware(config.AppConfig.Server.DefaultRequestTimeout),
		middleware.TraceHeaderMiddleware,
	)
	router.NoRoute(middleware.NotFound())

	router.GET(config.AppConfig.Server.Health.Check.Path, health.HealthCheck())
	router.HEAD(config.AppConfig.Server.Health.Check.Path, health.HealthCheck())

	router.GET(config.AppConfig.Server.Health.Status.Path, health.HealthStatus(app.Core.HealthRegistry))
	router.HEAD(config.AppConfig.Server.Health.Status.Path, health.HealthStatus(app.Core.HealthRegistry))

	dltingressconfig.RegisterEndpoints(app.DltIngress, router)

	return router, nil
}
