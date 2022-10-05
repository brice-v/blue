package object

// CreateBasicMapObject creates an object that looks like {'t': objType, 'v': objValue}
// This is currently being used for `spawn` and `db.open()` so that a unique return value
// is created (and it allows the functions defined for them to work)
func CreateBasicMapObject(objType string, objValue int64) *Map {
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
	valueValueStr := &Integer{Value: objValue}

	valueKeyHashedKey := HashObject(valueKeyStr)
	valueKeyHashKey := HashKey{Type: STRING_OBJ, Value: valueKeyHashedKey}
	m.Pairs.Set(valueKeyHashKey, MapPair{
		Key:   valueKeyStr,
		Value: valueValueStr,
	})

	return m
}
