package routes

import (
	"github.com/shaik-aaron/fetch-spotify-backend/internal/app"
	"github.com/shaik-aaron/fetch-spotify-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func SetupRoutes(appInstance *app.App) *gin.Engine {
	router := gin.Default()
	router.Use(CORSMiddleware())
	h := &handlers.App{DB: appInstance.DB}

	router.GET("/", h.GetHello)

	v1 := router.Group("/api/v1")
	{
		v1.GET("/refresh-token", h.RefreshToken)
		v1.GET("/get-token", h.GetToken)
	}

	return router
}
