//go:build !wasm && !static

package object

import (
	"math/big"

	"github.com/duckdb/duckdb-go/v2"
	"github.com/shopspring/decimal"
)

func dbDuckdbUUIDBytesToString(b []byte) (string, bool) {
	if len(b) != 16 {
		return "", false
	}
	var id duckdb.UUID
	copy(id[:], b)
	return id.String(), true
}

func dbDuckdbDecimalBigInt(input any) (*big.Int, bool) {
	x, ok := input.(duckdb.Decimal)
	if !ok {
		return nil, false
	}
	return new(big.Int).Set(x.Value), true
}

func dbDuckdbDecimalFloat(input any) (float64, bool) {
	x, ok := input.(duckdb.Decimal)
	if !ok {
		return 0, false
	}
	return x.Float64(), true
}

func dbDuckdbUnionValue(input any) (any, bool) {
	x, ok := input.(duckdb.Union)
	if !ok {
		return nil, false
	}
	return x.Value, true
}

func dbDuckdbOrderedMap(input any, kk, vv dbColumnKind) (Object, bool) {
	x, ok := input.(duckdb.OrderedMap)
	if !ok {
		return nil, false
	}
	m := &Map{Pairs: NewPairsMapWithSize(x.Len())}
	for _, k := range x.Keys() {
		kAsObj := createObjectFromDbInterface(k, kk)
		v, _ := x.Get(k)
		vAsObj := createObjectFromDbInterface(v, vv)
		hk := HashKey{Type: kAsObj.Type(), Value: HashObject(kAsObj)}
		m.Pairs.Set(hk, MapPair{Key: kAsObj, Value: vAsObj})
	}
	return m, true
}

func dbDuckdbObject(input any) (Object, bool) {
	switch x := input.(type) {
	case duckdb.Decimal:
		// TODO: This is potentially inefficient
		bf, err := decimal.NewFromString(x.String())
		if err != nil {
			return newError("decimal conversion error: %s", err.Error()), true
		}
		return &BigFloat{Value: bf}, true
	case duckdb.OrderedMap:
		return dbDuckdbOrderedMap(x, dbColumnKind{base: dbColUnknown}, dbColumnKind{base: dbColUnknown})
	case duckdb.Union:
		return createObjectFromDbInterface(x.Value, dbColumnKind{base: dbColUnknown}), true
	case duckdb.Interval:
		m := NewOrderedMapWithSize[string, Object](3)
		m.Set("days", NewInteger(int64(x.Days)))
		m.Set("micros", NewInteger(x.Micros))
		m.Set("months", NewInteger(int64(x.Months)))
		m.Set("t", NewString("duckdb.Interval"))
		return CreateMapObjectForGoMap(*m), true
	}
	return nil, false
}
