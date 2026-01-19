package object

import (
	"blue/ast"
	"blue/code"
	"blue/consts"
	"bytes"
	"fmt"
	"hash/maphash"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

const (
	// INTEGER_OBJ is the integer object type string
	INTEGER_OBJ Type = "INTEGER"
	// BIG_INTEGER_OBJ is the big integer object type string
	BIG_INTEGER_OBJ Type = "BIG_INTEGER"
	// BOOLEAN_OBJ is the boolean object type string
	BOOLEAN_OBJ Type = "BOOLEAN"
	// NULL_OBJ is the null object type string
	NULL_OBJ Type = "NULL"
	// UINTEGER_OBJ is the uint object type string
	UINTEGER_OBJ Type = "UINTEGER"
	// FLOAT_OBJ is the float object type string
	FLOAT_OBJ Type = "FLOAT"
	// BIG_FLOAT_OBJ is the big float object type string
	BIG_FLOAT_OBJ Type = "BIG_FLOAT"
	// RETURN_VALUE_OBJ is the return object type string
	RETURN_VALUE_OBJ Type = "RETURN_VALUE"
	// ERROR_OBJ is the error object type string
	ERROR_OBJ Type = "ERROR"
	// FUNCTION_OBJ is the function object type string
	FUNCTION_OBJ Type = "FUNCTION"
	// STRING_OBJ is the string object type string
	STRING_OBJ Type = "STRING"
	// BYTES_OBJ is the bytes object type string
	BYTES_OBJ Type = "BYTES"
	// GO_OBJ is the go object type string
	GO_OBJ Type = "GO_OBJ"
	// REGEX_OBJ is the Regex object type string
	REGEX_OBJ Type = "REGEX"
	// BUILTIN_OBJ is the builtin function object type string
	BUILTIN_OBJ Type = "BUILTIN"
	// LIST_OBJ is the list object type string
	LIST_OBJ Type = "LIST"
	// MAP_OBJ is the map object type string
	MAP_OBJ Type = "MAP"
	// SET_OBJ is the set object type
	SET_OBJ Type = "SET"
	// LIST_COMP_OBJ is the list comprehension literal type string
	LIST_COMP_OBJ Type = "LIST_COMP_OBJ"
	// MAP_COMP_OBJ is the map comprehension literal type string
	MAP_COMP_OBJ Type = "MAP_COMP_OBJ"
	// SET_COMP_OBJ is the set comprehension literal type string
	SET_COMP_OBJ Type = "SET_COMP_OBJ"
	// MODULE_OBJ is the object type for an imported module
	MODULE_OBJ Type = "MODULE_OBJ"
	// PROCESS_OBJ is the process type for a process
	PROCESS_OBJ Type = "PROCESS"
	// BLUE_STRUCT_OBJ is the struct object type
	BLUE_STRUCT_OBJ Type = "BLUE_STRUCT"

	// BREAK_OBJ is the break statement type
	BREAK_OBJ Type = "BREAK_OBJ"
	// CONTINUE_OBJ is the continue statement type
	CONTINUE_OBJ Type = "CONTINUE_OBJ"

	// FOR VM

	// COMPILED_FUNCTION_OBJ is the compiled function object for the VM
	COMPILED_FUNCTION_OBJ = "COMPILED_FUNCTION_OBJ"
	// CLOSURE_OBJ is the closure function object for the VM
	CLOSURE_OBJ = "CLOSURE_OBJ"
	// EXEC_STRING is the string object type string
	EXEC_STRING_OBJ = "EXEC_STRING"
	// IGNORE_OBJ is the IGNORE object type string
	IGNORE_OBJ Type = "IGNORE"
	// DEFAULT_ARGS_OBJ is the default args object for use in vm functions
	DEFAULT_ARGS_OBJ = "DEFAULT_ARGS_OBJ"
)

// Type is the object type represented as a string
type Type string

// Object is the interface a value in the language must
// satisfy to be used
type Object interface {
	Type() Type      // Type is a function that returns the objects type
	Inspect() string // Inspect is used for debugging an object
	Help() string    // Help is used to get the help string for an object
	Encode() ([]byte, error)
	IType() iType
}

// Integer is the integer object type
type Integer struct {
	Value int64 // Value is the internal rep. of an integer, it is stored as an int64
}

var (
	// TRUE is the true object which should be the same everywhere
	TRUE = &Boolean{Value: true}
	// FALSE is the false object which should be the same everywhere
	FALSE = &Boolean{Value: false}
	// NULL is the null object which should be the same everywhere
	NULL = &Null{}
	// IGNORE is the object which is used to ignore variables when necessary
	IGNORE = &Null{}
	// VM_IGNORE is the object with is used to ignore things in match expressions
	VM_IGNORE = &Ignore{}

	// BREAK is the break object to be used the same everywhere
	BREAK = &BreakStatement{}
	// CONTINUE is the continue object to be used the same everywhere
	CONTINUE = &ContinueStatement{}

	// USE_PARAM_STR_OBJ is used for constant folding as its used by vm to setup functions with default params
	USE_PARAM_STR_OBJ = &Stringo{Value: USE_PARAM_STR}
)

const USE_PARAM_STR = "___USE_PARAM___"

// These are used for VM/Compiler constant folding
var OBJECT_CONSTANTS = NewObjectConstants()

func NewObjectConstants() []Object {
	return []Object{TRUE, FALSE, NULL, IGNORE, VM_IGNORE, BREAK, CONTINUE, USE_PARAM_STR_OBJ}
}

// Maybe replace with switch or map if this is really a problem
func IsConstantObject(obj Object) int {
	for i, o := range OBJECT_CONSTANTS {
		if o == obj {
			return i
		}
	}
	return -1
}

// Inspect returns the string value of the integer object
func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }

