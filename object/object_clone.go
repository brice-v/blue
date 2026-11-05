package object

import (
	"math/big"

	clone "github.com/huandu/go-clone/generic"
)

func cloneObjectAtDepth(obj Object, depth int) Object {
	if obj == nil {
		return nil
	}
	if depth > maxCloneDepth {
		return obj
	}
	switch x := obj.(type) {
	case *List:
		return cloneListAtDepth(x, depth)
	case *ListCompLiteral:
		return &ListCompLiteral{Elements: cloneSliceAtDepth(x.Elements, depth)}
	case *Map:
		return cloneMapAtDepth(x, depth)
	case *MapCompLiteral:
		return cloneMapCompLiteralAtDepth(x, depth)
	case *Set:
		return cloneSetAtDepth(x, depth)
	case *SetCompLiteral:
		return cloneSetCompLiteralAtDepth(x, depth)
	case *BlueStruct:
		return cloneBlueStructAtDepth(x, depth)
	case *DefaultArgs:
		m := make(map[string]Object, len(x.Value))
		for k, v := range x.Value {
			m[k] = cloneObjectAtDepth(v, depth+1)
		}
		return &DefaultArgs{Value: m}
	case *Closure:
		return &Closure{Fun: x.Fun.Clone().(*CompiledFunction), Free: cloneSliceAtDepth(x.Free, depth)}
	case *ReturnValue:
		return cloneObjectAtDepth(x.Value, depth+1)
	case *BuiltinObj:
		return &BuiltinObj{Obj: cloneObjectAtDepth(x.Obj, depth+1), HelpStr: x.HelpStr}
	}
	return obj.Clone()
}

func cloneSliceAtDepth(elements []Object, depth int) []Object {
	newElements := make([]Object, len(elements))
	for i, e := range elements {
		if e != nil {
			newElements[i] = cloneObjectAtDepth(e, depth+1)
		}
	}
	return newElements
}

func (x *Null) Clone() Object {
	return NULL
}

func (x *Ignore) Clone() Object {
	return VM_IGNORE
}

func (x *BreakStatement) Clone() Object {
	return BREAK
}

func (x *ContinueStatement) Clone() Object {
	return CONTINUE
}

func (x *Integer) Clone() Object {
	// Integers are immutable and already shared via the small-int caches,
	// so cloning can share the receiver.
	return x
}

func (x *BigInteger) Clone() Object {
	return &BigInteger{Value: new(big.Int).Set(x.Value)}
}

func (x *Boolean) Clone() Object {
	return nativeToBooleanObject(x.Value)
}

func (x *DefaultArgs) Clone() Object {
	m := make(map[string]Object)
	for k, v := range x.Value {
		m[k] = v.Clone()
	}
	return &DefaultArgs{Value: m}
}

func (x *UInteger) Clone() Object {
	// Immutable, safe to share.
	return x
}

func (x *Float) Clone() Object {
	// Immutable, safe to share.
	return x
}

func (x BigFloat) Clone() Object {
	return BigFloat{Value: x.Value.Copy()}
}

func (x *ReturnValue) Clone() Object {
	return x.Value.Clone()
}

func (x *Error) Clone() Object {
	return &Error{Message: x.Message}
}

func (x *StringFunction) Clone() Object {
	return &StringFunction{Value: x.Value}
}

// Using go-clone for this as its only used in evaluator
func (x *Function) Clone() Object {
	return clone.Clone(x)
}

func (x *Process) Clone() Object {
	// Cannot create new channel, just copy it
	return &Process{Ch: x.Ch, Id: x.Id, NodeName: x.NodeName}
}

func (x *Stringo) Clone() Object {
	// Immutable value semantics; strings are already shared via the
	// intern tables, so cloning can share the receiver.
	return x
}

func (x *Bytes) Clone() Object {
	bs := make([]byte, len(x.Value))
	copy(bs, x.Value)
	return &Bytes{Value: bs}
}

func (x *GoObj[T]) Clone() Object {
	// Only copy the go-obj value
	return &GoObj[T]{Id: x.Id, Value: x.Value}
}

func (x *GoObjectGob) Clone() Object {
	bs := make([]byte, len(x.Value))
	copy(bs, x.Value)
	return &GoObjectGob{T: x.T, Value: bs}
}

func (x *Regex) Clone() Object {
	// Just copying as this shouldnt be able to change
	return &Regex{Value: x.Value}
}

func (x *Builtin) Clone() Object {
	// Just copying as this shouldnt be able to change
	return x
}

func (x *BuiltinObj) Clone() Object {
	return &BuiltinObj{Obj: x.Obj.Clone(), HelpStr: x.HelpStr}
}

func (x *List) Clone() Object {
	return cloneListAtDepth(x, 0)
}

func cloneListAtDepth(x *List, depth int) Object {
	return &List{Elements: cloneSliceAtDepth(x.Elements, depth)}
}

func (x *ListCompLiteral) Clone() Object {
	return &ListCompLiteral{Elements: cloneSliceAtDepth(x.Elements, 0)}
}

func (x *Map) Clone() Object {
	return cloneMapAtDepth(x, 0)
}

