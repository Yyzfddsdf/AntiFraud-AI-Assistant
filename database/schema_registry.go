package database

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
)

type MainDBSchemaInitializer struct {
	Name string
	Init func(db *gorm.DB) error
}

var (
	mainDBSchemaInitializersMu sync.RWMutex
	mainDBSchemaInitializers   []MainDBSchemaInitializer
)

// RegisterMainDBSchemaInitializer 注册主业务库 schema 初始化器。
func RegisterMainDBSchemaInitializer(name string, initializer func(db *gorm.DB) error) {
	if initializer == nil {
		return
	}
	mainDBSchemaInitializersMu.Lock()
	defer mainDBSchemaInitializersMu.Unlock()
	mainDBSchemaInitializers = append(mainDBSchemaInitializers, MainDBSchemaInitializer{
		Name: name,
		Init: initializer,
	})
}

// InitMainDBSchemas 顺序执行所有已注册的主业务库 schema 初始化器。
func InitMainDBSchemas() error {
	if DB == nil {
		return fmt.Errorf("main db is not initialized")
	}

	mainDBSchemaInitializersMu.RLock()
	initializers := append([]MainDBSchemaInitializer{}, mainDBSchemaInitializers...)
	mainDBSchemaInitializersMu.RUnlock()

	for _, item := range initializers {
		if item.Init == nil {
			continue
		}
		if err := item.Init(DB); err != nil {
			return fmt.Errorf("init main db schema failed: %s: %w", item.Name, err)
		}
	}
	return nil
}

// InitPersistence 统一初始化数据库连接与 schema。
func InitPersistence() error {
	if err := ConnectDB(); err != nil {
		return fmt.Errorf("init main db failed: %w", err)
	}
	if err := InitMainDBSchemas(); err != nil {
		return err
	}
	if err := InitHistoricalCaseDB(); err != nil {
		return fmt.Errorf("init historical case db failed: %w", err)
	}
	return nil
}
