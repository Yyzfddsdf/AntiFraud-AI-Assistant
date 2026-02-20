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
)

var DB *gorm.DB

// ConnectDB 初始化 SQLite 连接、连接池和模型迁移。
func ConnectDB() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath()
	}

	dbDir := filepath.Dir(dbPath)
	if dbDir != "." && dbDir != "" {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			log.Fatal("创建数据库目录失败: ", err)
		}
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败: ", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("获取数据库连接池失败: ", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err = DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatal("自动迁移模型失败: ", err)
	}
}

// defaultDBPath 解析默认数据库路径（优先项目根目录下 DB/auth_system.db）。
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
