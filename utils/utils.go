package utils

import (
	"reflect"
	"time"
)

func RubyDate(date string) *time.Time {
	tm, err := time.Parse(time.RubyDate, date)
	if err != nil {
		return nil
	}
	return &tm
}

func IsMapKey(value interface{}, key string) (interface{}, bool) {
	if reflect.ValueOf(value).Kind() == reflect.Map {
		if newMap, ok := value.(map[string]interface{}); ok {
			if exist, ok := newMap[key]; ok {
				return exist, ok
			}
		}
	}
	return nil, false
}

func Map(value interface{}, key string) map[string]interface{} {
	if reflect.ValueOf(value).Kind() == reflect.Map {
		if newMap, ok := value.(map[string]interface{}); ok {
			if exist, ok := newMap[key]; ok {
				return exist.(map[string]interface{})
			}
		}
	}
	return nil
}

type DictType struct {
	data interface{}
}

func Dict(data interface{}) *DictType {
	return &DictType{data: data}
}

func (d *DictType) M(key string) *DictType {
	if reflect.ValueOf(d.data).Kind() == reflect.Map {
		return Dict(d.data.(map[string]interface{})[key])
	}
	return nil
}

func (d *DictType) Interface() interface{} {
	return d.data
}

func (d *DictType) Slice() []interface{} {
	return d.data.([]interface{})
}

func (d *DictType) MapInterface() map[string]interface{} {
	return d.data.(map[string]interface{})
}

func (d *DictType) Exists(key string) (*DictType, bool) {
	switch data := d.data.(type) {
	case map[string]interface{}:
		if v, ok := data[key]; ok {
			return Dict(v), ok
		}
	default:
		panic("value type is nil")
	}
	return nil, false
}

func (d *DictType) String() string {
	return d.data.(string)
}

func (d *DictType) Int() int {
	return d.data.(int)
}

func (d *DictType) Bool() bool {
	return d.data.(bool)
}
