package main

import (
	"context"
	config "exchangeapp/internal/config"
	"exchangeapp/internal/models"
	"exchangeapp/internal/repository"
	"exchangeapp/internal/router"
	"exchangeapp/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	appConfig, err := config.InitConfig()
	if err != nil {
		log.Fatalf("init config:%s",err)
	}

	// 迁移数据库字段
	if err := appConfig.Db.AutoMigrate(
		&models.User{},
		&models.Article{},
		&models.ExchangeRate{},
	); err != nil {
		log.Fatal("database migration failed error:", err)
	}

	// 依赖注入数据库和业务层
	repos := repository.NewRepositories(appConfig.Db)
	services := service.NewServices(repos, appConfig.RedisDB)

	// 注册路由
	r := router.SetupRouter(services)

	// port := config.AppConfig.App.Port

	srv := &http.Server{
		Addr:    appConfig.App.Port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

// 	r.Run(port) // listen and serve on 0.0.0.0:8080
// }