// Type returns the object type of integer
func (i *Integer) Type() Type { return INTEGER_OBJ }

func (i *Integer) Help() string {
	desc := fmt.Sprintf("is the object that represents numerical values %d to %d", math.MinInt64, math.MaxInt64)
	return createHelpStringForObject("Integer", desc, i)
}

// BigInteger is the big integer type
type BigInteger struct {
	Value *big.Int
}

// Inspect returns the string value of big integer
func (bi *BigInteger) Inspect() string { return bi.Value.String() }

// Type returns the object type of big integer
func (bi *BigInteger) Type() Type { return BIG_INTEGER_OBJ }

func (bi *BigInteger) Help() string {
	return createHelpStringForObject("BigInteger", "is the object that represents numerical values outside of the Integer range", bi)
}

// Boolean is the boolean object type
type Boolean struct {
	Value bool // Value is the internal rep. of a boolean, it is stored as a bool
}

// Inspect returns the string value of the boolean object
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }

// Type returns the object type of boolean
func (b *Boolean) Type() Type { return BOOLEAN_OBJ }

func (b *Boolean) Help() string {
	return createHelpStringForObject("Boolean", "is the object that represents true or false", b)
}

// Null is the null object struct
type Null struct{}

// Type is the object type of null
func (n *Null) Type() Type { return NULL_OBJ }

// Inspect returns the string value of null
func (n *Null) Inspect() string { return "null" }

func (n *Null) Help() string {
	return createHelpStringForObject("Null", "is the null object", n)
}

// Ignore is the Ignore object struct -> which is specifically used for matching
type Ignore struct{}

// Type is the object type of null
func (n *Ignore) Type() Type { return IGNORE_OBJ }

// Inspect returns the string value of ignore
func (n *Ignore) Inspect() string { return "##IGNORE##" }

func (n *Ignore) Help() string {
	return "SHOULD NOT BE CALLED"
}

// DefaultArgs is the DefaultArgs object struct -> which is specifically used for matching
type DefaultArgs struct {
	Value map[string]Object
}

// Type is the object type of null
func (n *DefaultArgs) Type() Type { return DEFAULT_ARGS_OBJ }