func cloneMapAtDepth(x *Map, depth int) Object {
	pairs := NewOrderedMapWithSize[HashKey, MapPair](x.Pairs.Len())
	for _, k := range x.Pairs.Keys {
		mp, _ := x.Pairs.Get(k)
		pairs.Set(k, MapPair{
			Key:   cloneObjectAtDepth(mp.Key, depth+1),
			Value: cloneObjectAtDepth(mp.Value, depth+1),
		})
	}
	return &Map{Pairs: *pairs, IsEnvBuiltin: x.IsEnvBuiltin}
}

func (x *MapCompLiteral) Clone() Object {
	return cloneMapCompLiteralAtDepth(x, 0)
}

func cloneMapCompLiteralAtDepth(x *MapCompLiteral, depth int) Object {
	pairs := NewOrderedMapWithSize[HashKey, MapPair](x.Pairs.Len())
	for _, k := range x.Pairs.Keys {
		mp, _ := x.Pairs.Get(k)
		pairs.Set(k, MapPair{
			Key:   cloneObjectAtDepth(mp.Key, depth+1),
			Value: cloneObjectAtDepth(mp.Value, depth+1),
		})
	}
	return &Map{Pairs: *pairs}
}

func (x *Set) Clone() Object {
	return cloneSetAtDepth(x, 0)
}

func cloneSetAtDepth(x *Set, depth int) Object {
	pairs := NewOrderedMapWithSize[uint64, SetPair](x.Elements.Len())
	for _, k := range x.Elements.Keys {
		mp, _ := x.Elements.Get(k)
		pairs.Set(k, SetPair{Value: cloneObjectAtDepth(mp.Value, depth+1), Present: mp.Present})
	}
	return &Set{Elements: pairs}
}

func (x *SetCompLiteral) Clone() Object {
	return cloneSetCompLiteralAtDepth(x, 0)
}

func cloneSetCompLiteralAtDepth(x *SetCompLiteral, depth int) Object {
	m := make(map[uint64]SetPair, len(x.Elements))
	for k, v := range x.Elements {
		m[k] = SetPair{Value: cloneObjectAtDepth(v.Value, depth+1), Present: v.Present}
	}
	return &SetCompLiteral{Elements: m}
}

func (x *Module) Clone() Object {
	m := &Module{Name: x.Name, HelpStr: x.HelpStr}
	if x.Env == nil {
		return m
	} else {
		// Using go-clone for this as its only used in evaluator
		m.Env = clone.Clone(m.Env)
		return m
	}
}

func (x *BlueStruct) Clone() Object {
	return cloneBlueStructAtDepth(x, 0)
}

func cloneBlueStructAtDepth(x *BlueStruct, depth int) Object {
	fields := make([]string, len(x.Fields))
	copy(fields, x.Fields)
	return &BlueStruct{Fields: fields, Values: cloneSliceAtDepth(x.Values, depth)}
}

func (x *CompiledFunction) Clone() Object {
	x.locker.Lock()
	instructions := make([]byte, len(x.Instructions))
	copy(instructions, x.Instructions)
	parameters := make([]string, len(x.Parameters))
	copy(parameters, x.Parameters)
	parametersHasDefault := make([]bool, len(x.ParameterHasDefault))
	copy(parametersHasDefault, x.ParameterHasDefault)
	specialFunctionParameters := make(map[NameIndexKey]map[NameIndexKey]Object)
	for k, v := range x.SpecialFunctionParameters {
		m := make(map[NameIndexKey]Object)
		for kk, vv := range v {
			if vv == nil {
				m[kk] = nil
			} else {
				m[kk] = vv.Clone()
			}
		}
		specialFunctionParameters[k] = m
	}
	x.locker.Unlock()
	return &CompiledFunction{
		Instructions:              instructions,
		NumLocals:                 x.NumLocals,
		NumParameters:             x.NumParameters,
		Parameters:                parameters,
		ParameterHasDefault:       parametersHasDefault,
		NumDefaultParams:          x.NumDefaultParams,
		DisplayString:             x.DisplayString,
		SpecialFunctionParameters: specialFunctionParameters,
	}
}

func (x *Closure) Clone() Object {
	return &Closure{Fun: x.Fun.Clone().(*CompiledFunction), Free: cloneSliceAtDepth(x.Free, 0)}
}

func (x *ExecString) Clone() Object {
	return &ExecString{Value: x.Value}
}

func CloneSlice(elements []Object) []Object {
	return cloneSliceAtDepth(elements, 0)
}

func immutableGlobalType(o Object) bool {
	switch o.IType() {
	case i_INTEGER_OBJ, i_BIG_INTEGER_OBJ, i_BOOLEAN_OBJ, i_NULL_OBJ,
		i_UINTEGER_OBJ, i_FLOAT_OBJ, i_BIG_FLOAT_OBJ, i_STRING_OBJ,
		i_FUNCTION_OBJ, i_COMPILED_FUNCTION_OBJ, i_CLOSURE_OBJ,
		i_BUILTIN_OBJ:
		return true
	}
	return false
}

func CloneGlobals(elements []Object, highWater int) []Object {
	if highWater > len(elements) {
		highWater = len(elements)
	}
	newElements := make([]Object, highWater)
	copy(newElements, elements[:highWater])
	for i := range newElements {
		if e := newElements[i]; e != nil && !immutableGlobalType(e) {
			newElements[i] = e.Clone()
		}
	}
	return newElements
}
