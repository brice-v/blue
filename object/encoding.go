package object

import (
	"fmt"
	"regexp"
	"sort"

	"blue/code"

	"github.com/fxamacker/cbor/v2"
)

// iType is the object type represented as an integer
type iType byte

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
	i_BLUE_STRUCT_OBJ
	i_LIST_COMP_OBJ
	i_MAP_COMP_OBJ
	i_SET_COMP_OBJ
	i_MODULE_OBJ
	i_PROCESS_OBJ

	i_BREAK_OBJ
	i_CONTINUE_OBJ

	i_COMPILED_FUNCTION_OBJ
	i_CLOSURE_OBJ
	i_EXEC_STRING_OBJ
	i_IGNORE_OBJ
	i_DEFAULT_ARGS_OBJ

	// i_STRUCT_FIELDS_OBJ is the faithful serializable form of the
	// GoObj[[]string] constants the compiler emits for struct literals
	// (field name lists). Generic GoObj values otherwise decode lossily
	// as GoObjectGob, which would break struct matching after an image
	// round-trip. Appended at the end so existing encodings stay valid.
	i_STRUCT_FIELDS_OBJ
)

type ObjectWrapper struct {
	Type iType           `cbor:"type"`
	Data cbor.RawMessage `cbor:"data"`
}

// maxSerializeDepth bounds recursion when marshaling/unmarshaling nested
// objects (lists of maps of lists ...). Constant pools are expected to be
// acyclic, this turns a cycle (or absurdly deep nesting) into an error
// instead of a stack overflow.
const maxSerializeDepth = 512

// errTooDeep is returned when the serialization depth limit is exceeded.
var errTooDeep = fmt.Errorf("object nesting too deep (possible cycle), aborting serialization at depth %d", maxSerializeDepth)

// encNameIndexEntry is the serializable form of one entry of a
// CompiledFunction's SpecialFunctionParameters inner map.
type encNameIndexEntry struct {
	Name  string        `cbor:"n"`
	Index int           `cbor:"i"`
	Value ObjectWrapper `cbor:"v"`
}

// encNameIndexKey is the serializable mirror of NameIndexKey.
type encNameIndexKey struct {
	Name  string `cbor:"n"`
	Index int    `cbor:"i"`
}

// encSFPGroup is one outer entry of SpecialFunctionParameters:
// a key plus its inner name/index -> object entries.
type encSFPGroup struct {
	Key     encNameIndexKey     `cbor:"k"`
	Entries []encNameIndexEntry `cbor:"e"`
}

// encCompiledFunction is the serializable mirror of CompiledFunction. It
// exists because CompiledFunction contains a sync.Mutex which must not be
// serialized, and its SpecialFunctionParameters map needs restructuring.
type encCompiledFunction struct {
	Instructions     []byte        `cbor:"ins"`
	NumLocals        int           `cbor:"nl"`
	NumParameters    int           `cbor:"np"`
	Parameters       []string      `cbor:"ps"`
	ParamHasDefault  []bool        `cbor:"phd"`
	NumDefaultParams int           `cbor:"ndp"`
	DisplayString    string        `cbor:"ds"`
	HelpStr          string        `cbor:"hs"`
	SFP              []encSFPGroup `cbor:"sfp"`
}

// encKV is a serializable string-keyed entry, used by DefaultArgs.
type encKV struct {
	Key   string        `cbor:"k"`
	Value ObjectWrapper `cbor:"v"`
}

// encDefaultArgs is the serializable mirror of DefaultArgs.
type encDefaultArgs struct {
	Entries []encKV `cbor:"e"`
}

// encModule is the serializable mirror of Module (Name + HelpStr only, the
// environment is nil at compile time).
type encModule struct {
	Name    string `cbor:"name"`
	HelpStr string `cbor:"hs"`
}

// encBlueStruct is the serializable mirror of BlueStruct.
type encBlueStruct struct {
	Fields []string        `cbor:"f"`
	Values []ObjectWrapper `cbor:"v"`
}

