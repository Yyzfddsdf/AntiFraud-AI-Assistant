package tool

import (
	"encoding/json"
	"fmt"
)

// ParseArgs 将 JSON 参数字符串解析为指定类型的输入结构体。
func ParseArgs[T any](args string) (T, error) {
	var input T
	if err := json.Unmarshal([]byte(args), &input); err != nil {
		return input, fmt.Errorf("parse arguments failed: %v", err)
	}
	return input, nil
}
