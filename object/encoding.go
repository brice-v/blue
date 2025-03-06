package object

import (
	"log"
	"math/big"

	"github.com/fxamacker/cbor/v2"
	"github.com/shopspring/decimal"
)

// iType is the object type represented as an integer
type iType int

const (
	i_INTEGER_OBJ iType = iota
	i_BIG_INTEGER_OBJ
	i_BOOLEAN_OBJ
	i_NULL_OBJ
	i_UINTEGER_OBJ
	i_FLOAT_OBJ
	i_BIG_FLOAT_OBJ
	i_RETURN_VALUE_OBJ
	i_ERROR_OBJ
	i_FUNCTION_OBJ
	i_STRING_OBJ
	i_BYTES_OBJ
	i_GO_OBJ
	i_REGEX_OBJ
	i_BUILTIN_OBJ
	i_LIST_OBJ
	i_MAP_OBJ
	i_SET_OBJ
	i_LIST_COMP_OBJ
	i_MAP_COMP_OBJ
	i_SET_COMP_OBJ
	i_MODULE_OBJ
	i_PROCESS_OBJ

	i_BREAK_OBJ
	i_CONTINUE_OBJ
)

type Decoder interface {
	Decode(data []byte) (Object, error)
}

type ObjectWrapper struct {
	_     struct{} `cbor:",toarray"`
	Type  iType
	Value any
}

var lookup = map[iType]func(a any) Object{
	i_INTEGER_OBJ: func(a any) Object {
		return &Integer{Value: int64(a.(uint64))}
	},
	i_BIG_INTEGER_OBJ: func(a any) Object {
		bi := big.NewInt(0)
		bi.SetString(a.(string), 10)
		return &BigInteger{Value: bi}
	},
	i_FLOAT_OBJ: func(a any) Object {
		return &Float{Value: a.(float64)}
	},
	i_BIG_FLOAT_OBJ: func(a any) Object {
		if d, err := decimal.NewFromString(a.(string)); err == nil {
			return &BigFloat{Value: d}
		} else {
			panic(err.Error())
		}
	},
	i_BOOLEAN_OBJ: func(a any) Object {
		return &Boolean{Value: a.(bool)}
	},
	i_UINTEGER_OBJ: func(a any) Object {
		return &UInteger{Value: a.(uint64)}
	},
	i_NULL_OBJ: func(a any) Object {
		return &Null{}
	},
}

func Decode(data []byte) (Object, error) {
	var a [2]any
	if s, err := cbor.Diagnose(data); err == nil {
		log.Printf("Diagnose %s", s)
	} else {
		log.Printf("Dianose error %s", err.Error())
	}
	err := cbor.Unmarshal(data, &a)
	key := iType(a[0].(uint64))
	if err != nil {
		return nil, err
	}
	for x := range a {
		log.Printf("x = %+#v, t = %T", x, x)
	}
	log.Printf("%T, %+#v", a, a)
	return lookup[key](a[1]), nil
}

func ow(obj Object, value any) ObjectWrapper {
	return ObjectWrapper{
		Type:  obj.IType(),
		Value: value,
	}
}

func (x *Integer) Encode() ([]byte, error) {
	return cbor.Marshal(ow(x, x.Value))
}

func (x *Integer) IType() iType {
	return i_INTEGER_OBJ
}

func (x *BigInteger) Encode() ([]byte, error) {
	return cbor.Marshal(ow(x, x.Value.String()))
}

func (x *BigInteger) IType() iType {
	return i_BIG_INTEGER_OBJ
}

func (x *Boolean) Encode() ([]byte, error) {
	return cbor.Marshal(ow(x, x.Value))
}

func (x *Boolean) IType() iType {
	return i_BOOLEAN_OBJ
}

func (x *Null) Encode() ([]byte, error) {
	return cbor.Marshal(ow(x, nil))
}

func (x *Null) IType() iType {
	return i_NULL_OBJ
}

func (x *UInteger) Encode() ([]byte, error) {
	return cbor.Marshal(ow(x, x.Value))
}

