package http

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(handler *Handler, sessionStore sessions.Store) *gin.Engine {
	router := gin.Default()

	// Prometheus middleware
	router.Use(PrometheusMiddleware())

	// Session middleware
	router.Use(sessions.Sessions("auth_session", sessionStore))

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Routes
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/health", handler.HealthCheck)
	router.GET("/login", handler.LoginGet)
	router.POST("/login", handler.LoginPost)
	router.GET("/consent", handler.ConsentGet)
	router.POST("/consent", handler.ConsentPost)
	router.GET("/logout", handler.LogoutGet)
	router.POST("/logout", handler.LogoutPost)
	router.GET("/error", handler.ErrorGet)

	// Static files for UI
	// Note: These paths are relative to where the binary is run
	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "./static")

	return router
}
