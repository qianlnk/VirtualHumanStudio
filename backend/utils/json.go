package utils

import (
	"encoding/json"
)

// ToJSONString 将任意类型转换为JSON字符串
func ToJSONString(v interface{}) string {
	if v == nil {
		return ""
	}

	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}

	return string(b)
}