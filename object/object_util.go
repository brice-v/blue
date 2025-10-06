package object

import (
	"fmt"
)

// CreateBasicMapObject creates an object that looks like {'t': objType, 'v': objValue}
// This is currently being used for `spawn` and `db.open()` so that a unique return value
// is created (and it allows the functions defined for them to work)
func CreateBasicMapObject(objType string, objValue uint64) *Map {
	m := &Map{
		Pairs: NewPairsMapWithSize(1),
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

func CreateBasicMapObjectForGoObj[T any](objType string, goObj *GoObj[T]) *Map {
	m := &Map{
		Pairs: NewPairsMapWithSize(1),
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

	valueKeyHashedKey := HashObject(valueKeyStr)
	valueKeyHashKey := HashKey{Type: STRING_OBJ, Value: valueKeyHashedKey}
	m.Pairs.Set(valueKeyHashKey, MapPair{
		Key:   valueKeyStr,
		Value: goObj,
	})

	return m
}

func CreateMapObjectForGoMap(input OrderedMap2[string, Object]) *Map {
	m := &Map{
		Pairs: NewPairsMapWithSize(len(input.Keys)),
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

func IsCollectionType(t Type) bool {
	return t == LIST_OBJ || t == SET_OBJ || t == MAP_OBJ
}

// newError is the wrapper function to add an error to the evaluator
func newError(format string, a ...any) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func newPositionalTypeError(funName string, pos int, expectedType Type, currentType Type) *Error {
	return newError("PositionalTypeError: `%s` expects argument %d to be %s. got=%s", funName, pos, expectedType, currentType)
}

func newPositionalTypeErrorForGoObj(funName string, pos int, expectedType Type, currentObj any) *Error {
	return newError("PositionalTypeError: `%s` expects argument %d to be %s. got=%T", funName, pos, expectedType, currentObj)
}

func newInvalidArgCountError(funName string, got, want int, otherCount string) *Error {
	if otherCount == "" {
		return newError("InvalidArgCountError: `%s` wrong number of args. got=%d, want=%d", funName, got, want)
	}
	return newError("InvalidArgCountError: `%s` wrong number of args. got=%d, want=%d %s", funName, got, want, otherCount)
}
