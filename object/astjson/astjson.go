// Package astjson converts blue AST expression nodes into objects. It is
// the historical implementation behind from_json, which parsed JSON using
// the blue lexer/parser. The builtin now uses the dedicated converter in
// package object (object/json.go); this package remains so parity tests can
// compare both implementations on a corpus.
//
// It lives in its own subpackage because it imports blue/ast, which the
// core runtime packages (object, vm, binc) must not depend on.
package astjson

import (
	"fmt"
	"log"
	"math/big"
	"sort"

	"blue/ast"
	"blue/object"
)

func isError(o object.Object) bool {
	return o != nil && o.Type() == object.ERROR_OBJ
}

func parseMapLiteral(node *ast.MapLiteral) object.Object {
	pairs := object.NewPairsMapWithSize(len(node.Pairs))

	indices := []int{}
	for k := range node.PairsIndex {
		indices = append(indices, k)
	}
	sort.Ints(indices)
	for _, i := range indices {
		keyNode := node.PairsIndex[i]
		valueNode := node.Pairs[keyNode]
		// Should always be an *ast.StringLiteral
		key := ParseJson(keyNode)
		if isError(key) {
			return key
		}
		// Should always be true
		ok := object.IsHashable(key)
		if !ok {
			return newErr("unusable as a map key: %s", key.Type())
		}
		hk := object.HashObject(key)
		hashed := object.HashKey{Type: key.Type(), Value: hk}

		value := ParseJson(valueNode)
		if isError(value) {
			return value
		}

		pairs.Set(hashed, object.MapPair{Key: key, Value: value})
	}

	return &object.Map{Pairs: pairs}
}

func parseListLiteral(node *ast.ListLiteral) object.Object {
	result := make([]object.Object, len(node.Elements))
	for i, e := range node.Elements {
		result[i] = ParseJson(e)
	}
	return &object.List{Elements: result}
}

// ParseJson converts an AST expression (as produced by parsing a JSON
// document with the blue parser) into an object.
func ParseJson(expr ast.Expression) object.Object {
	switch t := expr.(type) {
	case *ast.IntegerLiteral:
		return &object.Integer{Value: t.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: t.Value}
	case *ast.BigIntegerLiteral:
		return &object.BigInteger{Value: t.Value}
	case *ast.BigFloatLiteral:
		return &object.BigFloat{Value: t.Value}
	case *ast.Boolean:
		if t.Value {
			return object.TRUE
		}
		return object.FALSE
	case *ast.Null:
		return object.NULL
	case *ast.StringLiteral:
		return &object.Stringo{Value: t.Value}
	case *ast.MapLiteral:
		return parseMapLiteral(t)
	case *ast.ListLiteral:
		return parseListLiteral(t)
	case *ast.PrefixExpression:
		if t.TokenLiteral() != "-" {
			panic("Unexpected Prefix Expression Token " + t.TokenLiteral())
		}
		right := ParseJson(t.Right)
		switch rt := right.(type) {
		case *object.Integer:
			rt.Value = -rt.Value
		case *object.Float:
			rt.Value = -rt.Value
		case *object.BigInteger:
			bi := new(big.Int)
			rt.Value = bi.Neg(rt.Value)
		case *object.BigFloat:
			rt.Value = rt.Value.Neg()
		default:
			panic("Unexpected Type for Prefix Expression " + right.Type())
		}
		return right
	default:
		log.Fatalf("ParseJson: UNHANDLED t = %#+v (%T)", t, t)
	}
	panic("UNREACHABLE")
}

func newErr(format string, a ...any) object.Object {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}
