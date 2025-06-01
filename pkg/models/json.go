package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSON 是一个自定义类型，用于支持GORM中的JSON字段
// 它可以将结构体、map等类型存储为数据库中的JSON字段
type JSON json.RawMessage

// Value 实现 driver.Valuer 接口，将JSON转换为数据库可存储的值
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

// Scan 实现 sql.Scanner 接口，将数据库值转换为JSON
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = JSON("[]")
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("invalid scan source for JSON")
	}

	*j = JSON(bytes)
	return nil
}

// MarshalJSON 实现json.Marshaler接口
func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("[]"), nil
	}
	return j, nil
}

// UnmarshalJSON 实现json.Unmarshaler接口
func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("JSON: UnmarshalJSON on nil pointer")
	}
	*j = JSON(data)
	return nil
}
