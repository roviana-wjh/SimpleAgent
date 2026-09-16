package tools

import (
	"fmt"
	"reflect"
)

// Validator 参数验证器
type Validator struct{}

// NewValidator 创建新的验证器
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateRequired 验证必需参数
func (v *Validator) ValidateRequired(args map[string]interface{}, required []string) error {
	for _, key := range required {
		if _, exists := args[key]; !exists {
			return fmt.Errorf("missing required parameter: %s", key)
		}
	}
	return nil
}

// ValidateType 验证参数类型
func (v *Validator) ValidateType(args map[string]interface{}, key string, expectedType string) error {
	value, exists := args[key]
	if !exists {
		return nil // 如果参数不存在，由 ValidateRequired 处理
	}

	var valid bool
	switch expectedType {
	case "string":
		_, valid = value.(string)
	case "number":
		switch value.(type) {
		case float64, float32, int, int32, int64:
			valid = true
		}
	case "integer":
		switch value.(type) {
		case int, int32, int64, float64:
			// float64 需要检查是否为整数
			if f, ok := value.(float64); ok {
				valid = f == float64(int64(f))
			} else {
				valid = true
			}
		}
	case "boolean":
		_, valid = value.(bool)
	case "array":
		valid = reflect.TypeOf(value).Kind() == reflect.Slice
	case "object":
		_, valid = value.(map[string]interface{})
	default:
		return fmt.Errorf("unknown type: %s", expectedType)
	}

	if !valid {
		return fmt.Errorf("parameter %s must be of type %s, got %T", key, expectedType, value)
	}

	return nil
}

// GetString 获取字符串参数
func GetString(args map[string]interface{}, key string) (string, error) {
	value, exists := args[key]
	if !exists {
		return "", fmt.Errorf("parameter %s not found", key)
	}

	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("parameter %s must be string, got %T", key, value)
	}

	return str, nil
}

// GetNumber 获取数字参数
func GetNumber(args map[string]interface{}, key string) (float64, error) {
	value, exists := args[key]
	if !exists {
		return 0, fmt.Errorf("parameter %s not found", key)
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("parameter %s must be number, got %T", key, value)
	}
}

// GetInt 获取整数参数
func GetInt(args map[string]interface{}, key string) (int, error) {
	value, exists := args[key]
	if !exists {
		return 0, fmt.Errorf("parameter %s not found", key)
	}

	switch v := value.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		if v == float64(int64(v)) {
			return int(v), nil
		}
		return 0, fmt.Errorf("parameter %s must be integer, got float with decimal", key)
	default:
		return 0, fmt.Errorf("parameter %s must be integer, got %T", key, value)
	}
}

// GetBool 获取布尔参数
func GetBool(args map[string]interface{}, key string) (bool, error) {
	value, exists := args[key]
	if !exists {
		return false, fmt.Errorf("parameter %s not found", key)
	}

	b, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("parameter %s must be boolean, got %T", key, value)
	}

	return b, nil
}

// GetStringWithDefault 获取字符串参数，如果不存在返回默认值
func GetStringWithDefault(args map[string]interface{}, key, defaultValue string) string {
	if value, err := GetString(args, key); err == nil {
		return value
	}
	return defaultValue
}
