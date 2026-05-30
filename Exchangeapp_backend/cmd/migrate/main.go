package main

import (
	"exchangeapp/internal/config"
	"exchangeapp/internal/dbmigrate"
	"fmt"
	"log"
	"os"
)

func main() {
	appConfig, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("init config:%s", err)
	}

	dbURL := fmt.Sprintf("mysql://%s", config.ApplyDatabaseDSNEnv(appConfig.Database.Dsn))
	if os.Getenv("DB_URL") != "" {
		dbURL = os.Getenv("DB_URL")
	}

	if len(os.Args) < 2 {
		log.Fatal("Please provide a migration command: up, down, or version")
	}

	switch os.Args[1] {
	case "up":
		dbmigrate.RunMigrations(dbURL)
	case "down":
		dbmigrate.RollbackMigrations(dbURL)
	case "version":
		v, dirty, err := dbmigrate.Version(dbURL)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(v, dirty)
	default:
		log.Fatalf("unknown migration command %q, expected up, down, or version", os.Args[1])
	}
}
