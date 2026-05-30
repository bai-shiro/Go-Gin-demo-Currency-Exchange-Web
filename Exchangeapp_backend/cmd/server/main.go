package main

import (
	"context"
	"exchangeapp/internal/config"
	"exchangeapp/internal/dbmigrate"
	"exchangeapp/internal/repository"
	"exchangeapp/internal/router"
	"exchangeapp/internal/service"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	appConfig, err := config.InitConfig()
	if err != nil {
		log.Fatalf("init config:%s", err)
	}

	// 根据配置自动是否迁移数据库
	if appConfig.Database.AutoMigrate {
		log.Println("start database migration")
		dbmigrate.RunMigrations(fmt.Sprintf("mysql://%s", appConfig.Database.Dsn))
		log.Println("database migration success")
	}

	// 依赖注入数据库和业务层
	repos := repository.NewRepositories(appConfig.Db)
	services := service.NewServices(repos, appConfig.RedisDB, appConfig.JWT.Secret, appConfig.JWT.TTL)

	// 注册路由
	r := router.SetupRouter(appConfig, services)

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
