package dbmigrate

import (
	"errors"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(dbURL string) {
	// 创建migrate实例
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		log.Fatal("create migration instance failed error:", err)
	}

	defer m.Close()

	// 设置锁超时，防止多实例启动时长时间阻塞
	m.LockTimeout = 30 * time.Second

	// 迁移数据库字段
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("database has no changes to apply(latest version)")
			return
		} else {
			log.Fatal("migration failed error:", err)
		}
	}

	log.Println("database migration success")
}

func RollbackMigrations(dbURL string) {
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		log.Fatal("create migration instance failed error:", err)
	}

	defer m.Close()

	if err := m.Steps(-1); err != nil {
		log.Fatal("migration failed error:", err)
	}
}

func Version(dbURL string) (version uint, dirty bool, err error) {
	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		return 0, false, err
	}

	defer m.Close()

	return m.Version()
}
