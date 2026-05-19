package router

import (
	"exchangeapp/internal/controllers"
	"exchangeapp/internal/middlewares"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		// AllowOriginFunc: func(origin string) bool {
		// return origin == "https://github.com"
		// },
		MaxAge: 12 * time.Hour,
	}))

	auth := r.Group("/api/auth")
	{
		auth.POST("/login", controllers.Login)
		auth.POST("/register", controllers.Register)
	}

	api := r.Group("/api")
	api.GET("/exchangeRates", controllers.GetExchangeRates)
	api.Use(middlewares.AuthMiddleWare())
	{
		api.POST("/exchangeRates", controllers.CreateExchangeRate)
	}

	articles := r.Group("/api/articles")
	articles.GET("", controllers.GetArticles)
	articles.GET("/:id", controllers.GetArticlesByID)
	articles.GET("/:id/like", controllers.GetArticleLikes)
	articles.Use(middlewares.AuthMiddleWare())
	{
		articles.POST("", controllers.CreateArticle)

		articles.POST("/:id/like", controllers.LikeArticle)
	}

	return r
}