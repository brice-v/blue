package object

import (
	"fmt"
	"log"
	"regexp"

	"github.com/fxamacker/cbor/v2"
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

type ObjectWrapper struct {
	Type iType           `cbor:"type"`
	Data cbor.RawMessage `cbor:"data"`
}

func decodeFromType(t iType, data []byte) (Object, error) {
	switch t {
	case i_INTEGER_OBJ:
		var x *Integer
		diag("INTEGER", data)
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			return nil, err
		}
		return x, nil
	case i_LIST_OBJ:
		var cbws []ObjectWrapper
		diag("IN LIST", data)
		err := cbor.Unmarshal(data, &cbws)
		if err != nil {
			return nil, err
		}
		elems := make([]Object, len(cbws))
		for i, e := range cbws {
			diag("IN LOOP", e.Data)
			obj, err := decodeFromType(e.Type, e.Data)
			if err != nil {
				return nil, err
			}
			elems[i] = obj
		}
		return &List{Elements: elems}, nil
	case i_BIG_INTEGER_OBJ:
		var x *BigInteger
		diag("BIGINT", data)
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			return nil, err
		}
		return x, nil
	case i_FLOAT_OBJ:
		var x *Float
		diag("FLOAT", data)
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			return nil, err
		}
		return x, nil
	case i_BIG_FLOAT_OBJ:
		var x *BigFloat
		diag("BIGFLOAT", data)
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			return nil, err
		}
		return x, nil
	case i_BOOLEAN_OBJ:
		var x *Boolean
		diag("BOOL", data)
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			return nil, err
		}
		return x, nil
	case i_UINTEGER_OBJ:
		var x *UInteger
		diag("UINTEGER", data)
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			return nil, err
		}
		return x, nil
	case i_NULL_OBJ:
		var x *Null
		diag("NULL", data)
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			return nil, err
		}
		return x, nil
	case i_STRING_OBJ:
		var x *Stringo
		diag("STR", data)
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			return nil, err
		}
		return x, nil
	case i_REGEX_OBJ:
		var x string
		diag("REGEX", data)
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			return nil, err
		}
		return &Regex{Value: regexp.MustCompile(x)}, nil
	case i_BYTES_OBJ:
		var bs []byte
		diag("BYTES", data)
		err := cbor.Unmarshal(data, &bs)
		if err != nil {
			return nil, err
		}
		return &Bytes{Value: bs}, nil
	default:
		return nil, fmt.Errorf("decodeFromType: handle %d", t)
	}
}

func diag(prefix string, data []byte) {
	if s, err := cbor.Diagnose(data); err == nil {
		log.Printf("%s Diagnose %s", prefix, s)
	} else {
		log.Printf("%s Dianose error %s", prefix, err.Error())
	}
}

func Decode(data []byte) (Object, error) {
	var a ObjectWrapper
	diag("DECODE", data)
	err := cbor.Unmarshal(data, &a)
	if err != nil {
		return nil, err
	}
	return decodeFromType(a.Type, a.Data)
}

func marshalObject(obj Object) (ObjectWrapper, error) {
	data, err := cbor.Marshal(obj)
	if err != nil {
		return ObjectWrapper{}, err
	}
	return ObjectWrapper{
		Type: obj.IType(),
		Data: data,
	}, nil
}

func marshalObjectWrapper(obj Object) ([]byte, error) {
	o, err := marshalObject(obj)
	if err != nil {
		return nil, err
	}
	return cbor.Marshal(o)
}

func (x *Integer) Encode() ([]byte, error) {
	return marshalObjectWrapper(x)
}

func (x *Integer) IType() iType {
	return i_INTEGER_OBJ
}

func (x *BigInteger) Encode() ([]byte, error) {
	return marshalObjectWrapper(x)
}

func (x *BigInteger) IType() iType {
	return i_BIG_INTEGER_OBJ
}

func (x *Boolean) Encode() ([]byte, error) {
	return marshalObjectWrapper(x)
}

func (x *Boolean) IType() iType {
	return i_BOOLEAN_OBJ
}

func (x *Null) Encode() ([]byte, error) {
	return marshalObjectWrapper(x)
}

func (x *Null) IType() iType {
	return i_NULL_OBJ
}

func (x *UInteger) Encode() ([]byte, error) {
	return marshalObjectWrapper(x)
}

func (x *UInteger) IType() iType {
	return i_UINTEGER_OBJ
}

func (x *Float) Encode() ([]byte, error) {
	return marshalObjectWrapper(x)
}

func (x *Float) IType() iType {
	return i_FLOAT_OBJ
}

func (x BigFloat) Encode() ([]byte, error) {
	return marshalObjectWrapper(x)
}

func (x BigFloat) IType() iType {
	return i_BIG_FLOAT_OBJ
}

