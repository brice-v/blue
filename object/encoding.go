package object

import (
	"fmt"
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

type CBORWrapper struct {
	Type iType           `cbor:"type"`
	Data cbor.RawMessage `cbor:"data"`
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
	// i_LIST_OBJ: func(a any) Object {
	// 	bs := a.([]byte)

	// },
}

func decodeFromType(t iType, data []byte) (Object, error) {
	switch t {
	case i_INTEGER_OBJ:
		var x *Integer
		diag("INTEGER", data)
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			log.Printf("err here4 = %s data = %+#v", err.Error(), data)
			return nil, err
		}
		return x, nil
	case i_LIST_OBJ:
		var cbws []*CBORWrapper
		diag("IN LIST", data)
		err := cbor.Unmarshal(data, &cbws)
		log.Printf("data here -- %+#v", data)
		if err != nil {
			log.Printf("err here3 = %s", err.Error())
			return nil, err
		}
		elems := make([]Object, len(cbws))
		for i, e := range cbws {
			log.Printf("data here ----- %+#v", data)
			diag("IN LOOP", e.Data)
			// obj, err := Decode(e.Data)
			obj, err := decodeFromType(e.Type, e.Data)
			if err != nil {
				log.Printf("err here2 = %s, data = %+#v", err.Error(), data)
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
			log.Printf("err here7 = %s data = %+#v", err.Error(), data)
			return nil, err
		}
		return x, nil
	case i_FLOAT_OBJ:
		var x *Float
		diag("FLOAT", data)
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			log.Printf("err here8 = %s data = %+#v", err.Error(), data)
			return nil, err
		}
		return x, nil
	case i_BIG_FLOAT_OBJ:
		var x *BigFloat
		diag("BIGFLOAT", data)
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			log.Printf("err here9 = %s data = %+#v", err.Error(), data)
			return nil, err
		}
		return x, nil
	case i_BOOLEAN_OBJ:
		var x *Boolean
		diag("BOOL", data)
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			log.Printf("err here10 = %s data = %+#v", err.Error(), data)
			return nil, err
		}
		return x, nil
	case i_UINTEGER_OBJ:
		var x *UInteger
		diag("UINTEGER", data)
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			log.Printf("err here11 = %s data = %+#v", err.Error(), data)
			return nil, err
		}
		return x, nil
	case i_NULL_OBJ:
		var x *Null
		diag("NULL", data)
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			log.Printf("err here12 = %s data = %+#v", err.Error(), data)
			return nil, err
		}
		return x, nil
	case i_STRING_OBJ:
		var x *Stringo
		diag("STR", data)
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			log.Printf("err here12 = %s data = %+#v", err.Error(), data)
			return nil, err
		}
		return x, nil
	}
	panic("TODO")
}

func diag(prefix string, data []byte) {
	if s, err := cbor.Diagnose(data); err == nil {
		log.Printf("%s Diagnose %s", prefix, s)
	} else {
		log.Printf("%s Dianose error %s", prefix, err.Error())
	}
}

func Decode(data []byte) (Object, error) {
	var a CBORWrapper
	diag("DECODE", data)
	log.Printf("DATA HERE = %#+v", data)
	err := cbor.Unmarshal(data, &a)
	log.Printf("type = %d", a.Type)
	// if o, ok := a.(map[interface{}]interface{}); ok {
	// 	for k, v := range o {
	// 		log.Printf("k = %+#v, (%T), v = %+#v (%T)", k, k, v, v)
	// 	}
	// }
	// key := iType(a[0].(uint64))
	if err != nil {
		log.Printf("err here1 = %s", err.Error())
		return nil, err
	}
	return decodeFromType(a.Type, a.Data)
	// return &Null{}, nil
	// for x := range a {
	// 	log.Printf("x = %+#v, t = %T", x, x)
	// }
	// log.Printf("%T, %+#v", a, a)
	// return lookup[key](a[1]), nil
}

func ow(obj Object, value any) ObjectWrapper {
	return ObjectWrapper{
		Type:  obj.IType(),
		Value: value,
	}
}

var emptyCborWrapper = CBORWrapper{}

func cborWrapper(obj Object) (CBORWrapper, error) {
	data, err := cbor.Marshal(obj)
	if err != nil {
		log.Printf("err here = %s", err.Error())
		return emptyCborWrapper, err
	}
	return CBORWrapper{
		Type: obj.IType(),
		Data: data,
	}, nil
}

func (x *Integer) Encode() ([]byte, error) {
	cw, err := cborWrapper(x)
	if err != nil {
		return nil, err
	}
	return cbor.Marshal(cw)
	// return cbor.Marshal(ow(x, x.Value))
	// return nil, fmt.Errorf("TODO")
}