func (x *UInteger) IType() iType {
	return i_UINTEGER_OBJ
}

func (x *Float) Encode() ([]byte, error) {
	return cbor.Marshal(ow(x, x.Value))
}

func (x *Float) IType() iType {
	return i_FLOAT_OBJ
}

func (x BigFloat) Encode() ([]byte, error) {
	return cbor.Marshal(ow(x, x.Value.String()))
}

func (x BigFloat) IType() iType {
	return i_BIG_FLOAT_OBJ
}

func (x *ReturnValue) Encode() ([]byte, error) {
	return cbor.Marshal(ow(x, x.Value))
}

func (x *ReturnValue) IType() iType {
	return i_RETURN_VALUE_OBJ
}

func (x *Error) Encode() ([]byte, error) {
	return cbor.Marshal(ow(x, x.Message))
}

func (x *Error) IType() iType {
	return i_ERROR_OBJ
}

func (x *Function) Encode() ([]byte, error) {
	panic("TODO")
}

func (x *Function) IType() iType {
	return i_FUNCTION_OBJ
}

func (x *Process) Encode() ([]byte, error) {
	panic("TODO")
}

func (x *Process) IType() iType {
	return i_PROCESS_OBJ
}

func (x *Stringo) Encode() ([]byte, error) {
	return cbor.Marshal(ow(x, x.Value))
}

func (x *Stringo) IType() iType {
	return i_STRING_OBJ
}

func (x *Bytes) Encode() ([]byte, error) {
	return cbor.Marshal(ow(x, x.Value))
}

func (x *Bytes) IType() iType {
	return i_BYTES_OBJ
}

func (x *GoObj[T]) Encode() ([]byte, error) {
	panic("TODO")
}

func (x *GoObj[T]) IType() iType {
	return i_GO_OBJ
}

func (x *Regex) Encode() ([]byte, error) {
	return cbor.Marshal(ow(x, x.Value))
}

func (x *Regex) IType() iType {
	return i_REGEX_OBJ
}

func (x *Builtin) Encode() ([]byte, error) {
	panic("TODO")
}

func (x *Builtin) IType() iType {
	return i_BUILTIN_OBJ
}

func (x *BuiltinObj) Encode() ([]byte, error) {
	panic("TODO")
}

func (x *BuiltinObj) IType() iType {
	// TODO: Should IType return different type for this object?
	return i_BUILTIN_OBJ
}

func (x *List) Encode() ([]byte, error) {
	// TODO: Should be smarter about that?
	// TODO: Should call encode on each element?
	return cbor.Marshal(ow(x, x.Elements))
}

func (x *List) IType() iType {
	return i_LIST_OBJ
}

func (x *ListCompLiteral) Encode() ([]byte, error) {
	return nil, nil
}

func (x *ListCompLiteral) IType() iType {
	return i_LIST_COMP_OBJ
}

func (x *Map) Encode() ([]byte, error) {
	return nil, nil
}

func (x *Map) IType() iType {
	return i_MAP_OBJ
}

func (x *MapCompLiteral) Encode() ([]byte, error) {
	return nil, nil
}

func (x *MapCompLiteral) IType() iType {
	return i_MAP_COMP_OBJ
}

func (x *Set) Encode() ([]byte, error) {
	return nil, nil
}

func (x *Set) IType() iType {
	return i_SET_OBJ
}

func (x *SetCompLiteral) Encode() ([]byte, error) {
	return nil, nil
}

func (x *SetCompLiteral) IType() iType {
	return i_SET_COMP_OBJ
}

func (x *Module) Encode() ([]byte, error) {
	return nil, nil
}

func (x *Module) IType() iType {
	return i_MODULE_OBJ
}

func (x *BreakStatement) Encode() ([]byte, error) {
	return nil, nil
}

func (x *BreakStatement) IType() iType {
	return i_BREAK_OBJ
}

func (x *ContinueStatement) Encode() ([]byte, error) {
	return nil, nil
}

func (x *ContinueStatement) IType() iType {
	return i_CONTINUE_OBJ
}