func (x *ReturnValue) Encode() ([]byte, error) {
	panic(fmt.Sprintf("encode handle %T?", x))
}

func (x *ReturnValue) IType() iType {
	return i_RETURN_VALUE_OBJ
}

func (x *Error) Encode() ([]byte, error) {
	panic(fmt.Sprintf("encode handle %T?", x))
}

func (x *Error) IType() iType {
	return i_ERROR_OBJ
}

func (x *Function) Encode() ([]byte, error) {
	panic(fmt.Sprintf("encode handle %T", x))
}

func (x *Function) IType() iType {
	return i_FUNCTION_OBJ
}

func (x *Stringo) Encode() ([]byte, error) {
	return marshalObjectWrapper(x)
}

func (x *Stringo) IType() iType {
	return i_STRING_OBJ
}

func (x *Bytes) Encode() ([]byte, error) {
	data, err := cbor.Marshal(x.Value)
	if err != nil {
		return nil, err
	}
	ow := ObjectWrapper{
		Type: x.IType(),
		Data: data,
	}
	return cbor.Marshal(ow)
}

func (x *Bytes) IType() iType {
	return i_BYTES_OBJ
}

func (x *GoObj[T]) Encode() ([]byte, error) {
	panic(fmt.Sprintf("encode handle %T", x))
}

func (x *GoObj[T]) IType() iType {
	return i_GO_OBJ
}

func (x *Regex) Encode() ([]byte, error) {
	res := x.Value.String()
	data, err := cbor.Marshal(res)
	if err != nil {
		return nil, err
	}
	ow := ObjectWrapper{
		Type: x.IType(),
		Data: data,
	}
	return cbor.Marshal(ow)
}

func (x *Regex) IType() iType {
	return i_REGEX_OBJ
}

func (x *List) Encode() ([]byte, error) {
	ows := make([]ObjectWrapper, len(x.Elements))
	for i, e := range x.Elements {
		cw, err := marshalObject(e)
		if err != nil {
			return nil, err
		}
		ows[i] = cw
	}
	data, err := cbor.Marshal(ows)
	if err != nil {
		return nil, err
	}
	o := ObjectWrapper{
		Type: x.IType(),
		Data: data,
	}
	return cbor.Marshal(o)
}

func (x *List) IType() iType {
	return i_LIST_OBJ
}

func (x *Map) Encode() ([]byte, error) {
	panic(fmt.Sprintf("encode handle %T", x))
}

func (x *Map) IType() iType {
	return i_MAP_OBJ
}

func (x *Set) Encode() ([]byte, error) {
	panic(fmt.Sprintf("encode handle %T", x))
}

func (x *Set) IType() iType {
	return i_SET_OBJ
}

func (x *Module) Encode() ([]byte, error) {
	panic(fmt.Sprintf("encode handle %T", x))
}

func (x *Module) IType() iType {
	return i_MODULE_OBJ
}

func (x *BreakStatement) Encode() ([]byte, error) {
	panic(fmt.Sprintf("encode handle %T", x))
}

func (x *BreakStatement) IType() iType {
	return i_BREAK_OBJ
}

func (x *ContinueStatement) Encode() ([]byte, error) {
	panic(fmt.Sprintf("encode handle %T", x))
}

func (x *ContinueStatement) IType() iType {
	return i_CONTINUE_OBJ
}

// The Objects Below cannot be encoded but are included to satisfy the Object interface

func (x *ListCompLiteral) Encode() ([]byte, error) {
	panic(fmt.Sprintf("%T cannot be encoded", x))
}

func (x *ListCompLiteral) IType() iType {
	return i_LIST_COMP_OBJ
}

func (x *MapCompLiteral) Encode() ([]byte, error) {
	panic(fmt.Sprintf("%T cannot be encoded", x))
}

func (x *MapCompLiteral) IType() iType {
	return i_MAP_COMP_OBJ
}

func (x *SetCompLiteral) Encode() ([]byte, error) {
	panic(fmt.Sprintf("%T cannot be encoded", x))
}

func (x *SetCompLiteral) IType() iType {
	return i_SET_COMP_OBJ
}

func (x *Builtin) Encode() ([]byte, error) {
	panic(fmt.Sprintf("%T cannot be encoded", x))
}

func (x *Builtin) IType() iType {
	return i_BUILTIN_OBJ
}

func (x *BuiltinObj) Encode() ([]byte, error) {
	panic(fmt.Sprintf("%T cannot be encoded", x))
}

func (x *BuiltinObj) IType() iType {
	return i_BUILTIN_OBJ
}

func (x *Process) Encode() ([]byte, error) {
	panic(fmt.Sprintf("%T cannot be encoded", x))
}

func (x *Process) IType() iType {
	return i_PROCESS_OBJ
}
