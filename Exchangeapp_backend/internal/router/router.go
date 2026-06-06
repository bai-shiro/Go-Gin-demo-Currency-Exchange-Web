package router

import (
	"exchangeapp/internal/config"
	"exchangeapp/internal/controllers"
	"exchangeapp/internal/middlewares"
	"exchangeapp/internal/service"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(appConfig *config.Config, services *service.Services) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
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
		auth.POST("/refresh", authController.Refresh)
		auth.POST("/logout", authController.Logout)
	}

	rates := r.Group("/api/rates")
	rates.GET("/latest/all", rateController.Latest)
	rates.GET("/latest", rateController.LatestPair)
	rates.GET("/convert", rateController.Convert)
	rates.Use(middlewares.AuthMiddleWare(appConfig.JWT.Secret))
	{
		rates.POST("/create", rateController.Create)
		rates.GET("/history", rateController.History)
	}

	articles := r.Group("/api/articles")
	articles.GET("", articleController.ListPage)
	articles.GET("/:id", articleController.GetByID)
	articles.GET("/:id/likes", articleController.GetLikes)
	articles.Use(middlewares.AuthMiddleWare(appConfig.JWT.Secret))
	{
		articles.POST("", articleController.Create)
		articles.PUT("/:id", articleController.Update)
		articles.DELETE("/:id", articleController.Delete)

		articles.POST("/:id/like", articleController.Like)
	}

	return r
}
