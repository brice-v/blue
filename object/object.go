package object

import (
	"blue/ast"
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math/big"
	"strings"

	"github.com/shopspring/decimal"
)

const (
	// INTEGER_OBJ is the integer object type string
	INTEGER_OBJ = "INTEGER"
	// BIG_INTEGER_OBJ is the big integer object type string
	BIG_INTEGER_OBJ = "BIG_INTEGER"
	// BOOLEAN_OBJ is the boolean object type string
	BOOLEAN_OBJ = "BOOLEAN"
	// NULL_OBJ is the null object type string
	NULL_OBJ = "NULL"
	// UINTEGER_OBJ is the uint object type string
	UINTEGER_OBJ = "UINTEGER"
	// FLOAT_OBJ is the float object type string
	FLOAT_OBJ = "FLOAT"
	// BIG_FLOAT_OBJ is the big float object type string
	BIG_FLOAT_OBJ = "BIG_FLOAT"
	// RETURN_VALUE_OBJ is the return object type string
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	// ERROR_OBJ is the error object type string
	ERROR_OBJ = "ERROR"
	// FUNCTION_OBJ is the function object type string
	FUNCTION_OBJ = "FUNCTION"
	// STRING_OBJ is the string object type string
	STRING_OBJ = "STRING"
	// BYTES_OBJ is the string object type string
	BYTES_OBJ = "BYTES"
	// BUILTIN_OBJ is the builtin function object type string
	BUILTIN_OBJ = "BUILTIN"
	// LIST_OBJ is the list object type string
	LIST_OBJ = "LIST"
	// MAP_OBJ is the map object type string
	MAP_OBJ = "MAP"
	// SET_OBJ is the set object type
	SET_OBJ = "SET"
	// LIST_COMP_OBJ is the list comprehension literal type string
	LIST_COMP_OBJ = "LIST_COMP_OBJ"
	// MAP_COMP_OBJ is the map comprehension literal type string
	MAP_COMP_OBJ = "MAP_COMP_OBJ"
	// SET_COMP_OBJ is the set comprehension literal type string
	SET_COMP_OBJ = "SET_COMP_OBJ"
	// MODULE_OBJ is the object type for an imported module
	MODULE_OBJ = "MODULE_OBJ"
	// PROCESS_OBJ is the process type for a process
	PROCESS_OBJ = "PROCESS"

	// BREAK_OBJ is the break statement type
	BREAK_OBJ = "BREAK_OBJ"
	// CONTINUE_OBJ is the continue statement type
	CONTINUE_OBJ = "CONTINUE_OBJ"
)

// Type is the object type represented as a string
type Type string

// Object is the interface a value in the language must
// satisfy to be used
type Object interface {
	Type() Type      // Type is a function that returns the objects type
	Inspect() string // Inspect is used for debugging an object
}

// Integer is the integer object type
type Integer struct {
	Value int64 // Value is the internal rep. of an integer, it is stored as an int64
}

// Inspect returns the string value of the integer object
func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }

// Type returns the object type of integer
func (i *Integer) Type() Type { return INTEGER_OBJ }

// BigInteger is the big integer type
type BigInteger struct {
	Value *big.Int
}

// Inspect returns the string value of big integer
func (bi *BigInteger) Inspect() string { return bi.Value.String() }

// Type returns the object type of big integer
func (bi *BigInteger) Type() Type { return BIG_INTEGER_OBJ }

// Boolean is the boolean object type
type Boolean struct {
	Value bool // Value is the internal rep. of a boolean, it is stored as a bool
}

// Inspect returns the string value of the boolean object
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }

// Type returns the object type of boolean
func (b *Boolean) Type() Type { return BOOLEAN_OBJ }

// Null is the null object struct
type Null struct{}

// Type is the object type of null
func (n *Null) Type() Type { return NULL_OBJ }

// Inspect returns the string value of null
func (n *Null) Inspect() string { return "null" }

// UInteger is the hex, octal, bin object struct
// TODO: Separate these all out to their own structs and objects
type UInteger struct {
	Value uint64
}

// Type returns the UINTEGER_OBJ type
func (ui *UInteger) Type() Type { return UINTEGER_OBJ }

// Inspect returns the string value of the uint
func (ui *UInteger) Inspect() string { return fmt.Sprintf("%d", ui.Value) }

// Float is the float object struct
type Float struct {
	Value float64
}

// Type returns the FLOAT_OBJ type
func (f *Float) Type() Type { return FLOAT_OBJ }

// Inspect returns the string value of the float
func (f *Float) Inspect() string { return fmt.Sprintf("%f", f.Value) }

// BigFloat is the big float object struct
type BigFloat struct {
	Value decimal.Decimal
}

// Inspect returns the big float object as a string
func (bf BigFloat) Inspect() string { return bf.Value.String() }

// Type returns the big float object type
func (bf BigFloat) Type() Type { return BIG_FLOAT_OBJ }

// ReturnValue is the struct type for the return value object
type ReturnValue struct {
	Value Object
}

