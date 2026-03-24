package database

import (
	"os"
	"path/filepath"
	"runtime"
)

const projectRootDepthFromDatabasePackage = 3

// defaultProjectDBPath 返回仓库根目录下 DB 目录中的默认数据库路径。
func defaultProjectDBPath(fileName string) string {
	projectRoot := resolveProjectRootFromDatabasePackage()
	if projectRoot != "" {
		return filepath.Join(projectRoot, "DB", fileName)
	}

	workingDir, err := os.Getwd()
	if err == nil {
		return filepath.Join(workingDir, "DB", fileName)
	}

	return filepath.Join("DB", fileName)
}

func resolveProjectRootFromDatabasePackage() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}

	projectRoot := filepath.Dir(currentFile)
	for i := 0; i < projectRootDepthFromDatabasePackage; i++ {
		projectRoot = filepath.Dir(projectRoot)
	}
	return filepath.Clean(projectRoot)
}
