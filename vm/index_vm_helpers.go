package vm

import (
	"blue/object"
)

func (vm *VM) executeListIndexExpression(list, indx object.Object) error {
	listObj := list.(*object.List)
	var idx int64
	switch indx.Type() {
	case object.INTEGER_OBJ:
		idx = indx.(*object.Integer).Value
	case object.STRING_OBJ:
		// stringVal := indx.(*object.Stringo).Value
		// envVal, ok := e.env.Get(stringVal)
		// if !ok {
		// 	return object.NULL
		// }
		// intVal, ok := envVal.(*object.Integer)
		// if !ok {
		// 	return object.NULL
		// }
		// idx = intVal.Value
	case object.LIST_OBJ:
		// Handle range expressions (1..3) or (1..<3) => they come back as a list
		// indxList := indx.(*object.List).Elements
		// indexes := make([]int64, len(indxList))
		// for i, e := range indxList {
		// 	if e.Type() != object.INTEGER_OBJ {
		// 		return newError("index range needs to be INTEGER. got=%s", e.Type())
		// 	}
		// 	indexes[i] = e.(*object.Integer).Value
		// }
		// // Support setting arbitrary index with value for list
		// if listObj.Elements == nil {
		// 	listObj.Elements = []object.Object{}
		// }
		// for _, index := range indexes {
		// 	for index > int64(len(listObj.Elements)-1) {
		// 		listObj.Elements = append(listObj.Elements, object.NULL)
		// 	}
		// }
		// max := int64(len(listObj.Elements) - 1)
		// for _, index := range indexes {
		// 	if index < 0 || index > max {
		// 		return newError("index out of bounds: length=%d, index=%d", len(listObj.Elements), index)
		// 	}
		// }
		// newList := &object.List{Elements: make([]object.Object, len(indexes))}
		// for i, index := range indexes {
		// 	newList.Elements[i] = listObj.Elements[index]
		// }
		// return newList
	default:
		return vm.push(object.NULL)
	}
	// Support setting arbitrary index with value for list
	if listObj.Elements == nil {
		listObj.Elements = []object.Object{}
	}
	for idx > int64(len(listObj.Elements)-1) {
		listObj.Elements = append(listObj.Elements, object.NULL)
	}
	max := int64(len(listObj.Elements) - 1)
	if idx < 0 || idx > max {
		return vm.push(newError("index out of bounds: length=%d, index=%d", len(listObj.Elements), idx))
	}
	return vm.push(listObj.Elements[idx])
}

func (vm *VM) executeSetIndexExpression(set, indx object.Object) error {
	setObj := set.(*object.Set)
	var idx int64
	switch indx.Type() {
	case object.INTEGER_OBJ:
		idx = indx.(*object.Integer).Value
	case object.STRING_OBJ:
		// stringVal := indx.(*object.Stringo).Value
		// envVal, ok := e.env.Get(stringVal)
		// if !ok {
		// 	return object.NULL
		// }
		// intVal, ok := envVal.(*object.Integer)
		// if !ok {
		// 	return object.NULL
		// }
		// idx = intVal.Value
	case object.LIST_OBJ:
		// Handle range expressions (1..3) or (1..<3) => they come back as a list
		// indxList := indx.(*object.List).Elements
		// indexes := make([]int64, len(indxList))
		// for i, e := range indxList {
		// 	if e.Type() != object.INTEGER_OBJ {
		// 		return newError("index range needs to be INTEGER. got=%s", e.Type())
		// 	}
		// 	indexes[i] = e.(*object.Integer).Value
		// }
		// newSet := &object.Set{Elements: object.NewSetElements()}
		// for _, index := range indexes {
		// 	var j int64
		// 	for _, k := range setObj.Elements.Keys {
		// 		if v, ok := setObj.Elements.Get(k); ok {
		// 			if j == index {
		// 				newSet.Elements.Set(k, v)
		// 			}
		// 		}
		// 		j++
		// 	}
		// }
		// return newSet
	default:
		return vm.push(newError("set index: expected index to be INT, STRING, or LIST. got=%s", indx.Type()))
	}
	var i int64
	for _, k := range setObj.Elements.Keys {
		if v, ok := setObj.Elements.Get(k); ok {
			if i == idx {
				return vm.push(v.Value)
			}
		}
		i++
	}
	return vm.push(object.NULL)
}

