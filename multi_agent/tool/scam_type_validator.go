package tool

import (
	"fmt"
	"strings"

	"antifraud/multi_agent/case_library"
)

// buildScamTypeSchema 构建诈骗类型字段的 JSON Schema。
// 枚举值来自 config/scam_types.json，避免在工具定义里硬编码。
func buildScamTypeSchema(description string) map[string]interface{} {
	allowed := listAllowedScamTypes()
	trimmedDesc := strings.TrimSpace(description)
	if len(allowed) > 0 {
		trimmedDesc = fmt.Sprintf("%s 可选值：%s。", trimmedDesc, strings.Join(allowed, "、"))
	} else {
		trimmedDesc = fmt.Sprintf("%s 可选值读取失败，请检查 config/scam_types.json。", trimmedDesc)
	}

	schema := map[string]interface{}{
		"type":        "string",
		"description": trimmedDesc,
	}

	if len(allowed) > 0 {
		schema["enum"] = allowed
	}
	return schema
}

// normalizeAndValidateScamType 校验并标准化诈骗类型。
func normalizeAndValidateScamType(raw string) (string, error) {
	allowed := listAllowedScamTypes()
	if len(allowed) == 0 {
		return "", fmt.Errorf("scam_type config is empty, please check config/scam_types.json")
	}

	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("scam_type is required, allowed values: %s", strings.Join(allowed, ", "))
	}

	for _, item := range allowed {
		if strings.TrimSpace(item) == trimmed {
			return trimmed, nil
		}
	}
	return "", fmt.Errorf("scam_type is invalid, allowed values: %s", strings.Join(allowed, ", "))
}

func listAllowedScamTypes() []string {
	return append([]string{}, case_library.ListScamTypes()...)
}
