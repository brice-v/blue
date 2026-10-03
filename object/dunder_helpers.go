package object

type DunderType int

const (
	DunderInvalid DunderType = iota
	DunderStr
	DunderAdd
	DunderRAdd
	DunderSub
	DunderRSub
	DunderMul
	DunderRMul
	DunderDiv
	DunderRDiv
	DunderMod
	DunderRMod
	DunderFdiv
	DunderFRdiv
	DunderPow
	DunderRPow
	DunderAnd
	DunderRAnd
	DunderOr
	DunderROr
	DunderXor
	DunderRXor
	DunderRshift
	DunderRRshift
	DunderLshift
	DunderRLshift
	DunderMatmul
	DunderRMatmul
	DunderNeg
	DunderInv
	DunderEq
	DunderNotEq
	DunderGt
	DunderRGt
	DunderGte
	DunderRGte
	DunderGet
	DunderLen
	DunderTType
)

func (dt DunderType) GetRightVariant() DunderType {
	switch dt {
	case DunderAdd:
		return DunderRAdd
	case DunderSub:
		return DunderRSub
	case DunderMul:
		return DunderRMul
	case DunderDiv:
		return DunderRDiv
	case DunderMod:
		return DunderRMod
	case DunderFdiv:
		return DunderFRdiv
	case DunderPow:
		return DunderRPow
	case DunderAnd:
		return DunderRAnd
	case DunderOr:
		return DunderROr
	case DunderXor:
		return DunderRXor
	case DunderRshift:
		return DunderRRshift
	case DunderLshift:
		return DunderRLshift
	case DunderMatmul:
		return DunderRMatmul
	case DunderGt:
		return DunderRGt
	case DunderGte:
		return DunderRGte
	}
	return DunderInvalid
}

var (
	_dunderStr        = &Stringo{Value: "__str"}
	_hashedDunderStr  = HashObject(_dunderStr)
	_dunderStrHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderStr}

	_dunderAdd        = &Stringo{Value: "__add"}
	_hashedDunderAdd  = HashObject(_dunderAdd)
	_dunderAddHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderAdd}

	_dunderRAdd        = &Stringo{Value: "__radd"}
	_hashedDunderRAdd  = HashObject(_dunderRAdd)
	_dunderRAddHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderRAdd}

	_dunderSub        = &Stringo{Value: "__sub"}
	_hashedDunderSub  = HashObject(_dunderSub)
	_dunderSubHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderSub}

	_dunderRSub        = &Stringo{Value: "__rsub"}
	_hashedDunderRSub  = HashObject(_dunderRSub)
	_dunderRSubHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderRSub}

	_dunderMul        = &Stringo{Value: "__mul"}
	_hashedDunderMul  = HashObject(_dunderMul)
	_dunderMulHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderMul}

	_dunderRMul        = &Stringo{Value: "__rmul"}
	_hashedDunderRMul  = HashObject(_dunderRMul)
	_dunderRMulHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderRMul}

	_dunderDiv        = &Stringo{Value: "__div"}
	_hashedDunderDiv  = HashObject(_dunderDiv)
	_dunderDivHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderDiv}

	_dunderRDiv        = &Stringo{Value: "__rdiv"}
	_hashedDunderRDiv  = HashObject(_dunderRDiv)
	_dunderRDivHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderRDiv}

	_dunderMod        = &Stringo{Value: "__mod"}
	_hashedDunderMod  = HashObject(_dunderMod)
	_dunderModHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderMod}

	_dunderRMod        = &Stringo{Value: "__rmod"}
	_hashedDunderRMod  = HashObject(_dunderRMod)
	_dunderRModHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderRMod}

	_dunderFdiv        = &Stringo{Value: "__fdiv"}
	_hashedDunderFdiv  = HashObject(_dunderFdiv)
	_dunderFdivHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderFdiv}

	_dunderRFdiv        = &Stringo{Value: "__rfdiv"}
	_hashedDunderRFdiv  = HashObject(_dunderRFdiv)
	_dunderRFdivHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderRFdiv}

	_dunderPow        = &Stringo{Value: "__pow"}
	_hashedDunderPow  = HashObject(_dunderPow)
	_dunderPowHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderPow}

	_dunderRPow        = &Stringo{Value: "__rpow"}
	_hashedDunderRPow  = HashObject(_dunderRPow)
	_dunderRPowHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderRPow}

	_dunderAnd        = &Stringo{Value: "__and"}
	_hashedDunderAnd  = HashObject(_dunderAnd)
	_dunderAndHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderAnd}

	_dunderRAnd        = &Stringo{Value: "__rand"}
	_hashedDunderRAnd  = HashObject(_dunderRAnd)
	_dunderRAndHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderRAnd}

	_dunderOr        = &Stringo{Value: "__or"}
	_hashedDunderOr  = HashObject(_dunderOr)
	_dunderOrHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderOr}

	_dunderROr        = &Stringo{Value: "__ror"}
	_hashedDunderROr  = HashObject(_dunderROr)
	_dunderROrHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderROr}

	_dunderXor        = &Stringo{Value: "__xor"}
	_hashedDunderXor  = HashObject(_dunderXor)
	_dunderXorHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderXor}

	_dunderRXor        = &Stringo{Value: "__rxor"}
	_hashedDunderRXor  = HashObject(_dunderRXor)
	_dunderRXorHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderRXor}

	_dunderRshift        = &Stringo{Value: "__rshift"}
	_hashedDunderRshift  = HashObject(_dunderRshift)
	_dunderRshiftHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderRshift}

	_dunderRRshift        = &Stringo{Value: "__rrshift"}
	_hashedDunderRRshift  = HashObject(_dunderRRshift)
	_dunderRRshiftHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderRRshift}

	_dunderLshift        = &Stringo{Value: "__lshift"}
	_hashedDunderLshift  = HashObject(_dunderLshift)
	_dunderLshiftHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderLshift}

	_dunderRLshift        = &Stringo{Value: "__rlshift"}
	_hashedDunderRLshift  = HashObject(_dunderRLshift)
	_dunderRLshiftHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderRLshift}

	_dunderMatmul        = &Stringo{Value: "__matmul"}
	_hashedDunderMatmul  = HashObject(_dunderMatmul)
	_dunderMatmulHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderMatmul}

	_dunderRMatmul        = &Stringo{Value: "__rmatmul"}
	_hashedDunderRMatmul  = HashObject(_dunderRMatmul)
	_dunderRMatmulHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderRMatmul}

	_dunderNeg        = &Stringo{Value: "__neg"}
	_hashedDunderNeg  = HashObject(_dunderNeg)
	_dunderNegHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderNeg}

	_dunderInv        = &Stringo{Value: "__inv"}
	_hashedDunderInv  = HashObject(_dunderInv)
	_dunderInvHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderInv}

	_dunderEq        = &Stringo{Value: "__eq"}
	_hashedDunderEq  = HashObject(_dunderEq)
	_dunderEqHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderEq}

	_dunderNotEq        = &Stringo{Value: "__ne"}
	_hashedDunderNotEq  = HashObject(_dunderNotEq)
	_dunderNotEqHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderNotEq}

	_dunderGt        = &Stringo{Value: "__gt"}
	_hashedDunderGt  = HashObject(_dunderGt)
	_dunderGtHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderGt}

	_dunderRGt        = &Stringo{Value: "__rgt"}
	_hashedDunderRGt  = HashObject(_dunderRGt)
	_dunderRGtHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderRGt}

	_dunderGte        = &Stringo{Value: "__gte"}
	_hashedDunderGte  = HashObject(_dunderGte)
	_dunderGteHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderGte}

	_dunderRGte        = &Stringo{Value: "__rgte"}
	_hashedDunderRGte  = HashObject(_dunderRGte)
	_dunderRGteHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderRGte}

	_dunderGet        = &Stringo{Value: "__get"}
	_hashedDunderGet  = HashObject(_dunderGet)
	_dunderGetHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderGet}

	_dunderLen        = &Stringo{Value: "__len"}
	_hashedDunderLen  = HashObject(_dunderLen)
	_dunderLenHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderLen}

	_dunderTType        = &Stringo{Value: "__type"}
	_hashedDunderTType  = HashObject(_dunderTType)
	_dunderTTypeHashKey = HashKey{Type: STRING_OBJ, Value: _hashedDunderTType}
)

