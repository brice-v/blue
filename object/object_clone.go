package object

import (
	"math/big"

	clone "github.com/huandu/go-clone/generic"
	"github.com/puzpuzpuz/xsync/v3"
)

func (x *Null) Clone() Object {
	return NULL
}

func (x *Ignore) Clone() Object {
	return IGNORE
}

func (x *BreakStatement) Clone() Object {
	return BREAK
}

func (x *ContinueStatement) Clone() Object {
	return CONTINUE
}

func (x *Integer) Clone() Object {
	return &Integer{Value: x.Value}
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
	return &UInteger{Value: x.Value}
}

func (x *Float) Clone() Object {
	return &Float{Value: x.Value}
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
	return &Stringo{Value: x.Value}
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
	return &BuiltinObj{Obj: x.Clone(), HelpStr: x.HelpStr}
}

func (x *List) Clone() Object {
	return &List{Elements: CloneSlice(x.Elements)}
}

func (x *ListCompLiteral) Clone() Object {
	return &ListCompLiteral{Elements: CloneSlice(x.Elements)}
}

func (x *Map) Clone() Object {
	pairs := NewOrderedMapWithSize[HashKey, MapPair](x.Pairs.Len())
	for _, k := range x.Pairs.Keys {
		mp, _ := x.Pairs.Get(k)
		pairs.Set(k, MapPair{Key: mp.Key.Clone(), Value: mp.Value.Clone()})
	}
	return &Map{Pairs: *pairs, IsEnvBuiltin: x.IsEnvBuiltin}
}

func (x *MapCompLiteral) Clone() Object {
	pairs := NewOrderedMapWithSize[HashKey, MapPair](x.Pairs.Len())
	for _, k := range x.Pairs.Keys {
		mp, _ := x.Pairs.Get(k)
		pairs.Set(k, MapPair{Key: mp.Key.Clone(), Value: mp.Value.Clone()})
	}
	return &Map{Pairs: *pairs}
}

func (x *Set) Clone() Object {
	pairs := NewOrderedMapWithSize[uint64, SetPair](x.Elements.Len())
	for _, k := range x.Elements.Keys {
		mp, _ := x.Elements.Get(k)
		pairs.Set(k, SetPair{Value: mp.Value.Clone(), Present: mp.Present})
	}
	return &Set{Elements: pairs}
}

func (x *SetCompLiteral) Clone() Object {
	m := make(map[uint64]SetPair)
	for k, v := range x.Elements {
		m[k] = SetPair{Value: v.Value.Clone(), Present: v.Present}
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
	fields := make([]string, len(x.Fields))
	copy(fields, x.Fields)
	return &BlueStruct{Fields: fields, Values: CloneSlice(x.Values)}
}

func (x *CompiledFunction) Clone() Object {
	x.locker.Lock()
	instructions := make([]byte, len(x.Instructions))
	copy(instructions, x.Instructions)
	parameters := make([]string, len(x.Parameters))
	copy(parameters, x.Parameters)
	parametersHasDefault := make([]bool, len(x.ParameterHasDefault))
	copy(parametersHasDefault, x.ParameterHasDefault)
	posAlreadyIncremented := xsync.NewMapOf[int, struct{}]()
	x.PosAlreadyIncremented.Range(func(key int, value struct{}) bool {
		posAlreadyIncremented.Store(key, value)
		return true
	})
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
		PosAlreadyIncremented:     posAlreadyIncremented,
		SpecialFunctionParameters: specialFunctionParameters,
	}
}

func (x *Closure) Clone() Object {
	return &Closure{Fun: x.Fun.Clone().(*CompiledFunction), Free: CloneSlice(x.Free)}
}

func (x *ExecString) Clone() Object {
	return &ExecString{Value: x.Value}
}

func CloneSlice(elements []Object) []Object {
	newElements := make([]Object, len(elements))
	for i, e := range elements {
		if e != nil {
			newElements[i] = e.Clone()
		}
	}
	return newElements
}