// Inspect returns the string value of DefaultArgs
func (n *DefaultArgs) Inspect() string {
	var out bytes.Buffer
	out.WriteString("DefaultArgs{")
	for k, v := range n.Value {
		fmt.Fprintf(&out, "%s=%s, ", k, v.Inspect())
	}
	s := out.String()
	s = strings.TrimSuffix(s, ", ")
	return s + "}"
}

func (n *DefaultArgs) Help() string {
	return "SHOULD NOT BE CALLED"
}

// UInteger is the hex, octal, bin object struct
type UInteger struct {
	Value uint64
}

// Type returns the UINTEGER_OBJ type
func (ui *UInteger) Type() Type { return UINTEGER_OBJ }

// Inspect returns the string value of the uint
func (ui *UInteger) Inspect() string { return fmt.Sprintf("%d", ui.Value) }

func (ui *UInteger) Help() string {
	desc := "is the object that represents numerical values 0 to 18446744073709551615"
	return createHelpStringForObject("UInteger", desc, ui)
}

// Float is the float object struct
type Float struct {
	Value float64
}

// Type returns the FLOAT_OBJ type
func (f *Float) Type() Type { return FLOAT_OBJ }

// Inspect returns the string value of the float
func (f *Float) Inspect() string { return strconv.FormatFloat(f.Value, 'g', -1, 64) }

func (f *Float) Help() string {
	desc := fmt.Sprintf("is the object that represents numerical values %f to %f", math.SmallestNonzeroFloat64, math.MaxFloat64)
	return createHelpStringForObject("Float", desc, f)
}

// BigFloat is the big float object struct
type BigFloat struct {
	Value decimal.Decimal
}

// Inspect returns the big float object as a string
func (bf BigFloat) Inspect() string { return bf.Value.String() }

// Type returns the big float object type
func (bf BigFloat) Type() Type { return BIG_FLOAT_OBJ }

func (bf BigFloat) Help() string {
	return createHelpStringForObject("BigFloat", "is the object that represents numerical values outside of the Float range", bf)
}

// ReturnValue is the struct type for the return value object
type ReturnValue struct {
	Value Object
}

// Type returns the return value object type
func (rv *ReturnValue) Type() Type { return RETURN_VALUE_OBJ }

// Inspect returns the string version of the object to return
func (rv *ReturnValue) Inspect() string { return rv.Value.Inspect() }

func (rv *ReturnValue) Help() string {
	return createHelpStringForObject("ReturnValue", "is the object that represents a return value from a function or block", rv)
}

// Error is the error object struct.  It conatins a message as a string
type Error struct {
	Message string
}

// Type returns the error object type
func (e *Error) Type() Type { return ERROR_OBJ }

// Inspect returns a string representation of the error
func (e *Error) Inspect() string { return consts.EVAL_ERROR_PREFIX + e.Message }

func (e *Error) Help() string {
	return createHelpStringForObject("Error", "is the object that represents an error raised during runtime execution", e)
}

// Just used for encoding/decoding functions
type StringFunction struct {
	Value string
}

func (f *StringFunction) Type() Type { return FUNCTION_OBJ }

func (f *StringFunction) Inspect() string {
	return f.Value
}

func (f *StringFunction) Help() string {
	return ""
}

// Function is the function object struct
type Function struct {
	Parameters []*ast.Identifier   // Parameters is a slice of identifiers
	Body       *ast.BlockStatement // Body is a block statement node
	Env        *Environment        // Env stores the function's environment

	DefaultParameters []Object // DefaultParameters holds the expression of the default parameter, if it exists otherwise nil

	HelpStr string
}

// Type returns the function objects type
func (f *Function) Type() Type { return FUNCTION_OBJ }

