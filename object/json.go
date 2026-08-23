package object

import (
	"encoding/json"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// This file implements from_json WITHOUT the blue lexer/parser: JSON is far
// simpler than blue, and keeping this builtin free of the parser lets
// minimal VM-only builds (no lexer/parser/compiler) still use it.
//
// The old implementation parsed JSON with the blue parser via ParseJson
// (see package object/astjson). That path is kept for parity tests.

// maxJSONDepth bounds nesting so malformed deeply-nested input errors out
// instead of overflowing the stack.
const maxJSONDepth = 512

// FromJsonString converts a JSON string into blue objects. The returned
// Object is an *Error when the input is not valid JSON or exceeds the
// nesting limit (matching how blue builtins report failures).
func FromJsonString(s string) Object {
	if !json.Valid([]byte(s)) {
		return newError("`from_json` error: invalid json: %q", s)
	}
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	obj := decodeJsonValue(dec, 0)
	if isError(obj) {
		return obj
	}
	// Ensure there is no trailing content after the first value
	if _, err := dec.Token(); err == nil {
		return newError("`from_json` error: too many values in json string")
	}
	return obj
}

func jsonErr(format string, a ...any) Object {
	return newError("`from_json` error: "+format, a...)
}

func decodeJsonValue(dec *json.Decoder, depth int) Object {
	if depth > maxJSONDepth {
		return jsonErr("json nested too deep (max %d)", maxJSONDepth)
	}
	tok, err := dec.Token()
	if err != nil {
		return jsonErr("%s", err.Error())
	}
	return convertJsonToken(dec, tok, depth)
}

func convertJsonToken(dec *json.Decoder, tok json.Token, depth int) Object {
	switch t := tok.(type) {
	case json.Number:
		return convertJsonNumber(string(t))
	case string:
		return &Stringo{Value: t}
	case bool:
		if t {
			return TRUE
		}
		return FALSE
	case nil:
		return NULL
	case json.Delim:
		switch t {
		case '[':
			return decodeJsonList(dec, depth)
		case '{':
			return decodeJsonMap(dec, depth)
		default:
			return jsonErr("unexpected delimiter %q", t)
		}
	default:
		return jsonErr("unexpected token %T", tok)
	}
}

func decodeJsonList(dec *json.Decoder, depth int) Object {
	elems := []Object{}
	for {
		next, err := dec.Token()
		if err != nil {
			return jsonErr("%s", err.Error())
		}
		if d, ok := next.(json.Delim); ok && d == ']' {
			break
		}
		elem := convertJsonToken(dec, next, depth+1)
		if isError(elem) {
			return elem
		}
		elems = append(elems, elem)
	}
	return &List{Elements: elems}
}

func decodeJsonMap(dec *json.Decoder, depth int) Object {
	pairs := NewPairsMap()
	for {
		keyTok, err := dec.Token()
		if err != nil {
			return jsonErr("%s", err.Error())
		}
		if d, ok := keyTok.(json.Delim); ok && d == '}' {
			break
		}
		keyStr, ok := keyTok.(string)
		if !ok {
			return jsonErr("map keys must be strings, got %T", keyTok)
		}
		value := decodeJsonValue(dec, depth+1)
		if isError(value) {
			return value
		}
		key := &Stringo{Value: keyStr}
		hk := HashKey{Type: key.Type(), Value: HashObject(key)}
		pairs.Set(hk, MapPair{Key: key, Value: value})
	}
	return &Map{Pairs: pairs}
}

func convertJsonNumber(s string) Object {
	if !strings.ContainsAny(s, ".eE") {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return &Integer{Value: i}
		}
		bi, ok := new(big.Int).SetString(s, 10)
		if ok {
			return &BigInteger{Value: bi}
		}
	} else if f, exact, dec, ok := jsonNumberAsFloat64(s); ok {
		if exact {
			return &Float{Value: f}
		}
		return &BigFloat{Value: *dec}
	}
	// Numbers outside every supported range fall back to big float
	d, err := decimal.NewFromString(s)
	if err != nil {
		return jsonErr("unsupported number %q", s)
	}
	return &BigFloat{Value: d}
}

// jsonNumberAsFloat64 mirrors the parser's ExactFloat64 semantics for float
// literals (see parser/parser.go): a JSON number becomes a Float only when
// its decimal value survives the float64 round-trip exactly; otherwise it
// is promoted to BigFloat so no precision is silently lost.
func jsonNumberAsFloat64(s string) (float64, bool, *decimal.Decimal, bool) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false, nil, false
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return f, true, nil, true
	}
	fromRounded := decimal.NewFromFloat(f)
	fromString, err := decimal.NewFromString(s)
	if err != nil {
		return 0, false, nil, false
	}
	if fromRounded.Equal(fromString) {
		return f, true, nil, true
	}
	return 0, false, &fromString, true
}