func decodeFromType(t iType, data []byte, depth int) (Object, error) {
	if depth > maxSerializeDepth {
		return nil, errTooDeep
	}
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
			obj, err := decodeFromType(e.Type, e.Data, depth+1)
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
			obj, err := decodeFromType(e.Type, e.Data, depth+1)
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
			kobj, err := decodeFromType(kow.Type, kow.Data, depth+1)
			if err != nil {
				return nil, err
			}
			vobj, err := decodeFromType(vow.Type, vow.Data, depth+1)
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
	case i_CLOSURE_OBJ:
		return nil, fmt.Errorf("TODO: Closure unsupported for unmarhaling currently")
	case i_GO_OBJ:
		var ows []ObjectWrapper
		diag("GOOBJ", data)
		err := cbor.Unmarshal(data, &ows)
		if err != nil {
			return nil, err
		}
		elems := make([]Object, len(ows))
		for i, e := range ows {
			diag("IN LOOP", e.Data)
			obj, err := decodeFromType(e.Type, e.Data, depth+1)
			if err != nil {
				return nil, err
			}
			elems[i] = obj
		}
		return &GoObjectGob{
			T:     elems[0].(*Stringo).Value,
			Value: elems[1].(*Bytes).Value,
		}, nil
	case i_STRUCT_FIELDS_OBJ:
		var fields []string
		if cerr := cbor.Unmarshal(data, &fields); cerr != nil {
			return nil, cerr
		}
		return NewGoObj(fields), nil
	case i_COMPILED_FUNCTION_OBJ:
		var x encCompiledFunction
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			return nil, err
		}
		sfp := make(map[NameIndexKey]map[NameIndexKey]Object, len(x.SFP))
		for _, group := range x.SFP {
			inner := make(map[NameIndexKey]Object, len(group.Entries))
			for _, e := range group.Entries {
				vobj, derr := decodeFromType(e.Value.Type, e.Value.Data, depth+1)
				if derr != nil {
					return nil, derr
				}
				inner[NameIndexKey{Name: e.Name, Index: e.Index}] = vobj
			}
			sfp[NameIndexKey{Name: group.Key.Name, Index: group.Key.Index}] = inner
		}
		return &CompiledFunction{
			Instructions:             code.Instructions(x.Instructions),
			NumLocals:                x.NumLocals,
			NumParameters:            x.NumParameters,
			Parameters:               x.Parameters,
			ParameterHasDefault:      x.ParamHasDefault,
			NumDefaultParams:         x.NumDefaultParams,
			DisplayString:            x.DisplayString,
			HelpStr:                  x.HelpStr,
			SpecialFunctionParameters: sfp,
		}, nil
	case i_MODULE_OBJ:
		var x encModule
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			return nil, err
		}
		// Env is always nil for serialized modules: at compile time a
		// module constant carries no live environment.
		return &Module{Name: x.Name, HelpStr: x.HelpStr}, nil
	case i_BLUE_STRUCT_OBJ:
		var x encBlueStruct
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			return nil, err
		}
		values := make([]Object, len(x.Values))
		for i, v := range x.Values {
			vobj, derr := decodeFromType(v.Type, v.Data, depth+1)
			if derr != nil {
				return nil, derr
			}
			values[i] = vobj
		}
		return &BlueStruct{Fields: x.Fields, Values: values}, nil
	case i_EXEC_STRING_OBJ:
		var x string
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			return nil, err
		}
		return &ExecString{Value: x}, nil
	case i_IGNORE_OBJ:
		return VM_IGNORE, nil
	case i_DEFAULT_ARGS_OBJ:
		var x encDefaultArgs
		err := cbor.Unmarshal(data, &x)
		if err != nil {
			return nil, err
		}
		value := make(map[string]Object, len(x.Entries))
		for _, e := range x.Entries {
			vobj, derr := decodeFromType(e.Value.Type, e.Value.Data, depth+1)
			if derr != nil {
				return nil, derr
			}
			value[e.Key] = vobj
		}
		return &DefaultArgs{Value: value}, nil
	case i_BREAK_OBJ:
		return BREAK, nil
	case i_CONTINUE_OBJ:
		return CONTINUE, nil
	default:
		return nil, fmt.Errorf("decodeFromType: handle %d", t)
	}
}

func diag(prefix string, data []byte) {
	// Note: Uncomment the lines below for some debugging info
	// if s, err := cbor.Diagnose(data); err == nil {
	// 	log.Printf("%s Diagnose %s", prefix, s)
	// } else {
	// 	log.Printf("%s Dianose error %s", prefix, err.Error())
	// }
}

