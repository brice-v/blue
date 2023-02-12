package object

import "fmt"

// CreateBasicMapObject creates an object that looks like {'t': objType, 'v': objValue}
// This is currently being used for `spawn` and `db.open()` so that a unique return value
// is created (and it allows the functions defined for them to work)
func CreateBasicMapObject(objType string, objValue uint64) *Map {
	m := &Map{
		Pairs: NewPairsMap(),
	}
	typeKeyStr := &Stringo{Value: "t"}
	typeValueStr := &Stringo{Value: objType}

	typeKeyHashedKey := HashObject(typeKeyStr)
	typeKeyHashKey := HashKey{Type: STRING_OBJ, Value: typeKeyHashedKey}
	m.Pairs.Set(typeKeyHashKey, MapPair{
		Key:   typeKeyStr,
		Value: typeValueStr,
	})

	valueKeyStr := &Stringo{Value: "v"}
	valueValueStr := &UInteger{Value: objValue}

	valueKeyHashedKey := HashObject(valueKeyStr)
	valueKeyHashKey := HashKey{Type: STRING_OBJ, Value: valueKeyHashedKey}
	m.Pairs.Set(valueKeyHashKey, MapPair{
		Key:   valueKeyStr,
		Value: valueValueStr,
	})

	return m
}

func CreateMapObjectForGoMap(input OrderedMap2[string, Object]) *Map {
	m := &Map{
		Pairs: NewPairsMap(),
	}

	for _, k := range input.Keys {
		v, _ := input.Get(k)
		kStr := &Stringo{Value: k}

		hkStr := HashObject(kStr)
		hk := HashKey{Type: STRING_OBJ, Value: hkStr}
		m.Pairs.Set(hk, MapPair{
			Key:   kStr,
			Value: v,
		})
	}

	return m
}

func CreateObjectFromDbInterface(input interface{}) Object {
	switch x := input.(type) {
	case int64:
		return &Integer{Value: x}
	case string:
		return &Stringo{Value: x}
	case float64:
		return &Float{Value: x}
	case []byte:
		return &Bytes{Value: x}
	case bool:
		return &Boolean{Value: x}
	default:
		return nil
	}
}

func createHelpStringForObject(name, desc string, obj Object) string {
	return fmt.Sprintf("Help:    `%s` %s\nType:    '%s'\nInspect: %s", name, desc, obj.Type(), obj.Inspect())
}
