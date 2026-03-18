package case_library_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	case_library "antifraud/multi_agent/case_library"
	_ "unsafe"
)

//go:linkname generateCaseEmbedding antifraud/multi_agent/case_library.generateCaseEmbedding
var generateCaseEmbedding func(context.Context, string) ([]float64, string, error)

//go:linkname searchHistoricalCasesByVector antifraud/multi_agent/case_library.searchHistoricalCasesByVector
var searchHistoricalCasesByVector func([]float64, int) ([]case_library.SimilarCaseResult, int, error)

//go:linkname historicalCaseVectorCacheReadyGet antifraud/multi_agent/case_library.historicalCaseVectorCacheReadyGet
var historicalCaseVectorCacheReadyGet func(string, interface{}) (bool, error)

//go:linkname historicalCaseVectorCacheReadySet antifraud/multi_agent/case_library.historicalCaseVectorCacheReadySet
var historicalCaseVectorCacheReadySet func(string, interface{}, time.Duration) error

//go:linkname historicalCaseVectorCacheHashGet antifraud/multi_agent/case_library.historicalCaseVectorCacheHashGet
var historicalCaseVectorCacheHashGet func(string, string, interface{}) (bool, error)

//go:linkname historicalCaseVectorCacheHashSetJSON antifraud/multi_agent/case_library.historicalCaseVectorCacheHashSetJSON
var historicalCaseVectorCacheHashSetJSON func(string, string, interface{}) error

//go:linkname historicalCaseVectorCacheHashDelete antifraud/multi_agent/case_library.historicalCaseVectorCacheHashDelete
var historicalCaseVectorCacheHashDelete func(string, string) error

//go:linkname historicalCaseDBOnce antifraud/database.historicalCaseDBOnce
var historicalCaseDBOnce sync.Once

//go:linkname historicalCaseDB antifraud/database.historicalCaseDB
var historicalCaseDB *gorm.DB

//go:linkname historicalCaseDBErr antifraud/database.historicalCaseDBErr
var historicalCaseDBErr error

func resetHistoricalCaseDB() {
	if historicalCaseDB != nil {
		rawDB, err := historicalCaseDB.DB()
		if err == nil && rawDB != nil {
			_ = rawDB.Close()
		}
	}
	historicalCaseDB = nil
	historicalCaseDBErr = nil
	historicalCaseDBOnce = sync.Once{}
}

func prepareHistoricalCaseDBPath() (string, error) {
	baseDir := filepath.Join("C:\\Users\\user\\.codex\\memories", "case_library_test_db")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", err
	}

	fileName := strings.NewReplacer("\\", "_", "/", "_", ":", "_", " ", "_").Replace(time.Now().Format("20060102150405.000000000")) + ".db"
	return filepath.Join(baseDir, fileName), nil
}
