package state

import (
	"antifraud/internal/platform/database"

	"gorm.io/gorm"
)

// currentStateDB 是 state 包的默认基础设施适配器入口。
func currentStateDB() *gorm.DB {
	return database.DB
}
