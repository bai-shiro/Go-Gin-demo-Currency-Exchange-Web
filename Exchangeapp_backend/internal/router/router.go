package router

import (
	"exchangeapp/internal/controllers"
	"exchangeapp/internal/middlewares"
	"exchangeapp/internal/service"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(services *service.Services) *gin.Engine {
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

	authController := controllers.NewAuthController(services.Auth)
	articleController := controllers.NewArticleController(services.Articles)
	rateController := controllers.NewRateController(services.Rates)

	auth := r.Group("/api/auth")
	{
		auth.POST("/login", authController.Login)
		auth.POST("/register", authController.Register)
	}

	api := r.Group("/api")
	api.GET("/exchangeRates", rateController.Latest)
	api.Use(middlewares.AuthMiddleWare())
	{
		api.POST("/exchangeRates", rateController.Create)
	}

	articles := r.Group("/api/articles")
	articles.GET("", articleController.List)
	articles.GET("/:id", articleController.GetByID)
	articles.GET("/:id/likes", articleController.GetLikes)
	articles.Use(middlewares.AuthMiddleWare())
	{
		articles.POST("", articleController.Create)
		articles.PUT("/:id", articleController.Update)
		articles.DELETE("/:id", articleController.Delete)

		articles.POST("/:id/like", articleController.Like)
	}

	return r
}