package maputil

import (
	"strings"

	"github.com/gookit/goutil/mathutil"
	"github.com/gookit/goutil/strutil"
)

// Data an map data type
type Data map[string]any
type Map = Data

// Has value on the data map
func (d Data) Has(key string) bool {
	_, ok := d.GetByPath(key)
	return ok
}

// IsEmtpy if the data map
func (d Data) IsEmtpy() bool {
	return len(d) == 0
}

// Value get from the data map
func (d Data) Value(key string) (interface{}, bool) {
	val, ok := d.GetByPath(key)
	return val, ok
}

// Get value from the data map.
// Supports dot syntax to get deep values. eg: top.sub
func (d Data) Get(key string) interface{} {
	if val, ok := d.GetByPath(key); ok {
		return val
	}
	return nil
}

// GetByPath get value from the data map by path. eg: top.sub
// Supports dot syntax to get deep values.
func (d Data) GetByPath(path string) (interface{}, bool) {
	if val, ok := d[path]; ok {
		return val, true
	}

	// is key path.
	if strings.ContainsRune(path, '.') {
		val, ok := GetByPath(path, d)
		if ok {
			return val, true
		}
	}
	return nil, false
}

// Set value to the data map
func (d Data) Set(key string, val any) {
	d[key] = val
}

// SetByPath sets a value in the map.
// Supports dot syntax to set deep values.
//
// For example:
//
//	d.SetByPath("name.first", "Mat")
func (d Data) SetByPath(path string, value any) error {
	if path == "" {
		return nil
	}
	return d.SetByKeys(strings.Split(path, KeySepStr), value)
}

// SetByKeys sets a value in the map by path keys.
// Supports dot syntax to set deep values.
//
// For example:
//
//	d.SetByKeys([]string{"name", "first"}, "Mat")
func (d Data) SetByKeys(keys []string, value any) error {
	kln := len(keys)
	if kln == 0 {
		return nil
	}

	// special handle d is empty.
	if len(d) == 0 {
		if kln == 1 {
			d.Set(keys[0], value)
		} else {
			d.Set(keys[0], MakeByKeys(keys[1:], value))
		}
		return nil
	}

	return SetByKeys((*map[string]any)(&d), keys, value)
	// It's ok, but use `func (d *Data)`
	// return SetByKeys((*map[string]any)(d), keys, value)
}

// Default get value from the data map with default value
func (d Data) Default(key string, def any) interface{} {
	if val, ok := d.GetByPath(key); ok {
		return val
	}
	return def
}

// Int value get
func (d Data) Int(key string) int {
	if val, ok := d.GetByPath(key); ok {
		return mathutil.QuietInt(val)
	}
	return 0
}

// Int64 value get
func (d Data) Int64(key string) int64 {
	if val, ok := d.GetByPath(key); ok {
		return mathutil.QuietInt64(val)
	}
	return 0
}

// Str value get by key
func (d Data) Str(key string) string {
	if val, ok := d.GetByPath(key); ok {
		return strutil.QuietString(val)
	}
	return ""
}

// Bool value get
func (d Data) Bool(key string) bool {
	val, ok := d.GetByPath(key)
	if !ok {
		return false
	}
	if bl, ok := val.(bool); ok {
		return bl
	}

	if str, ok := val.(string); ok {
		return strutil.QuietBool(str)
	}
	return false
}

// Strings get []string value
func (d Data) Strings(key string) []string {
	val, ok := d.GetByPath(key)
	if !ok {
		return nil
	}

	if ss, ok := val.([]string); ok {
		return ss
	}
	return nil
}

// StrSplit get strings by split key value
func (d Data) StrSplit(key, sep string) []string {
	if val, ok := d.GetByPath(key); ok {
		return strings.Split(strutil.QuietString(val), sep)
	}
	return nil
}

// StringsByStr value get by key
func (d Data) StringsByStr(key string) []string {
	if val, ok := d.GetByPath(key); ok {
		return strings.Split(strutil.QuietString(val), ",")
	}
	return nil
}

// StringMap get map[string]string value
func (d Data) StringMap(key string) map[string]string {
	val, ok := d.GetByPath(key)
	if !ok {
		return nil
	}

	if smp, ok := val.(map[string]string); ok {
		return smp
	}
	return nil
}

// Sub get sub value as new Data
func (d Data) Sub(key string) Data {
	if val, ok := d.GetByPath(key); ok {
		if sub, ok := val.(map[string]any); ok {
			return sub
		}
	}
	return nil
}

// Keys of the data map
func (d Data) Keys() []string {
	keys := make([]string, 0, len(d))
	for k := range d {
		keys = append(keys, k)
	}
	return keys
}

// ToStringMap convert to map[string]string
func (d Data) ToStringMap() map[string]string {
	return ToStringMap(d)
}

// String data to string
func (d Data) String() string {
	return ToString(d)
}