func Decode(data []byte) (Object, error) {
	var a ObjectWrapper
	diag("DECODE", data)
	err := cbor.Unmarshal(data, &a)
	if err != nil {
		return nil, err
	}
	return decodeFromType(a.Type, a.Data, 0)
}

var EmptyOW = ObjectWrapper{}

func marshalObject(obj Object) (ObjectWrapper, error) {
	return marshalObjectDepth(obj, 0)
}

func marshalObjectDepth(obj Object, depth int) (ObjectWrapper, error) {
	if obj == nil {
		return EmptyOW, fmt.Errorf("cannot encode nil object")
	}
	if depth > maxSerializeDepth {
		return EmptyOW, errTooDeep
	}
	// The compiler emits GoObj[[]string] constants for struct literals;
	// they serialize under their own type tag for a faithful round trip.
	if fields, isFields := obj.(*GoObj[[]string]); isFields {
		fieldData, ferr := cbor.Marshal(fields.Value)
		if ferr != nil {
			return EmptyOW, ferr
		}
		return ObjectWrapper{Type: i_STRUCT_FIELDS_OBJ, Data: fieldData}, nil
	}
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
			ow, err := marshalObjectDepth(e, depth+1)
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
			ow, err := marshalObjectDepth(v.Value, depth+1)
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
			kow, err := marshalObjectDepth(v.Key, depth+1)
			if err != nil {
				return EmptyOW, err
			}
			vow, err := marshalObjectDepth(v.Value, depth+1)
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
	case i_CLOSURE_OBJ:
		err = fmt.Errorf("TODO: Closure unsupported for marshaling currently")
	case i_COMPILED_FUNCTION_OBJ:
		cf := obj.(*CompiledFunction)
		sfp := make([]encSFPGroup, 0, len(cf.SpecialFunctionParameters))
		for ok, inner := range cf.SpecialFunctionParameters {
			group := encSFPGroup{
				Key:     encNameIndexKey(ok),
				Entries: make([]encNameIndexEntry, 0, len(inner)),
			}
			for ik, iv := range inner {
				vow, merr := marshalObjectDepth(iv, depth+1)
				if merr != nil {
					return EmptyOW, merr
				}
				group.Entries = append(group.Entries, encNameIndexEntry{Name: ik.Name, Index: ik.Index, Value: vow})
			}
			sort.Slice(group.Entries, func(a, b int) bool {
				if group.Entries[a].Name != group.Entries[b].Name {
					return group.Entries[a].Name < group.Entries[b].Name
				}
				return group.Entries[a].Index < group.Entries[b].Index
			})
			sfp = append(sfp, group)
		}
		sort.Slice(sfp, func(a, b int) bool {
			if sfp[a].Key.Name != sfp[b].Key.Name {
				return sfp[a].Key.Name < sfp[b].Key.Name
			}
			return sfp[a].Key.Index < sfp[b].Key.Index
		})
		data, err = cbor.Marshal(encCompiledFunction{
			Instructions:     []byte(cf.Instructions),
			NumLocals:        cf.NumLocals,
			NumParameters:    cf.NumParameters,
			Parameters:       cf.Parameters,
			ParamHasDefault:  cf.ParameterHasDefault,
			NumDefaultParams: cf.NumDefaultParams,
			DisplayString:    cf.DisplayString,
			HelpStr:          cf.HelpStr,
			SFP:              sfp,
		})
	case i_MODULE_OBJ:
		m := obj.(*Module)
		data, err = cbor.Marshal(encModule{Name: m.Name, HelpStr: m.HelpStr})
	case i_BLUE_STRUCT_OBJ:
		bs := obj.(*BlueStruct)
		values := make([]ObjectWrapper, len(bs.Values))
		for i, v := range bs.Values {
			vow, merr := marshalObjectDepth(v, depth+1)
			if merr != nil {
				return EmptyOW, merr
			}
			values[i] = vow
		}
		data, err = cbor.Marshal(encBlueStruct{Fields: bs.Fields, Values: values})
	case i_EXEC_STRING_OBJ:
		data, err = cbor.Marshal(obj.(*ExecString).Value)
	case i_IGNORE_OBJ:
		data, err = cbor.Marshal(nil)
	case i_DEFAULT_ARGS_OBJ:
		da := obj.(*DefaultArgs)
		entries := make([]encKV, 0, len(da.Value))
		for k, v := range da.Value {
			vow, merr := marshalObjectDepth(v, depth+1)
			if merr != nil {
				return EmptyOW, merr
			}
			entries = append(entries, encKV{Key: k, Value: vow})
		}
		// Sorted so encoding is deterministic (map iteration is not)
		sort.Slice(entries, func(a, b int) bool { return entries[a].Key < entries[b].Key })
		data, err = cbor.Marshal(encDefaultArgs{Entries: entries})
	case i_BREAK_OBJ:
		data, err = cbor.Marshal(nil)
	case i_CONTINUE_OBJ:
		data, err = cbor.Marshal(nil)
	case i_GO_OBJ:
		// GoObj[[]string] struct field lists are handled before the
		// switch (i_STRUCT_FIELDS_OBJ). Live Go objects cannot be stored
		// in a binary image, so anything else here is rejected.
		return EmptyOW, fmt.Errorf("%T holds live Go state and cannot be serialized", obj)
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

func (x *BlueStruct) Encode() ([]byte, error) {
	return marshalObjectWrapper(x)
}

func (x *BlueStruct) IType() iType {
	return i_BLUE_STRUCT_OBJ
}

// Encode stays unavailable for the runtime `save` builtin: a live module
// holds an Environment that cannot be serialized. Compile-time module
// CONSTANTS (Name + HelpStr only, Env always nil) ARE serialized by the
// binary container via marshalObjectDepth.
func (x *Module) Encode() ([]byte, error) {
	return nil, fmt.Errorf("%s is not supported for encoding", x.Type())
}

func (x *Module) IType() iType {
	return i_MODULE_OBJ
}

func (x *GoObj[T]) Encode() ([]byte, error) {
	// return marshalObjectWrapper(x)
	return nil, fmt.Errorf("%T is not supported for encoding", x)
}

func (x *GoObj[T]) IType() iType {
	return i_GO_OBJ
}

func (x *CompiledFunction) Encode() ([]byte, error) {
	return marshalObjectWrapper(x)
}

func (x *CompiledFunction) IType() iType {
	return i_COMPILED_FUNCTION_OBJ
}

func (x *Closure) Encode() ([]byte, error) {
	return marshalObjectWrapper(x)
}

func (x *Closure) IType() iType {
	return i_CLOSURE_OBJ
}

func (x *ExecString) Encode() ([]byte, error) {
	return marshalObjectWrapper(x)
}

func (x *ExecString) IType() iType {
	return i_EXEC_STRING_OBJ
}

func (x *Ignore) Encode() ([]byte, error) {
	return marshalObjectWrapper(x)
}

func (x *Ignore) IType() iType {
	return i_IGNORE_OBJ
}

func (x *DefaultArgs) Encode() ([]byte, error) {
	return marshalObjectWrapper(x)
}

func (x *DefaultArgs) IType() iType {
	return i_DEFAULT_ARGS_OBJ
}

// The Objects Below cannot be encoded but are included to satisfy the Object interface

func (x *Error) Encode() ([]byte, error) {
	panic(fmt.Sprintf("%T cannot be encoded", x))
}

func (x *Error) IType() iType {
	return i_ERROR_OBJ
}

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

func (x *ReturnValue) Encode() ([]byte, error) {
	panic(fmt.Sprintf("%T cannot be encoded", x))
}

func (x *ReturnValue) IType() iType {
	return i_RETURN_VALUE_OBJ
}

func (x *GoObjectGob) Encode() ([]byte, error) {
	panic(fmt.Sprintf("%T cannot be encoded", x))
}

func (x *GoObjectGob) IType() iType {
	return i_GO_OBJ
}

func (x *BreakStatement) Encode() ([]byte, error) {
	panic(fmt.Sprintf("%T cannot be encoded", x))
}

func (x *BreakStatement) IType() iType {
	return i_BREAK_OBJ
}

func (x *ContinueStatement) Encode() ([]byte, error) {
	panic(fmt.Sprintf("%T cannot be encoded", x))
}

func (x *ContinueStatement) IType() iType {
	return i_CONTINUE_OBJ
}
