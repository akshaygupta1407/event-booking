package router

import (
	"event-booking/internal/handlers"
	"event-booking/internal/middleware"
	"event-booking/internal/services"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	AuthHandler    *handlers.AuthHandler
	EventHandler   *handlers.EventHandler
	BookingHandler *handlers.BookingHandler
	HealthHandler  *handlers.HealthHandler
	AuthService    *services.AuthService
}

func New(deps Dependencies) *gin.Engine {
	engine := gin.Default()

	engine.GET("/health", deps.HealthHandler.Check)
	engine.POST("/auth/register", deps.AuthHandler.Register)
	engine.POST("/auth/login", deps.AuthHandler.Login)

	authenticated := engine.Group("/")
	authenticated.Use(middleware.Auth(deps.AuthService))
	{
		authenticated.GET("/events", deps.EventHandler.ListAll)
		authenticated.GET("/events/:id", deps.EventHandler.GetByID)

		organizer := authenticated.Group("/")
		organizer.Use(middleware.RequireRole("organizer"))
		{
			organizer.POST("/events", deps.EventHandler.Create)
			organizer.PUT("/events/:id", deps.EventHandler.Update)
			organizer.GET("/organizer/events", deps.EventHandler.ListOrganizerEvents)
		}

		customer := authenticated.Group("/")
		customer.Use(middleware.RequireRole("customer"))
		{
			customer.POST("/events/:id/bookings", deps.BookingHandler.Create)
			customer.GET("/bookings/me", deps.BookingHandler.ListMine)
		}
	}

	return engine
}
