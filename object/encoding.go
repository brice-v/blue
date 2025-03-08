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
	case i_LIST_OBJ:
		var ows []ObjectWrapper
		diag("IN LIST", data)
		err := cbor.Unmarshal(data, &ows)
		if err != nil {
			return nil, err
		}
		elems := make([]Object, len(ows))
		for i, e := range ows {
			diag("IN LOOP", e.Data)
			obj, err := decodeFromType(e.Type, e.Data)
			if err != nil {
				return nil, err
			}
			elems[i] = obj
		}
		return &List{Elements: elems}, nil
	case i_SET_OBJ:
		var ows []ObjectWrapper
		diag("IN SET", data)
		err := cbor.Unmarshal(data, &ows)
		if err != nil {
			return nil, err
		}
		elems := NewSetElementsWithSize(len(ows))
		for _, e := range ows {
			diag("IN SET LOOP", e.Data)
			obj, err := decodeFromType(e.Type, e.Data)
			if err != nil {
				return nil, err
			}
			hashKey := HashObject(obj)
			elems.Set(hashKey, SetPair{Value: obj, Present: struct{}{}})
		}
		return &Set{Elements: elems}, nil
	case i_MAP_OBJ:
		var ows []ObjectWrapper
		diag("IN MAP", data)
		err := cbor.Unmarshal(data, &ows)
		if err != nil {
			return nil, err
		}
		// /2 because length is keys+values
		pairs := NewPairsMapWithSize(len(ows) / 2)
		for i := 0; i < len(ows); i += 2 {
			kow := ows[i]
			vow := ows[i+1]
			diag("IN MAP KOW", kow.Data)
			diag("IN MAP VOW", vow.Data)
			kobj, err := decodeFromType(kow.Type, kow.Data)
			if err != nil {
				return nil, err
			}
			vobj, err := decodeFromType(vow.Type, vow.Data)
			if err != nil {
				return nil, err
			}
			hashKey := HashObject(kobj)
			hk := HashKey{
				Type:  kobj.Type(),
				Value: hashKey,
			}
			pairs.Set(hk, MapPair{Key: kobj, Value: vobj})
		}
		return &Map{Pairs: pairs}, nil
	case i_FUNCTION_OBJ:
		var x string
		diag("FUNCTION", data)
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			return nil, err
		}
		return &StringFunction{Value: x}, nil
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

var EmptyOW = ObjectWrapper{}

func marshalObject(obj Object) (ObjectWrapper, error) {
	var data []byte
	var err error
	switch obj.IType() {
	case i_REGEX_OBJ:
		s := obj.(*Regex).Value.String()
		data, err = cbor.Marshal(s)
	case i_BYTES_OBJ:
		bs := obj.(*Bytes).Value
		data, err = cbor.Marshal(bs)
	case i_LIST_OBJ:
		elems := obj.(*List).Elements
		ows := make([]ObjectWrapper, len(elems))
		for i, e := range elems {
			ow, err := marshalObject(e)
			if err != nil {
				return EmptyOW, err
			}
			ows[i] = ow
		}
		data, err = cbor.Marshal(ows)
	case i_SET_OBJ:
		elems := obj.(*Set).Elements
		ows := make([]ObjectWrapper, elems.Len())
		for i, key := range elems.Keys {
			v, _ := elems.Get(key)
			ow, err := marshalObject(v.Value)
			if err != nil {
				return EmptyOW, err
			}
			ows[i] = ow
		}
		data, err = cbor.Marshal(ows)
	case i_MAP_OBJ:
		pairs := obj.(*Map).Pairs
		// *2 to store keys and values
		// When decoding, value comes after key
		ows := make([]ObjectWrapper, 0, pairs.Len()*2)
		for _, key := range pairs.Keys {
			v, _ := pairs.Get(key)
			kow, err := marshalObject(v.Key)
			if err != nil {
				return EmptyOW, err
			}
			vow, err := marshalObject(v.Value)
			if err != nil {
				return EmptyOW, err
			}
			ows = append(ows, kow)
			ows = append(ows, vow)
		}
		data, err = cbor.Marshal(ows)
	case i_FUNCTION_OBJ:
		s := obj.(*Function).Inspect()
		data, err = cbor.Marshal(s)
	default:
		data, err = cbor.Marshal(obj)
	}
	if err != nil {
		return EmptyOW, err
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
	return marshalObjectWrapper(x)
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
	return marshalObjectWrapper(x)
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
	return marshalObjectWrapper(x)
}

func (x *Regex) IType() iType {
	return i_REGEX_OBJ
}

func (x *List) Encode() ([]byte, error) {
	return marshalObjectWrapper(x)
}

func (x *List) IType() iType {
	return i_LIST_OBJ
}

func (x *Map) Encode() ([]byte, error) {
	return marshalObjectWrapper(x)
}

func (x *Map) IType() iType {
	return i_MAP_OBJ
}

func (x *Set) Encode() ([]byte, error) {
	return marshalObjectWrapper(x)
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

func (x *StringFunction) Encode() ([]byte, error) {
	panic(fmt.Sprintf("%T cannot be encoded", x))
}

func (x *StringFunction) IType() iType {
	return i_FUNCTION_OBJ
}
