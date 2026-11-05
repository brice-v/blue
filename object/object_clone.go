package object

import (
	"maps"
	"math/big"

	clone "github.com/huandu/go-clone/generic"
)

type Cloner struct {
	seen map[Object]Object
}

func newCloner() *Cloner {
	return &Cloner{seen: make(map[Object]Object)}
}

func (c *Cloner) clone(obj Object) Object {
	if obj == nil {
		return nil
	}

	switch x := obj.(type) {
	case *List:
		if v, ok := c.seen[x]; ok {
			return v
		}
		v := &List{Elements: make([]Object, len(x.Elements))}
		c.seen[x] = v
		for i, e := range x.Elements {
			v.Elements[i] = c.clone(e)
		}
		return v

	case *ListCompLiteral:
		if v, ok := c.seen[x]; ok {
			return v
		}
		v := &ListCompLiteral{Elements: make([]Object, len(x.Elements))}
		c.seen[x] = v
		for i, e := range x.Elements {
			v.Elements[i] = c.clone(e)
		}
		return v

	case *Map:
		if v, ok := c.seen[x]; ok {
			return v
		}
		pairs := NewOrderedMapWithSize[HashKey, MapPair](x.Pairs.Len())
		v := &Map{Pairs: *pairs, IsEnvBuiltin: x.IsEnvBuiltin}
		c.seen[x] = v
		for _, k := range x.Pairs.Keys {
			mp, _ := x.Pairs.Get(k)
			v.Pairs.Set(k, MapPair{
				Key:   c.clone(mp.Key),
				Value: c.clone(mp.Value),
			})
		}
		return v

	case *MapCompLiteral:
		if v, ok := c.seen[x]; ok {
			return v
		}
		pairs := NewOrderedMapWithSize[HashKey, MapPair](x.Pairs.Len())
		v := &MapCompLiteral{Pairs: *pairs}
		c.seen[x] = v
		for _, k := range x.Pairs.Keys {
			mp, _ := x.Pairs.Get(k)
			v.Pairs.Set(k, MapPair{
				Key:   c.clone(mp.Key),
				Value: c.clone(mp.Value),
			})
		}
		return v

	case *Set:
		if v, ok := c.seen[x]; ok {
			return v
		}
		pairs := NewOrderedMapWithSize[uint64, SetPair](x.Elements.Len())
		v := &Set{Elements: pairs}
		c.seen[x] = v
		for _, k := range x.Elements.Keys {
			mp, _ := x.Elements.Get(k)
			v.Elements.Set(k, SetPair{Value: c.clone(mp.Value), Present: mp.Present})
		}
		return v

	case *SetCompLiteral:
		if v, ok := c.seen[x]; ok {
			return v
		}
		m := make(map[uint64]SetPair, len(x.Elements))
		v := &SetCompLiteral{Elements: m}
		c.seen[x] = v
		for k, e := range x.Elements {
			v.Elements[k] = SetPair{Value: c.clone(e.Value), Present: e.Present}
		}
		return v

	case *BlueStruct:
		if v, ok := c.seen[x]; ok {
			return v
		}
		values := make([]Object, len(x.Values))
		v := &BlueStruct{Fields: append([]string(nil), x.Fields...), Values: values}
		c.seen[x] = v
		for i, e := range x.Values {
			v.Values[i] = c.clone(e)
		}
		return v

	case *DefaultArgs:
		if v, ok := c.seen[x]; ok {
			return v
		}
		m := make(map[string]Object, len(x.Value))
		v := &DefaultArgs{Value: m}
		c.seen[x] = v
		for k, e := range x.Value {
			v.Value[k] = c.clone(e)
		}
		return v

	case *Closure:
		if v, ok := c.seen[x]; ok {
			return v
		}
		v := &Closure{Fun: nil, Free: make([]Object, len(x.Free))}
		c.seen[x] = v
		if x.Fun != nil {
			v.Fun = c.clone(x.Fun).(*CompiledFunction)
		}
		for i, e := range x.Free {
			v.Free[i] = c.clone(e)
		}
		return v

	case *CompiledFunction:
		if v, ok := c.seen[x]; ok {
			return v
		}
		x.locker.Lock()
		instructions := append([]byte(nil), x.Instructions...)
		parameters := append([]string(nil), x.Parameters...)
		parametersHasDefault := append([]bool(nil), x.ParameterHasDefault...)
		special := make(map[NameIndexKey]map[NameIndexKey]Object, len(x.SpecialFunctionParameters))
		for k, inner := range x.SpecialFunctionParameters {
			copied := make(map[NameIndexKey]Object, len(inner))
			maps.Copy(copied, inner)
			special[k] = copied
		}
		v := &CompiledFunction{
			Instructions:              instructions,
			NumLocals:                 x.NumLocals,
			NumParameters:             x.NumParameters,
			Parameters:                parameters,
			ParameterHasDefault:       parametersHasDefault,
			NumDefaultParams:          x.NumDefaultParams,
			DisplayString:             x.DisplayString,
			SpecialFunctionParameters: special,
			HelpStr:                   x.HelpStr,
		}
		c.seen[x] = v
		x.locker.Unlock()
		for k, inner := range v.SpecialFunctionParameters {
			for kk, vv := range inner {
				v.SpecialFunctionParameters[k][kk] = c.clone(vv)
			}
		}
		return v

	case *BuiltinObj:
		if v, ok := c.seen[x]; ok {
			return v
		}
		v := &BuiltinObj{Name: x.Name, Module: x.Module, HelpStr: x.HelpStr}
		c.seen[x] = v
		v.Obj = c.clone(x.Obj)
		return v

	case *ReturnValue:
		return c.clone(x.Value)

	case *Module:
		if v, ok := c.seen[x]; ok {
			return v
		}
		v := x.Clone().(*Module)
		c.seen[x] = v
		return v
	}

	return obj.Clone()
}