func getDunderHashKey(t DunderType) *HashKey {
	switch t {
	case DunderInvalid:
		return nil
	case DunderStr:
		return &_dunderStrHashKey
	case DunderAdd:
		return &_dunderAddHashKey
	case DunderRAdd:
		return &_dunderRAddHashKey
	case DunderSub:
		return &_dunderSubHashKey
	case DunderRSub:
		return &_dunderRSubHashKey
	case DunderMul:
		return &_dunderMulHashKey
	case DunderRMul:
		return &_dunderRMulHashKey
	case DunderDiv:
		return &_dunderDivHashKey
	case DunderRDiv:
		return &_dunderRDivHashKey
	case DunderMod:
		return &_dunderModHashKey
	case DunderRMod:
		return &_dunderRModHashKey
	case DunderFdiv:
		return &_dunderFdivHashKey
	case DunderFRdiv:
		return &_dunderRFdivHashKey
	case DunderPow:
		return &_dunderPowHashKey
	case DunderRPow:
		return &_dunderRPowHashKey
	case DunderAnd:
		return &_dunderAndHashKey
	case DunderRAnd:
		return &_dunderRAndHashKey
	case DunderOr:
		return &_dunderOrHashKey
	case DunderROr:
		return &_dunderROrHashKey
	case DunderXor:
		return &_dunderXorHashKey
	case DunderRXor:
		return &_dunderRXorHashKey
	case DunderRshift:
		return &_dunderRshiftHashKey
	case DunderRRshift:
		return &_dunderRRshiftHashKey
	case DunderLshift:
		return &_dunderLshiftHashKey
	case DunderRLshift:
		return &_dunderRLshiftHashKey
	case DunderMatmul:
		return &_dunderMatmulHashKey
	case DunderRMatmul:
		return &_dunderRMatmulHashKey
	case DunderNeg:
		return &_dunderNegHashKey
	case DunderInv:
		return &_dunderInvHashKey
	case DunderEq:
		return &_dunderEqHashKey
	case DunderNotEq:
		return &_dunderNotEqHashKey
	case DunderGt:
		return &_dunderGtHashKey
	case DunderRGt:
		return &_dunderRGtHashKey
	case DunderGte:
		return &_dunderGteHashKey
	case DunderRGte:
		return &_dunderRGteHashKey
	case DunderGet:
		return &_dunderGetHashKey
	case DunderLen:
		return &_dunderLenHashKey
	case DunderTType:
		return &_dunderTTypeHashKey
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
