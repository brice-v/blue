package object

import (
	"fmt"
	"strings"
)

type Inspector struct {
	seen map[Object]struct{}
}

func newInspector() *Inspector {
	return &Inspector{seen: make(map[Object]struct{})}
}

func (c *Inspector) enter(obj Object) bool {
	if _, ok := c.seen[obj]; ok {
		return false
	}
	c.seen[obj] = struct{}{}
	return true
}

func (c *Inspector) leave(obj Object) {
	delete(c.seen, obj)
}

func (c *Inspector) inspect(obj Object) string {
	if obj == nil {
		return "null"
	}

	switch x := obj.(type) {
	case *List:
		if !c.enter(x) {
			return "[...]"
		}
		defer c.leave(x)
		return "[" + c.inspectList(x.Elements) + "]"

	case *ListCompLiteral:
		if !c.enter(x) {
			return "[...]"
		}
		defer c.leave(x)
		return "[" + c.inspectList(x.Elements) + "]"

	case *Map:
		if !c.enter(x) {
			return "{...:...}"
		}
		defer c.leave(x)
		pairs := make([]string, 0, x.Pairs.Len())
		for _, k := range x.Pairs.Keys {
			pair, _ := x.Pairs.Get(k)
			pairs = append(pairs, fmt.Sprintf("%s: %s", c.inspect(pair.Key), c.inspect(pair.Value)))
		}
		return "{" + strings.Join(pairs, ", ") + "}"

	case *MapCompLiteral:
		if !c.enter(x) {
			return "{...:...}"
		}
		defer c.leave(x)
		pairs := make([]string, 0, x.Pairs.Len())
		for _, k := range x.Pairs.Keys {
			pair, _ := x.Pairs.Get(k)
			pairs = append(pairs, fmt.Sprintf("%s: %s", c.inspect(pair.Key), c.inspect(pair.Value)))
		}
		return "{" + strings.Join(pairs, ", ") + "}"

	case *Set:
		if !c.enter(x) {
			return "{...}"
		}
		defer c.leave(x)
		elems := make([]string, 0, x.Elements.Len())
		for _, k := range x.Elements.Keys {
			pair, _ := x.Elements.Get(k)
			elems = append(elems, c.inspect(pair.Value))
		}
		return "{" + strings.Join(elems, ", ") + "}"

	case *SetCompLiteral:
		if !c.enter(x) {
			return "{...}"
		}
		defer c.leave(x)
		elems := make([]string, 0, len(x.Elements))
		for _, pair := range x.Elements {
			elems = append(elems, c.inspect(pair.Value))
		}
		return "{" + strings.Join(elems, ", ") + "}"

	case *BlueStruct:
		if !c.enter(x) {
			return "@{...}"
		}
		defer c.leave(x)
		pairs := make([]string, 0, len(x.Fields))
		for i, field := range x.Fields {
			pairs = append(pairs, fmt.Sprintf("%s: %s", field, c.inspect(x.Values[i])))
		}
		return "@{" + strings.Join(pairs, ", ") + "}"

	case *DefaultArgs:
		if !c.enter(x) {
			return "DefaultArgs{...}"
		}
		defer c.leave(x)
		pairs := make([]string, 0, len(x.Value))
		for k, v := range x.Value {
			pairs = append(pairs, fmt.Sprintf("%s=%s", k, c.inspect(v)))
		}
		return "DefaultArgs{" + strings.Join(pairs, ", ") + "}"

	case *ReturnValue:
		return c.inspect(x.Value)
	}

	return obj.Inspect()
}

func (c *Inspector) inspectList(elements []Object) string {
	parts := make([]string, len(elements))
	for i, e := range elements {
		parts[i] = c.inspect(e)
	}
	return strings.Join(parts, ", ")
}