func (vm *VM) executeMapIndexExpression(mapObject, indx object.Object) error {
	mapObj := mapObject.(*object.Map)
	if ok := object.IsHashable(indx); !ok {
		return vm.push(newError("unusable as a map key: %s", indx.Type()))
	}
	hashed := object.HashObject(indx)
	key := object.HashKey{Type: indx.Type(), Value: hashed}
	pair, ok := mapObj.Pairs.Get(key)
	if !ok {
		return vm.push(object.NULL)
	}
	return vm.push(pair.Value)
}

func (vm *VM) executeStringIndexExpression(str, indx object.Object) error {
	strObj := str.(*object.Stringo)
	var idx int64
	switch indx.Type() {
	case object.INTEGER_OBJ:
		idx = indx.(*object.Integer).Value
	case object.STRING_OBJ:
		// stringVal := indx.(*object.Stringo).Value
		// envVal, ok := e.env.Get(stringVal)
		// if !ok {
		// 	return object.NULL
		// }
		// intVal, ok := envVal.(*object.Integer)
		// if !ok {
		// 	return object.NULL
		// }
		// idx = intVal.Value
	case object.LIST_OBJ:
		// Handle range expressions (1..3) or (1..<3) => they come back as a list
		// indxList := indx.(*object.List).Elements
		// indexes := make([]int64, len(indxList))
		// for i, e := range indxList {
		// 	if e.Type() != object.INTEGER_OBJ {
		// 		return newError("index range needs to be INTEGER. got=%s", e.Type())
		// 	}
		// 	indexes[i] = e.(*object.Integer).Value
		// }
		// max := int64(runeLen(strObj.Value) - 1)
		// for _, index := range indexes {
		// 	if index < 0 || index > max {
		// 		return newError("index out of bounds: length=%d, index=%d", runeLen(strObj.Value), index)
		// 	}
		// }
		// newStr := make([]rune, len(indexes))
		// runeStr := []rune(strObj.Value)
		// for i, index := range indexes {
		// 	newStr[i] = runeStr[index]
		// }
		// return &object.Stringo{Value: string(newStr)}
	default:
		return vm.push(object.NULL)
	}
	max := int64(runeLen(strObj.Value) - 1)
	if idx < 0 || idx > max {
		return vm.push(newError("index out of bounds: length=%d, index=%d", runeLen(strObj.Value), idx))
	}
	return vm.push(&object.Stringo{Value: string([]rune(strObj.Value)[idx])})
}

func (vm *VM) executeProcessIndexExpression(process *object.Process, name string) error {
	switch name {
	case "id":
		return vm.push(&object.UInteger{Value: process.Id})
	case "name":
		return vm.push(&object.Stringo{Value: process.NodeName})
	case "send":
		p := process
		return vm.push(&object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newInvalidArgCountError("send", len(args), 1, "")
				}
				p.Ch <- args[0]
				return object.NULL
			},
			HelpStr: helpStrArgs{
				explanation: "`send` will take the given value and send it to the process",
				signature:   "send(pid: PROCESS, val: any) -> null",
				errors:      "InvalidArgCount,PositionalType",
				example:     "send(#{name: '', id: 1}, 'hello') => null",
			}.String(),
		})
	case "recv":
		p := process
		return vm.push(&object.Builtin{
			Fun: func(args ...object.Object) object.Object {
				if len(args) != 0 {
					return newInvalidArgCountError("recv", len(args), 0, "")
				}
				val := <-p.Ch
				if val == nil {
					return newError("`recv` error: process channel was closed")
				}
				return val
			},
			HelpStr: helpStrArgs{
				explanation: "`recv` waits for a value on the given process and returns it",
				signature:   "recv(pid: PROCESS) -> any",
				errors:      "InvalidArgCount,PositionalType,CustomError",
				example:     "recv(#{name: '', id: 1}) => 'something'",
			}.String(),
		})
	}
	panic("Unsupported Process Index Operation: " + name)
}
