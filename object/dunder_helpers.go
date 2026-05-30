package object

type DunderType int

const (
	DunderInvalid DunderType = iota
	DunderStr
	DunderAdd
	DunderSub
	DunderMul
	DunderDiv
	DunderMod
	DunderFdiv
	DunderPow
)

var (
	_dunderStr        = &Stringo{Value: "__str"}
	_hashedDunderStr  = HashObject(_dunderStr)
	_dunderStrHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderStr}

	_dunderAdd        = &Stringo{Value: "__add"}
	_hashedDunderAdd  = HashObject(_dunderAdd)
	_dunderAddHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderAdd}

	_dunderSub        = &Stringo{Value: "__sub"}
	_hashedDunderSub  = HashObject(_dunderSub)
	_dunderSubHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderSub}

	_dunderMul        = &Stringo{Value: "__mul"}
	_hashedDunderMul  = HashObject(_dunderMul)
	_dunderMulHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderMul}

	_dunderDiv        = &Stringo{Value: "__div"}
	_hashedDunderDiv  = HashObject(_dunderDiv)
	_dunderDivHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderDiv}

	_dunderMod        = &Stringo{Value: "__mod"}
	_hashedDunderMod  = HashObject(_dunderMod)
	_dunderModHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderMod}

	_dunderFdiv        = &Stringo{Value: "__fdiv"}
	_hashedDunderFdiv  = HashObject(_dunderFdiv)
	_dunderFdivHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderFdiv}

	_dunderPow        = &Stringo{Value: "__pow"}
	_hashedDunderPow  = HashObject(_dunderPow)
	_dunderPowHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderPow}
)

func getDunderHashKey(t DunderType) *HashKey {
	switch t {
	case DunderInvalid:
		return nil
	case DunderStr:
		return &_dunderStrHashKey
	case DunderAdd:
		return &_dunderAddHashKey
	case DunderSub:
		return &_dunderSubHashKey
	case DunderMul:
		return &_dunderMulHashKey
	case DunderDiv:
		return &_dunderDivHashKey
	case DunderMod:
		return &_dunderModHashKey
	case DunderFdiv:
		return &_dunderFdivHashKey
	case DunderPow:
		return &_dunderPowHashKey
	default:
		return nil
	}
}

func HasDunderFun(t DunderType, o Object) (*Closure, bool) {
	if o == nil {
		return nil, false
	}
	m, ok := o.(*Map)
	if !ok {
		return nil, false
	}
	hk := getDunderHashKey(t)
	if hk == nil {
		return nil, false
	}
	mp, ok := m.Pairs.Get(*hk)
	if !ok {
		return nil, false
	}
	fn, ok := mp.Value.(*Closure)
	if !ok {
		return nil, false
	}
	return fn, true
}
