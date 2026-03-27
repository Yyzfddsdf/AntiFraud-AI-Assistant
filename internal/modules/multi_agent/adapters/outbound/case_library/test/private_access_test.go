package case_library_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	case_library "antifraud/internal/modules/multi_agent/adapters/outbound/case_library"
	_ "unsafe"
)

//go:linkname generateCaseEmbedding antifraud/internal/modules/multi_agent/adapters/outbound/case_library.generateCaseEmbedding
var generateCaseEmbedding func(context.Context, string) ([]float64, string, error)

//go:linkname searchHistoricalCasesByVector antifraud/internal/modules/multi_agent/adapters/outbound/case_library.searchHistoricalCasesByVector
var searchHistoricalCasesByVector func([]float64, int) ([]case_library.SimilarCaseResult, int, error)

//go:linkname historicalCaseVectorCacheReadyGet antifraud/internal/modules/multi_agent/adapters/outbound/case_library.historicalCaseVectorCacheReadyGet
var historicalCaseVectorCacheReadyGet func(string, interface{}) (bool, error)

//go:linkname historicalCaseVectorCacheReadySet antifraud/internal/modules/multi_agent/adapters/outbound/case_library.historicalCaseVectorCacheReadySet
var historicalCaseVectorCacheReadySet func(string, interface{}, time.Duration) error

//go:linkname historicalCaseVectorCacheHashGet antifraud/internal/modules/multi_agent/adapters/outbound/case_library.historicalCaseVectorCacheHashGet
var historicalCaseVectorCacheHashGet func(string, string, interface{}) (bool, error)

//go:linkname historicalCaseVectorCacheHashSetJSON antifraud/internal/modules/multi_agent/adapters/outbound/case_library.historicalCaseVectorCacheHashSetJSON
var historicalCaseVectorCacheHashSetJSON func(string, string, interface{}) error

//go:linkname historicalCaseVectorCacheHashDelete antifraud/internal/modules/multi_agent/adapters/outbound/case_library.historicalCaseVectorCacheHashDelete
var historicalCaseVectorCacheHashDelete func(string, string) error

//go:linkname historicalCaseVectorCacheHashGetAll antifraud/internal/modules/multi_agent/adapters/outbound/case_library.historicalCaseVectorCacheHashGetAll
var historicalCaseVectorCacheHashGetAll func(string) (map[string]string, error)

//go:linkname historicalCaseVectorCacheDelete antifraud/internal/modules/multi_agent/adapters/outbound/case_library.historicalCaseVectorCacheDelete
var historicalCaseVectorCacheDelete func(string) error

//go:linkname historicalCaseDBOnce antifraud/internal/platform/database.historicalCaseDBOnce
var historicalCaseDBOnce sync.Once

//go:linkname historicalCaseDB antifraud/internal/platform/database.historicalCaseDB
var historicalCaseDB *gorm.DB

//go:linkname historicalCaseDBErr antifraud/internal/platform/database.historicalCaseDBErr
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