func (x *Integer) IType() iType {
	return i_INTEGER_OBJ
}

func (x *BigInteger) Encode() ([]byte, error) {
	// return cbor.Marshal(ow(x, x.Value.String()))
	// return nil, fmt.Errorf("TODO")
	cw, err := cborWrapper(x)
	if err != nil {
		return nil, err
	}
	return cbor.Marshal(cw)
}

func (x *BigInteger) IType() iType {
	return i_BIG_INTEGER_OBJ
}

func (x *Boolean) Encode() ([]byte, error) {
	// return cbor.Marshal(ow(x, x.Value))
	// return nil, fmt.Errorf("TODO")
	cw, err := cborWrapper(x)
	if err != nil {
		return nil, err
	}
	return cbor.Marshal(cw)
}

func (x *Boolean) IType() iType {
	return i_BOOLEAN_OBJ
}

func (x *Null) Encode() ([]byte, error) {
	// return cbor.Marshal(ow(x, nil))
	// return nil, fmt.Errorf("TODO")
	cw, err := cborWrapper(x)
	if err != nil {
		return nil, err
	}
	return cbor.Marshal(cw)
}

func (x *Null) IType() iType {
	return i_NULL_OBJ
}

func (x *UInteger) Encode() ([]byte, error) {
	// return cbor.Marshal(ow(x, x.Value))
	// return nil, fmt.Errorf("TODO")
	cw, err := cborWrapper(x)
	if err != nil {
		return nil, err
	}
	return cbor.Marshal(cw)
}

func (x *UInteger) IType() iType {
	return i_UINTEGER_OBJ
}

func (x *Float) Encode() ([]byte, error) {
	// return cbor.Marshal(ow(x, x.Value))
	// return nil, fmt.Errorf("TODO")
	cw, err := cborWrapper(x)
	if err != nil {
		return nil, err
	}
	return cbor.Marshal(cw)
}

func (x *Float) IType() iType {
	return i_FLOAT_OBJ
}

func (x BigFloat) Encode() ([]byte, error) {
	// return cbor.Marshal(ow(x, x.Value.String()))
	// return nil, fmt.Errorf("TODO")
	cw, err := cborWrapper(x)
	if err != nil {
		return nil, err
	}
	return cbor.Marshal(cw)
}

func (x BigFloat) IType() iType {
	return i_BIG_FLOAT_OBJ
}

func (x *ReturnValue) Encode() ([]byte, error) {
	// return cbor.Marshal(ow(x, x.Value))
	return nil, fmt.Errorf("TODO")
}

func (x *ReturnValue) IType() iType {
	return i_RETURN_VALUE_OBJ
}

func (x *Error) Encode() ([]byte, error) {
	// return cbor.Marshal(ow(x, x.Message))
	return nil, fmt.Errorf("TODO")
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
	// return cbor.Marshal(ow(x, x.Value))
	// return nil, fmt.Errorf("TODO")
	cw, err := cborWrapper(x)
	if err != nil {
		return nil, err
	}
	return cbor.Marshal(cw)
}

func (x *Stringo) IType() iType {
	return i_STRING_OBJ
}

func (x *Bytes) Encode() ([]byte, error) {
	// return cbor.Marshal(ow(x, x.Value))
	// return nil, fmt.Errorf("TODO")
	cw, err := cborWrapper(x)
	if err != nil {
		return nil, err
	}
	return cbor.Marshal(cw)
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
	// return cbor.Marshal(ow(x, x.Value))
	return nil, fmt.Errorf("TODO")
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
	// var buf bytes.Buffer
	// for _, e := range x.Elements {
	// 	bs, err := e.Encode()
	// 	if err != nil {
	// 		return nil, fmt.Errorf("list encode error: %w", err)
	// 	}
	// 	_, err = buf.Write(bs)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("list encode error: %w", err)
	// 	}
	// }
	// TODO: Should be smarter about that?
	// TODO: Should call encode on each element?
	// return cbor.Marshal(ow(x, x.Elements))
	cws := make([]CBORWrapper, len(x.Elements))
	for i, e := range x.Elements {
		cw, err := cborWrapper(e)
		if err != nil {
			log.Printf("err here5 = %s", err.Error())
			return nil, err
		}
		cws[i] = cw
	}
	data, err := cbor.Marshal(cws)
	if err != nil {
		log.Printf("err6 here = %s", err.Error())
		return nil, err
	}
	cw := CBORWrapper{
		Type: x.IType(),
		Data: data,
	}
	return cbor.Marshal(cw)
	// return nil, fmt.Errorf("TODO")
	// return cbor.Marshal(ow(x, buf.Bytes()))
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