// Inspect returns the string representation of the function
func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for i, p := range f.Parameters {
		dp := f.DefaultParameters[i]
		if dp != nil {
			params = append(params, fmt.Sprintf("%s=%s", p.String(), dp.Inspect()))
		} else {
			params = append(params, p.String())
		}
	}

	out.WriteString("fun(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}

func (f *Function) Help() string {
	return f.HelpStr
}

type Process struct {
	Ch chan Object
	Id uint64

	NodeName string
}

func (p *Process) Inspect() string {
	return fmt.Sprintf("#{name: %q, id: %d}", p.NodeName, p.Id)
}

func (p *Process) Type() Type {
	return PROCESS_OBJ
}

func (p *Process) Help() string {
	return createHelpStringForObject("Process", "is the object that represents a goroutine process with an associated channel", p)
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

func (s *Stringo) Help() string {
	return createHelpStringForObject("String", "is the utf-8 bytes representation of a string object", s)
}

// Bytes is the bytes object struct which contains a []byte value
type Bytes struct {
	Value []byte
}

// Type returns the bytes object type string
func (b *Bytes) Type() Type { return BYTES_OBJ }

// Inspect returns the byte slice as it is (in format %#v)
func (b *Bytes) Inspect() string { return fmt.Sprintf("%#v", b.Value) }

func (b *Bytes) Help() string {
	return createHelpStringForObject("Bytes", "is the object that represents a slice of arbitrary bytes", b)
}

// GoObj is the go object struct which contains a generic value
type GoObj[T any] struct {
	Value T
	Id    uint64
}

// Type return the go object type string
func (g *GoObj[T]) Type() Type {
	return GO_OBJ
}

// Inspect returns the string representation of the GoObj with
func (g *GoObj[T]) Inspect() string {
	return fmt.Sprintf("GoObj{Type: (%T), ID: %x}", g.Value, g.Id)
}

func (g *GoObj[T]) Help() string {
	return createHelpStringForObject("GoObj", "is the object that represents an arbitrary go object", g)
}

// Note: This is disabled for now due to the complexity of handling all Go Object Types supported by blue
// func (g *GoObj[T]) Decoder(data []byte) (any, error) {
// 	var goObj GoObj[T]
// 	err := gob.NewDecoder(bytes.NewReader(data)).Decode(&goObj)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return goObj, nil
// }

// Just used for encoding/decoding go objects
type GoObjectGob struct {
	T     string
	Value []byte
}

func (g *GoObjectGob) Type() Type { return FUNCTION_OBJ }

func (g *GoObjectGob) Inspect() string {
	return fmt.Sprintf("GoObjectGob{Type: %s, Value: %v}", g.T, g.Value)
}

func (g *GoObjectGob) Help() string {
	return ""
}

// Regex is the string oject struct which contains a string value
type Regex struct {
	Value *regexp.Regexp
}

// Type returns the string object type
func (r *Regex) Type() Type { return REGEX_OBJ }

// Inspect returns the string value
func (r *Regex) Inspect() string { return "/" + r.Value.String() + "/" }

func (r *Regex) Help() string {
	return createHelpStringForObject("Regex", "is the object that represents the Regex", r)
}

// BuiltinFunction is the type that will allow us to support
// adding functions from the host language (ie. go)
type BuiltinFunction func(args ...Object) Object

// Builtin is the Builtin function object struct
type Builtin struct {
	Fun     BuiltinFunction
	HelpStr string

	Mutates bool // Mutates signifies whether this function can mutate its arguments
}

// Type returns the BUILTIN_OBJ type string
func (b *Builtin) Type() Type { return BUILTIN_OBJ }

// Inspect returns "builtin function"
func (b *Builtin) Inspect() string { return "builtin function" }

func (b *Builtin) Help() string {
	// TODO: Do we use createHelpStringForObject()?
	return fmt.Sprintf("%s\n    type = '%s'\n    inspect = '%s'", b.HelpStr, b.Type(), b.Inspect())
}

// BuiltinObj allows us to define a map object to be used for any builtins
// that work better as a sort of module
type BuiltinObj struct {
	Obj     Object
	HelpStr string
}

func (bo *BuiltinObj) Type() Type { return BUILTIN_OBJ }

func (bo *BuiltinObj) Inspect() string { return "builtin object" }

func (bo *BuiltinObj) Help() string {
	return fmt.Sprintf("%s\n    type = '%s'\n    inspect = '%s'", bo.HelpStr, bo.Type(), bo.Inspect())
}

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

func (l *List) Help() string {
	return createHelpStringForObject("List", "is the object that represents an arbitrary list of objects", l)
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

func (lcl *ListCompLiteral) Help() string {
	return createHelpStringForObject("ListCompLiteral", "is the object that represents a List Comprehension", lcl)
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

func (m *Map) Help() string {
	return createHelpStringForObject("Map", "is the object that represents a key value pair where keys and values are arbitrary objects", m)
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

func (mcl *MapCompLiteral) Help() string {
	return createHelpStringForObject("MapCompLiteral", "is the object that represents a Map Comprehension", mcl)
}

// SetPair is the set object and bool to represent its precense in the set
type SetPair struct {
	Value   Object
	Present struct{}
}

type SetPairGo struct {
	Value   any
	Present struct{}
}

// Set is the set object type struct
type Set struct {
	Elements *OrderedMap2[uint64, SetPair]
}

func NewSetElements() *OrderedMap2[uint64, SetPair] {
	return NewOrderedMap[uint64, SetPair]()
}

func NewSetElementsWithSize(size int) *OrderedMap2[uint64, SetPair] {
	return NewOrderedMapWithSize[uint64, SetPair](size)
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

func (s *Set) Help() string {
	return createHelpStringForObject("Set", "is the object that represents a set of arbitrary objects", s)
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

func (scl *SetCompLiteral) Help() string {
	return createHelpStringForObject("SetCompLiteral", "is the object that represents a Set Comprehension", scl)
}

// Module is the type that represents imported values
type Module struct {
	Name string
	Env  *Environment

	HelpStr string
}

// Type returns the module object type
func (m *Module) Type() Type { return MODULE_OBJ }

// Inspect only returns the modules name for now
func (m *Module) Inspect() string {
	return fmt.Sprintf("Module '%s'", m.Name)
}

func (m *Module) Help() string {
	return m.HelpStr
}

// For loop stuff
type BreakStatement struct{}

func (bks *BreakStatement) Type() Type {
	return BREAK_OBJ
}

func (bks *BreakStatement) Inspect() string {
	return "break;"
}

func (bks *BreakStatement) Help() string {
	return createHelpStringForObject("Break", "is the object that stops the execution of a loop right where it is and breaks out of the enclosing scope", bks)
}

type ContinueStatement struct{}

func (cs *ContinueStatement) Type() Type {
	return CONTINUE_OBJ
}

func (cs *ContinueStatement) Inspect() string {
	return "continue;"
}

func (cs *ContinueStatement) Help() string {
	return createHelpStringForObject("Continue", "is the object that stops the current execution and moves to the next iteration in the loop's scope", cs)
}

// BlueStruct is the struct object type struct
type BlueStruct struct {
	Fields []string
	Values []Object
}

// Type returns the blue struct object type
func (bs *BlueStruct) Type() Type { return BLUE_STRUCT_OBJ }

// Inspect returns the stringified version of the blue struct
func (bs *BlueStruct) Inspect() string {
	var out bytes.Buffer

	pairs := make([]string, 0, len(bs.Fields))
	for i := range bs.Fields {
		value := bs.Values[i]
		pairs = append(pairs, fmt.Sprintf("%s: %s", bs.Fields[i], value.Inspect()))
	}

	out.WriteString("@{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

func (bs *BlueStruct) Help() string {
	return createHelpStringForObject("BlueStruct", "is the object that represents a struct where keys are identifier strings and values are arbitary objects and the size or types cannot change", bs)
}

func (bs *BlueStruct) Get(name string) (Object, int) {
	for index, n := range bs.Fields {
		if n == name {
			return bs.Values[index], index
		}
	}
	return nil, -1
}

func (bs *BlueStruct) Set(index int, val Object) error {
	existingValue := bs.Values[index]
	if existingValue != nil && existingValue.Type() != val.Type() {
		return fmt.Errorf("failed to set on struct literal: existing value type = %s, new value type = %s", existingValue.Type(), val.Type())
	}
	bs.Values[index] = val
	return nil
}

// SetWithFieldName should just be used for testing
func (bs *BlueStruct) SetWithFieldName(name string, val Object) error {
	for index, n := range bs.Fields {
		if n == name {
			existingValue := bs.Values[index]
			if existingValue != nil && existingValue.Type() != val.Type() {
				return fmt.Errorf("failed to set on struct literal: existing value type = %s, new value type = %s", existingValue.Type(), val.Type())
			}
			bs.Values[index] = val
			return nil
		}
	}
	return fmt.Errorf("failed to set on struct literal: %s=%s", name, val.Inspect())
}

func NewBlueStruct(names []string, values []Object) (Object, error) {
	if len(names) != len(values) {
		return nil, fmt.Errorf("failed to create struct: names and values do not have equal lengths")
	}
	return &BlueStruct{Fields: names, Values: values}, nil
}

// CompiledFunction is the compiled function object struct
type CompiledFunction struct {
	Instructions  code.Instructions
	NumLocals     int
	NumParameters int

	Parameters          []string
	ParameterHasDefault []bool
	NumDefaultParams    int

	DisplayString string

	PosAlreadyIncremented map[int]struct{}
}

// Type returns the function objects type
func (cf *CompiledFunction) Type() Type { return COMPILED_FUNCTION_OBJ }

// Inspect returns the string representation of the function
func (cf *CompiledFunction) Inspect() string {
	return cf.DisplayString
}

func (cf *CompiledFunction) Help() string {
	return ""
}

// Closure is the compiled function object struct
type Closure struct {
	Fun  *CompiledFunction
	Free []Object
}

// Type returns the function objects type
func (cf *Closure) Type() Type { return CLOSURE_OBJ }

// Inspect returns the string representation of the function
func (cf *Closure) Inspect() string {
	return cf.Fun.DisplayString
}

func (cf *Closure) Help() string {
	return ""
}

type ExecString struct {
	Value string
}

// Type returns the string object type
func (s *ExecString) Type() Type { return EXEC_STRING_OBJ }

// Inspect returns the string value
func (s *ExecString) Inspect() string { return s.Value }

func (s *ExecString) Help() string {
	return ""
}

// ------------------------------- HashKey Stuff --------------------------------

// HashKey is the hash key for any of the object types we want to use in maps
type HashKey struct {
	Type  Type   // Type is the objects type
	Value uint64 // Value is the value of the "hash" key
}

var _seed = maphash.MakeSeed()

func newHasher() *maphash.Hash {
	h := &maphash.Hash{}
	h.SetSeed(_seed)
	return h
}

// hashList implements hashing for list objects
func (l *List) hashList() uint64 {
	hasher := newHasher()
	for _, obj := range l.Elements {
		hashedObj := HashObject(obj)
		maphash.WriteComparable(hasher, hashedObj)
	}
	return hasher.Sum64()
}

// hashSet implements hashing for set objects
func (s *Set) hashSet() uint64 {
	hasher := newHasher()
	for _, k := range s.Elements.Keys {
		maphash.WriteComparable(hasher, k)
	}
	return hasher.Sum64()
}

// hashMap hashes the entire map to be used for checking equality
func (m *Map) hashMap() uint64 {
	hasher := newHasher()
	for _, k := range m.Pairs.Keys {
		v, _ := m.Pairs.Get(k)
		// Just using xor as a way to get a unique uint64 with the value hash
		hashedKeyObj := k.Value ^ HashObject(v.Value)
		maphash.WriteComparable(hasher, hashedKeyObj)
	}
	return hasher.Sum64()
}

func (da *DefaultArgs) hashDefaultArgs() uint64 {
	hasher := newHasher()
	for k, v := range da.Value {
		hasher.WriteString(k)
		maphash.WriteComparable(hasher, HashObject(v))
	}
	return hasher.Sum64()
}

// hashStruct hashes the entire struct to be used for checking equality
func (bs *BlueStruct) hashStruct() uint64 {
	hasher := newHasher()
	for i, k := range bs.Fields {
		hasher.WriteString(k)
		maphash.WriteComparable(hasher, HashObject(bs.Values[i]))
	}
	return hasher.Sum64()
}

// HashObject is a generic function to hash any of the hashable object types
// It is very likely I wont keep it like this because it will probably break things
// but for now this naive implementation should do
// TODO: Update to be generic
func HashObject(obj Object) uint64 {
	hasher := newHasher()
	switch obj.Type() {
	case INTEGER_OBJ:
		maphash.WriteComparable(hasher, obj.(*Integer).Value)
	case UINTEGER_OBJ:
		maphash.WriteComparable(hasher, obj.(*UInteger).Value)
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
		maphash.WriteComparable(hasher, obj.(*Float).Value)
	case STRING_OBJ:
		hasher.WriteString(obj.(*Stringo).Value)
	case EXEC_STRING_OBJ:
		hasher.WriteString(obj.(*ExecString).Value)
	case FUNCTION_OBJ:
		// Note: This is a naive way of determining if two functions are identical
		// come back and fix this or make it smarter if possible
		hasher.WriteString(obj.(*Function).Inspect())
	case CLOSURE_OBJ:
		hasher.Write(obj.(*Closure).Fun.Instructions)
	case ERROR_OBJ:
		// Although i dont think this should happen, lets give it a hash anyways
		hasher.WriteString(obj.(*Error).Message)
	case LIST_OBJ:
		maphash.WriteComparable(hasher, obj.(*List).hashList())
	case SET_OBJ:
		maphash.WriteComparable(hasher, obj.(*Set).hashSet())
	case MAP_OBJ:
		maphash.WriteComparable(hasher, obj.(*Map).hashMap())
	case DEFAULT_ARGS_OBJ:
		maphash.WriteComparable(hasher, obj.(*DefaultArgs).hashDefaultArgs())
	case BLUE_STRUCT_OBJ:
		maphash.WriteComparable(hasher, obj.(*BlueStruct).hashStruct())
	case BYTES_OBJ:
		hasher.Write(obj.(*Bytes).Value)
	case BIG_FLOAT_OBJ:
		hasher.WriteString(obj.(*BigFloat).Value.String())
	case BIG_INTEGER_OBJ:
		hasher.WriteString(obj.(*BigInteger).Value.String())
	case GO_OBJ:
		hasher.WriteString(obj.Inspect())
	case REGEX_OBJ:
		hasher.WriteString(obj.(*Regex).Value.String())
	case PROCESS_OBJ:
		hasher.WriteString(obj.Inspect())
	case COMPILED_FUNCTION_OBJ:
		hasher.Write(obj.(*CompiledFunction).Instructions)
	case MODULE_OBJ:
		hasher.WriteString(obj.(*Module).Name)
	default:
		fmt.Printf("This is the object trying to be hashed = %v\n\n", obj)
		fmt.Printf("Unsupported hashable object: %T\n", obj)
	}
	return hasher.Sum64()
}

func IsHashable(obj Object) bool {
	t := obj.Type()
	return t == INTEGER_OBJ ||
		t == UINTEGER_OBJ ||
		t == BOOLEAN_OBJ ||
		t == NULL_OBJ ||
		t == FLOAT_OBJ ||
		t == STRING_OBJ ||
		t == FUNCTION_OBJ ||
		t == ERROR_OBJ ||
		t == LIST_OBJ ||
		t == SET_OBJ ||
		t == MAP_OBJ ||
		t == BYTES_OBJ ||
		t == BIG_FLOAT_OBJ ||
		t == BIG_INTEGER_OBJ ||
		t == GO_OBJ
}
