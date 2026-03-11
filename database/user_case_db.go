package database

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"antifraud/login_system/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectDB 初始化主业务库连接、连接池参数和 users 表迁移。
func ConnectDB() error {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath()
	}

	dbDir := filepath.Dir(dbPath)
	if dbDir != "." && dbDir != "" {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return err
		}
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err = DB.AutoMigrate(&models.User{}); err != nil {
		return err
	}
	log.Printf("[database] main db initialized: %s", dbPath)
	return nil
}

// defaultDBPath 在未配置 DB_PATH 时给出默认数据库路径。
func defaultDBPath() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if ok {
		projectRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), ".."))
		return filepath.Join(projectRoot, "DB", "auth_system.db")
	}

	workingDir, err := os.Getwd()
	if err == nil {
		return filepath.Join(workingDir, "DB", "auth_system.db")
	}

	return filepath.Join("DB", "auth_system.db")
}