// Type returns the return value object type
func (rv *ReturnValue) Type() Type { return RETURN_VALUE_OBJ }

// Inspect returns the string version of the object to return
func (rv *ReturnValue) Inspect() string { return rv.Value.Inspect() }

// Error is the error object struct.  It conatins a message as a string
type Error struct {
	Message string
}

// Type returns the error object type
func (e *Error) Type() Type { return ERROR_OBJ }

// Inspect returns a string representation of the error
func (e *Error) Inspect() string { return "EvaluatorError: " + e.Message }

// Function is the function object struct
type Function struct {
	Parameters []*ast.Identifier   // Parameters is a slice of identifiers
	Body       *ast.BlockStatement // Body is a block statement node
	Env        *Environment        // Env stores the function's environment

	DefaultParameters []Object // DefaultParameters holds the expression of the default parameter, if it exists otherwise nil
}

// Type returns the function objects type
func (f *Function) Type() Type { return FUNCTION_OBJ }

// Inspect returns the string representation of the function
func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fun(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}

type Process struct {
	Fun *Function
	Ch  chan Object
}

func (p *Process) Inspect() string {
	return "TODO: Process.Inspect()"
}

func (p *Process) Type() Type {
	return PROCESS_OBJ
}

// Stringo is the string oject struct which contains a string value
// it is named stringo to avoid name clashes
type Stringo struct {
	Value string
}

// Type returns the string object type
func (s *Stringo) Type() Type { return STRING_OBJ }

// Inspect returns the string value
func (s *Stringo) Inspect() string { return s.Value }

// Bytes is the bytes oject struct which contains a []byte value
type Bytes struct {
	Value []byte
}

// Type returns the string object type
func (b *Bytes) Type() Type { return BYTES_OBJ }

// Inspect returns the string value
func (b *Bytes) Inspect() string { return fmt.Sprintf("%#v", b.Value) }

// BuiltinFunction is the type that will allow us to support
// adding functions from the host language (ie. go)
type BuiltinFunction func(args ...Object) Object

// Builtin is the Builtin function object struct
type Builtin struct {
	Fun BuiltinFunction
}

// Type returns the BUILTIN_OBJ type string
func (b *Builtin) Type() Type { return BUILTIN_OBJ }

// Inspect returns "builtin function"
func (b *Builtin) Inspect() string { return "builtin function" }

// BuiltinObj allows us to define a map object to be used for any builtins
// that work better as a sort of module
type BuiltinObj struct {
	Obj Object
}

func (bo *BuiltinObj) Type() Type { return BUILTIN_OBJ }

func (bo *BuiltinObj) Inspect() string { return "builtin object" }

// List is the list object type struct
type List struct {
	Elements []Object
}

// Type returns the list object type
func (l *List) Type() Type { return LIST_OBJ }

// Inspect returns the stringified version of the list
func (l *List) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range l.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

// ListCompLiteral is the list comprehension object struct
type ListCompLiteral struct {
	Elements []Object
}

// Type returns the list comprehension object type string
func (lcl *ListCompLiteral) Type() Type { return LIST_COMP_OBJ }

