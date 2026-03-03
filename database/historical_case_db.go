package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	casemodel "antifraud/multi_agent/case_library/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const historicalCaseDBPathEnv = "HISTORICAL_CASE_DB_PATH"

var (
	historicalCaseDBOnce sync.Once
	historicalCaseDB     *gorm.DB
	historicalCaseDBErr  error
)

// InitHistoricalCaseDB 在服务启动阶段主动初始化案件库数据库连接与表结构。
func InitHistoricalCaseDB() error {
	_, err := GetHistoricalCaseDB()
	return err
}

// GetHistoricalCaseDB 返回案件库数据库连接（单例）。
func GetHistoricalCaseDB() (*gorm.DB, error) {
	historicalCaseDBOnce.Do(func() {
		dbPath := resolveHistoricalCaseDBPath()
		dbDir := filepath.Dir(dbPath)
		if dbDir != "." && dbDir != "" {
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				historicalCaseDBErr = fmt.Errorf("create historical case db directory failed: %w", err)
				return
			}
		}

		db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Error),
		})
		if err != nil {
			historicalCaseDBErr = fmt.Errorf("open historical case db failed: %w", err)
			return
		}
		if err := db.AutoMigrate(&casemodel.HistoricalCaseEntity{}); err != nil {
			historicalCaseDBErr = fmt.Errorf("auto migrate historical case db failed: %w", err)
			return
		}

		log.Printf("[case_library] historical case db path: %s", dbPath)
		historicalCaseDB = db
	})
	if historicalCaseDBErr != nil {
		return nil, historicalCaseDBErr
	}
	return historicalCaseDB, nil
}

func resolveHistoricalCaseDBPath() string {
	if configuredPath := strings.TrimSpace(os.Getenv(historicalCaseDBPathEnv)); configuredPath != "" {
		return configuredPath
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if ok {
		projectRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), ".."))
		return filepath.Join(projectRoot, "DB", "historical_case_library.db")
	}

	workingDir, err := os.Getwd()
	if err == nil {
		return filepath.Join(workingDir, "DB", "historical_case_library.db")
	}
	return filepath.Join("DB", "historical_case_library.db")
}