func (c *Cloner) cloneSlice(elements []Object) []Object {
	newElements := make([]Object, len(elements))
	for i, e := range elements {
		if e != nil {
			newElements[i] = c.clone(e)
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
	return newCloner().clone(x)
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
	return newCloner().clone(x)
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
	return newCloner().clone(x)
}

func (x *List) Clone() Object {
	return newCloner().clone(x)
}

func (x *ListCompLiteral) Clone() Object {
	return newCloner().clone(x)
}

func (x *Map) Clone() Object {
	return newCloner().clone(x)
}

func (x *MapCompLiteral) Clone() Object {
	return newCloner().clone(x)
}

func (x *Set) Clone() Object {
	return newCloner().clone(x)
}

func (x *SetCompLiteral) Clone() Object {
	return newCloner().clone(x)
}

func (x *Module) Clone() Object {
	m := &Module{Name: x.Name, HelpStr: x.HelpStr}
	if x.Env == nil {
		return m
	}
	m.Env = clone.Clone(x.Env)
	return m
}

func (x *BlueStruct) Clone() Object {
	return newCloner().clone(x)
}

func (x *CompiledFunction) Clone() Object {
	return newCloner().clone(x)
}

func (x *Closure) Clone() Object {
	return newCloner().clone(x)
}

func (x *ExecString) Clone() Object {
	return &ExecString{Value: x.Value}
}

func CloneSlice(elements []Object) []Object {
	return newCloner().cloneSlice(elements)
}

func immutableGlobalType(o Object) bool {
	switch o.IType() {
	case i_INTEGER_OBJ, i_BIG_INTEGER_OBJ, i_BOOLEAN_OBJ, i_NULL_OBJ,
		i_UINTEGER_OBJ, i_FLOAT_OBJ, i_BIG_FLOAT_OBJ, i_STRING_OBJ,
		i_COMPILED_FUNCTION_OBJ, i_BUILTIN_OBJ:
		return true
	}
	return false
}

func CloneGlobals(elements []Object, highWater int) []Object {
	if highWater < 0 {
		highWater = 0
	}
	if highWater > len(elements) {
		highWater = len(elements)
	}
	for i := len(elements) - 1; i >= highWater; i-- {
		if elements[i] != nil {
			highWater = i + 1
			break
		}
	}

	c := newCloner()
	newElements := make([]Object, highWater)
	for i := 0; i < highWater; i++ {
		e := elements[i]
		if e == nil {
			continue
		}
		if immutableGlobalType(e) {
			newElements[i] = e
			continue
		}
		newElements[i] = c.clone(e)
	}
	return newElements
}