// Inspect returns a string representation of the lcl object
func (lcl *ListCompLiteral) Inspect() string {
	var out bytes.Buffer
	elements := []string{}
	for _, e := range lcl.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

// MapPair is a pair of key objects to value objects
type MapPair struct {
	Key   Object
	Value Object
}

// Map is the map object type struct
type Map struct {
	Pairs OrderedMap2[HashKey, MapPair] // Pairs is the map of HashKey to other MapPair objects
}

// Type returns the map object type
func (m *Map) Type() Type { return MAP_OBJ }

// Inspect returns the stringified version of the map
func (m *Map) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, key := range m.Pairs.Keys {
		pair, _ := m.Pairs.Get(key)
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

// hashMap hashes the entire map to be used for checking equality
func (m *Map) hashMap() uint64 {
	h := fnv.New64a()
	for _, k := range m.Pairs.Keys {
		v, _ := m.Pairs.Get(k)
		// Just using xor as a way to get a unique uint64 with the value hash
		hashedKeyObj := k.Value ^ HashObject(v.Value)
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, hashedKeyObj)
		h.Write(b)
	}
	return h.Sum64()
}

// MapCompLiteral is the map comprehension object struct
type MapCompLiteral struct {
	Pairs OrderedMap2[HashKey, MapPair] // Pairs is the map of HashKey to other MapPair objects
}

// Type returns the map comprehension object type string
func (mcl *MapCompLiteral) Type() Type { return MAP_COMP_OBJ }

// Inspect returns a string representation of the mcl object
func (mcl *MapCompLiteral) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, key := range mcl.Pairs.Keys {
		pair, _ := mcl.Pairs.Get(key)
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

// SetPair is the set object and bool to represent its precense in the set
type SetPair struct {
	Value   Object
	Present bool
}

// Set is the set object type struct
type Set struct {
	Elements *OrderedMap2[uint64, SetPair]
}

func NewSetElements() *OrderedMap2[uint64, SetPair] {
	return NewOrderedMap[uint64, SetPair]()
}

// Type returns the Set object type
func (s *Set) Type() Type { return SET_OBJ }

// Inspect returns the stringified version of the set
func (s *Set) Inspect() string {
	var out bytes.Buffer

	out.WriteString("{")
	keys := s.Elements.Keys
	for i, k := range keys {
		e, ok := s.Elements.Get(k)
		if !ok {
			continue
		}
		endStr := ""
		if i != len(keys)-1 {
			endStr = ", "
		}
		out.WriteString(e.Value.Inspect() + endStr)
	}
	out.WriteString("}")
	return out.String()
}

// SetCompLiteral is the set comprehension object struct
type SetCompLiteral struct {
	Elements map[uint64]SetPair
}

// Type returns the list comprehension object type string
func (scl *SetCompLiteral) Type() Type { return SET_COMP_OBJ }

// Inspect returns a string representation of the scl object
func (scl *SetCompLiteral) Inspect() string {
	var out bytes.Buffer
	elements := []string{}
	for _, e := range scl.Elements {
		elements = append(elements, e.Value.Inspect())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("}")

	return out.String()
}

// Module is the type that represents imported values
type Module struct {
	Name string
	Env  *Environment
}

// Type returns the module object type
func (m *Module) Type() Type { return MODULE_OBJ }

// Inspect only returns the modules name for now
func (m *Module) Inspect() string {
	return fmt.Sprintf("Module '%s'", m.Name)
}

// For loop stuff
type BreakStatement struct{}

func (bks *BreakStatement) Type() Type {
	return BREAK_OBJ
}

func (bks *BreakStatement) Inspect() string {
	return "break;"
}

type ContinueStatement struct{}

func (cs *ContinueStatement) Type() Type {
	return CONTINUE_OBJ
}

func (cs *ContinueStatement) Inspect() string {
	return "continue;"
}

// ------------------------------- HashKey Stuff --------------------------------

// TODO: cache the return value of HashKey to improve performance

// Hashable allows us to check if an object is hashable
type Hashable interface {
	HashKey() HashKey
}

// HashKey is the hash key for any of the object types we want to use in maps
type HashKey struct {
	Type  Type   // Type is the objects type
	Value uint64 // Value is the value of the "hash" key
}

// HashKey implements hashing for boolean values
func (b *Boolean) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

// HashKey implements hashing for integer values
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

// HashKey implements hashing for unsigned integer values
func (ui *UInteger) HashKey() HashKey {
	return HashKey{Type: ui.Type(), Value: ui.Value}
}

// HashKey implements hashing for string values it uses a
// hash method builtin from golang
func (s *Stringo) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

// hashList implements hashing for list objects
func (l *List) hashList() uint64 {
	h := fnv.New64a()
	for _, obj := range l.Elements {
		hashedObj := HashObject(obj)
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, hashedObj)
		h.Write(b)
	}
	return h.Sum64()
}

// HashObject is a generic function to hash any of the hashable object types
// It is very likely I wont keep it like this because it will probably break things
// but for now this naive implementation should do
func HashObject(obj Object) uint64 {
	h := fnv.New64a()
	switch obj.Type() {
	case INTEGER_OBJ:
		b := make([]byte, 8)
		binary.PutVarint(b, obj.(*Integer).Value)
		h.Write(b)
	case UINTEGER_OBJ:
		b := make([]byte, 8)
		binary.PutUvarint(b, obj.(*UInteger).Value)
		h.Write(b)
	case BOOLEAN_OBJ:
		if obj.(*Boolean).Value {
			// Use 1 for true
			return 1
		}
		return 0
	case NULL_OBJ:
		// Use 2 for null
		return 2
	case FLOAT_OBJ:
		b := make([]byte, 8)
		binary.PutUvarint(b, uint64(obj.(*Float).Value))
		h.Write(b)
	case STRING_OBJ:
		s := []byte(obj.(*Stringo).Value)
		h.Write([]byte(s))
	case FUNCTION_OBJ:
		// TODO: This is a naive way of determining if two functions are identical
		// come back and fix this or make it smarter if possible
		funObj := obj.(*Function).Inspect()
		h.Write([]byte(funObj))
	case ERROR_OBJ:
		// Although i dont think this should happen, lets give it a hash anyways
		h.Write([]byte(obj.(*Error).Message))
	case LIST_OBJ:
		b := make([]byte, 8)
		listObj := obj.(*List)
		binary.BigEndian.PutUint64(b, listObj.hashList())
		h.Write(b)
	case MAP_OBJ:
		b := make([]byte, 8)
		mapObj := obj.(*Map)
		binary.BigEndian.PutUint64(b, mapObj.hashMap())
		h.Write(b)
	case BYTES_OBJ:
		h.Write(obj.(*Bytes).Value)
	default:
		fmt.Printf("This is the object trying to be hashed = %v\n\n", obj)
		fmt.Printf("Unsupported hashable object: %T\n", obj)
	}
	return h.Sum64()
}
