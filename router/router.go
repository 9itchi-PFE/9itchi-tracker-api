package router

import (
	"9itchi-tracker-api/handlers"
	"9itchi-tracker-api/services"

	"github.com/gin-gonic/gin"
)

func InitializeRouter(services *services.Services) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	initializeMiddlewares(r)
	initializeRoutes(r, services)

	return r
}

func initializeRoutes(r *gin.Engine, services *services.Services) {
	// Health endpoints
	r.GET("/", handlers.GetHealth)
	r.GET("/health", handlers.GetHealth)

	// public endpoints
	{
		r.GET("/t/:trackingId", handlers.Pixel(services.Tracking))
	}

	// user endpoints
	userGroup := r.Group("/")
	userGroup.Use(AuthMiddleware())
	{
		userGroup.GET("/emails/:id/stats", handlers.EmailStats(services.Tracking))
	}

	// internal endpoints
	internalGroup := r.Group("/")
	internalGroup.Use(InternalAPIKeyMiddleware())
	{
		internalGroup.DELETE("/internal/users/:id", handlers.EraseUser(services.Erase))
		internalGroup.POST("/internal/tracking", handlers.RegisterTracking(services.Tracking))
	}

	// Fallback
	r.NoRoute(handlers.NoRoute)
}
