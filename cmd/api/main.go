package main

import (
	"antifraud/internal/bootstrap/server"
	"log"
)

// main 是纯启动壳：组合根与路由装配已迁移到 app/server。
func main() {
	if err := server.Run(); err != nil {
		log.Fatalf("run server failed: %v", err)
	}
}
