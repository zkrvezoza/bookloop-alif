package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bookloop-alif/internal/delivery/http/handler"
	"github.com/bookloop-alif/internal/delivery/http/middleware"
	"github.com/bookloop-alif/internal/repository/postgres"
	"github.com/bookloop-alif/internal/service"
)

type Deps struct {
	Pool      *pgxpool.Pool
	JWTSecret string
}

func New(d Deps) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery(), middleware.RequestID(), middleware.AccessLog())
	engine.GET("/health", healthHandler(d.Pool))

	userRepo := postgres.NewUserRepo(d.Pool)
	tokenRepo := postgres.NewRefreshTokenRepo(d.Pool)
	authSvc := service.NewAuthService(userRepo, tokenRepo, d.JWTSecret)
	authHandler := handler.NewAuthHandler(authSvc)

	api := engine.Group("/api/v1")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/refresh", authHandler.Refresh)
		api.POST("/auth/logout", authHandler.Logout)

		authed := api.Group("", middleware.Auth([]byte(d.JWTSecret)))
		_ = authed

		manage := authed.Group("/manage", middleware.RequireRole("librarian"))
		_ = manage
	}

	return engine
}

func healthHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := pool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "db unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
