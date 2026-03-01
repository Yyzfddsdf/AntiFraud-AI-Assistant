package database

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"image_recognition/login_system/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectDB 初始化 SQLite 连接、连接池参数和用户表迁移。
func ConnectDB() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath()
	}

	dbDir := filepath.Dir(dbPath)
	if dbDir != "." && dbDir != "" {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			log.Fatal("create database directory failed: ", err)
		}
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		// 仅打印错误日志，避免慢 SQL 和未命中查询刷屏。
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Fatal("connect database failed: ", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("get sql db handle failed: ", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 登录模块当前只依赖用户表，启动时自动迁移。
	if err = DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("auto migrate failed: ", err)
	}
}

// defaultDBPath 在未配置 DB_PATH 时给出默认数据库路径。
func defaultDBPath() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if ok {
		projectRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
		return filepath.Join(projectRoot, "DB", "auth_system.db")
	}

	workingDir, err := os.Getwd()
	if err == nil {
		return filepath.Join(workingDir, "DB", "auth_system.db")
	}

	return filepath.Join("DB", "auth_system.db")
}
