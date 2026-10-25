//go:build !wasm && static

package object

import "math/big"

func dbDuckdbUUIDBytesToString(_ []byte) (string, bool)          { return "", false }
func dbDuckdbDecimalBigInt(_ any) (*big.Int, bool)               { return nil, false }
func dbDuckdbDecimalFloat(_ any) (float64, bool)                 { return 0, false }
func dbDuckdbUnionValue(_ any) (any, bool)                       { return nil, false }
func dbDuckdbOrderedMap(_ any, _, _ dbColumnKind) (Object, bool) { return nil, false }
func dbDuckdbObject(_ any) (Object, bool)                        { return nil, false }
