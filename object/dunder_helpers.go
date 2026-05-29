package object

type DunderType int

const (
	DunderInvalid DunderType = iota
	DunderStr
	DunderAdd
)

var (
	_dunderStr        = &Stringo{Value: "__str"}
	_hashedDunderStr  = HashObject(_dunderStr)
	_dunderStrHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderStr}

	_dunderAdd        = &Stringo{Value: "__add"}
	_hashedDunderAdd  = HashObject(_dunderAdd)
	_dunderAddHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderAdd}
)

func getDunderHashKey(t DunderType) *HashKey {
	switch t {
	case DunderInvalid:
		return nil
	case DunderStr:
		return &_dunderStrHashKey
	case DunderAdd:
		return &_dunderAddHashKey
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
