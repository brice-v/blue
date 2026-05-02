package vm

import (
	"blue/blueutil"
	"blue/code"
	"blue/compiler"
	"blue/consts"
	"blue/lexer"
	"blue/object"
	"blue/parser"
	"blue/token"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/puzpuzpuz/xsync/v3"
)

const (
	StackSize   = 2048
	GlobalsSize = 65536
	MaxFrames   = 1024
)

type VM struct {
	constants   []object.Object
	tokens      []*token.Token
	lastNodePos int
	stack       []object.Object
	sp          int // Always points to the next value. Top of stack is stack[sp-1]

	globals []object.Object

	frames      []*Frame
	framesIndex int

	inTry      bool
	inCatch    bool
	catchError string

	TokensForErrorTrace []*token.Token

	// Process things
	NodeName string
	PID      uint64
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) incrementOpCallArgCount() bool {
	cf := vm.frames[vm.framesIndex-1]
	nextPos := blueutil.GetNextOpCallPos(cf.cl.Fun.Instructions, cf.ip)
	if nextPos != -1 {
		pos := nextPos + 1
		// When this function is called in a loop such as
		// split_lines.len(), then ensure we do not end up incrementing
		// the same byte position more then once.
		if _, ok := cf.cl.Fun.PosAlreadyIncremented.Load(pos); ok {
			return true
		}
		cf.cl.Fun.PosAlreadyIncremented.Store(pos, struct{}{})
		cf.cl.Fun.Instructions[pos]++
		return true
	}
	return false
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	if vm.framesIndex-1 == 0 {
		return vm.frames[vm.framesIndex]
	}
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func NewNode(nodeName string, bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions, PosAlreadyIncremented: xsync.NewMapOf[int, struct{}]()}
	mainClosure := &object.Closure{Fun: mainFn}
	mainFrame := NewFrame(mainClosure, 0)
	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame
	vm := &VM{
		constants:   bytecode.Constants,
		tokens:      bytecode.Tokens,
		lastNodePos: -1,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     make([]object.Object, GlobalsSize),
		frames:      frames,
		framesIndex: 1,
		inTry:       false,
		inCatch:     false,
		PID:         object.PidCount.Load(),
		NodeName:    nodeName,
	}
	// Create an empty process so we can recv without spawning
	process := &object.Process{
		// TODO: Eventually update to non-buffered and update send and recv as needed
		Ch: make(chan object.Object, 1),
		Id: vm.PID,

		NodeName: nodeName,
	}
	object.ProcessMap.LoadOrStore(object.Pk(nodeName, vm.PID), process)
	return vm
}

func New(bytecode *compiler.Bytecode) *VM {
	return NewNode("vm-node", bytecode)
}

func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VM) Clone(pid uint64) *VM {
	newFrames := make([]*Frame, len(vm.frames))
	for i, f := range vm.frames {
		newFrames[i] = f.Clone()
	}
	newVm := &VM{
		constants:   object.CloneSlice(vm.constants),
		stack:       object.CloneSlice(vm.stack),
		sp:          vm.sp,
		globals:     object.CloneSlice(vm.globals),
		frames:      newFrames,
		framesIndex: vm.framesIndex,
		NodeName:    vm.NodeName,
		PID:         pid,
	}
	return newVm
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) PushAndReturnError(err error) error {
	if vm.inTry {
		vm.push(newError("%s", err.Error()))
		return nil
	}
	return err
}

func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode
	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++
		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])
		if op == code.OpNode {
			vm.lastNodePos = vm.currentFrame().ip
			// Skip TokenIndex of OpNode
			vm.currentFrame().ip += 2
			continue
		}
		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpTrue:
			vm.push(object.TRUE)
		case code.OpFalse:
			vm.push(object.FALSE)
		case code.OpNull:
			vm.push(object.NULL)
		case code.OpPop:
			vm.pop()
		case code.OpAdd, code.OpMinus, code.OpStar, code.OpPow, code.OpDiv,
			code.OpFlDiv, code.OpPercent, code.OpCarat, code.OpAmpersand,
			code.OpPipe, code.OpIn, code.OpNotin, code.OpRange, code.OpNonIncRange,
			code.OpAnd, code.OpEqual, code.OpNotEqual, code.OpOr, code.OpGreaterThan, code.OpGreaterThanOrEqual,
			code.OpRshift:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpLshift:
			right := vm.pop()
			left := vm.peek()
			if left.Type() != object.LIST_OBJ && left.Type() != object.SET_OBJ {
				return vm.push(newError("unknown operator: %s << %s", left.Type(), right.Type()))
			}
			if left.Type() == object.LIST_OBJ {
				l := left.(*object.List)
				l.Elements = append(l.Elements, right)
			} else {
				s := left.(*object.Set)
				key := object.HashObject(right)
				if _, ok := s.Elements.Get(key); !ok {
					s.Elements.Set(key, object.SetPair{Value: right, Present: struct{}{}})
				}
			}
		case code.OpNot:
			err := vm.executeNotOperation()
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpNotIfNotNull:
			err := vm.executeNotIfNotNullOperation()
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpNeg:
			err := vm.executeNegOperation()
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpTilde:
			err := vm.executeBitwiseNotOperation()
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpLshiftPre:
			err := vm.executeLshiftPrefixOperation()
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpRshiftPost:
			err := vm.executeRshiftPostfixOperation()
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}
		case code.OpJumpNotTruthyAndPushTrue:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			// Pass null through (as this would mean we should re-push the item on the stack)
			if vm.peek() != object.NULL {
				condition := vm.pop()
				if !isTruthy(condition) {
					vm.currentFrame().ip = pos - 1
					vm.push(object.TRUE)
				} else {
					// Execute Not on the Condition to reverse OpNotIfNotNull which is always called prior
					err := vm.push(condition)
					if err != nil {
						return err
					}
					vm.executeNotOperation()
				}
			}
		case code.OpJumpNotTruthyAndPushFalse:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
				vm.push(object.FALSE)
			} else {
				err := vm.push(condition)
				if err != nil {
					return err
				}
			}
		case code.OpSetGlobal, code.OpSetGlobalImm:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[globalIndex] = vm.pop()
		case code.OpGetGlobal, code.OpGetGlobalImm:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpGetGlobalImmOrSpecial:
			globalIndex := code.ReadUint16(ins[ip+1:])
			processKeyIndex := code.ReadUint8(ins[ip+3:])
			vm.currentFrame().ip += 3
			var err error
			if processKeyIndex != 0 && vm.safePeek().Type() == object.PROCESS_OBJ && len(ins) >= vm.currentFrame().ip+1 && ins[vm.currentFrame().ip+1] == byte(code.OpIndex) {
				vm.currentFrame().ip += 1 // Skip over OpIndex, we are going to pop and then push the evaluated result back on the stack
				name := object.GetProcessKeyName(processKeyIndex)
				err = vm.executeProcessIndexExpression(vm.pop().(*object.Process), name)
			} else {
				err = vm.push(vm.globals[globalIndex])
			}
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpList:
			numElems := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			list := vm.buildList(vm.sp-numElems, vm.sp)
			vm.sp = vm.sp - numElems
			err := vm.push(list)
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpMap:
			numPairs := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			m := vm.buildMap(vm.sp-numPairs, vm.sp)
			vm.sp = vm.sp - numPairs
			err := vm.push(m)
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpSet:
			numElems := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			set := vm.buildSet(vm.sp-numElems, vm.sp)
			vm.sp = vm.sp - numElems
			err := vm.push(set)
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()
			if (index.Type() == object.CLOSURE || index.Type() == object.BUILTIN_OBJ) && vm.incrementOpCallArgCount() {
				err := vm.push(index)
				if err != nil {
					return err
				}
				err = vm.push(left)
				if err != nil {
					return err
				}
			} else {
				err := vm.executeIndexExpression(left, index)
				if err != nil {
					err = vm.PushAndReturnError(err)
					if err != nil {
						return err
					}
				}
			}
		case code.OpCall:
			numArgs := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			err := vm.executeCall(int(numArgs))
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpReturnValue:
			vm.currentFrame().cl.Fun.ClearSpecialFunctionParameters()
			returnValue := vm.pop()
			frame := vm.popFrame()
			if frame != nil {
				vm.sp = frame.bp - 1
			}
			err := vm.push(returnValue)
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
			if frame != nil {
				for deferFun := frame.PopDeferFun(); deferFun != nil; deferFun = frame.PopDeferFun() {
					err := vm.callClosure(deferFun, 0)
					if err != nil {
						err = vm.PushAndReturnError(err)
						if err != nil {
							return err
						}
					}
				}
			}
			if vm.currentFrame() == nil {
				// Break out if the current frame is actually empty
				// Should only ever happen with applyFunction
				return fmt.Errorf(consts.NORMAL_EXIT_ON_RETURN)
			}
		case code.OpReturn:
			vm.currentFrame().cl.Fun.ClearSpecialFunctionParameters()
			frame := vm.popFrame()
			if frame != nil {
				vm.sp = frame.bp - 1
			}
			vm.push(object.NULL)
			if frame != nil {
				for deferFun := frame.PopDeferFun(); deferFun != nil; deferFun = frame.PopDeferFun() {
					err := vm.callClosure(deferFun, 0)
					if err != nil {
						err = vm.PushAndReturnError(err)
						if err != nil {
							return err
						}
					}
				}
			}
			if vm.currentFrame() == nil {
				// Break out if the current frame is actually empty
				// Should only ever happen with applyFunction
				return fmt.Errorf(consts.NORMAL_EXIT_ON_RETURN)
			}
		case code.OpSetLocal, code.OpSetLocalImm:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			vm.stack[frame.bp+int(localIndex)] = vm.pop()
		case code.OpGetLocal, code.OpGetLocalImm:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			err := vm.push(vm.stack[frame.bp+int(localIndex)])
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpClosure:
			constIndex := code.ReadUint16(ins[ip+1:])
			numFree := code.ReadUint8(ins[ip+3:])
			vm.currentFrame().ip += 3
			err := vm.pushClosure(int(constIndex), int(numFree))
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpGetFree, code.OpGetFreeImm:
			freeIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			currentClosure := vm.currentFrame().cl
			err := vm.push(currentClosure.Free[freeIndex])
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpGetBuiltin:
			builtinModuleIndex := code.ReadUint8(ins[ip+1:])
			builtinIndex := code.ReadUint8(ins[ip+2:])
			vm.currentFrame().ip += 2
			var err error
			if builtinModuleIndex == object.BuiltinobjsModuleIndex {
				definition := object.BuiltinobjsList[builtinIndex]
				obj := definition.Builtin.Obj
				if definition.Name == "ENV" {
					obj.(*object.Map).IsEnvBuiltin = true
				}
				err = vm.push(obj)
			} else {
				definition := object.AllBuiltins[builtinModuleIndex].Builtins[builtinIndex]
				var builtin *object.Builtin
				if definition.Fun == nil {
					if builtinModuleIndex != 0 {
						modName := object.GetNameOfModuleByIndex(int(builtinModuleIndex))
						builtin = &object.Builtin{
							Name:    definition.Name,
							Fun:     GetStdBuiltinWithVm(modName, definition.Name, vm),
							HelpStr: definition.HelpStr,
							Mutates: definition.Mutates,
						}
					} else {
						if blueutil.ENABLE_VM_CACHING {
							// Lazy Evaluate Builtin that needs to use vm
							definition.Fun = GetBuiltinWithVm(definition.Name, vm)
							builtin = definition
						} else {
							builtin = &object.Builtin{
								Name:    definition.Name,
								Fun:     GetBuiltinWithVm(definition.Name, vm),
								HelpStr: definition.HelpStr,
								Mutates: definition.Mutates,
							}
						}
					}
				} else {
					builtin = definition
				}
				err = vm.push(builtin)
			}
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpStringInterp:
			stringIndex := code.ReadUint16(ins[ip+1:])
			numPairs := int(code.ReadUint8(ins[ip+3:]))
			vm.currentFrame().ip += 3
			s := vm.buildStringWithInterp(vm.sp-numPairs, vm.sp, int(stringIndex))
			vm.sp = vm.sp - numPairs - 1
			err := vm.push(s)
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpIndexSet:
			index := vm.pop()
			left := vm.pop()
			right := vm.pop()
			err := vm.executeIndexSetOperator(left, index, right)
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			} else {
				vm.push(object.NULL)
			}
		case code.OpTry:
			vm.inTry = true
		case code.OpCatch:
			if vm.catchError == "" {
				vm.gotoCatchEnd()
			} else {
				vm.inCatch = true
			}
		case code.OpFinallyEnd:
			if vm.catchError != "" {
				return vm.prepareStackTraceAndReturnError(fmt.Errorf("%s", vm.catchError))
			}
			vm.inTry = false
			vm.inCatch = false
		case code.OpCatchEnd:
			// If we were in catch, set catch error back to empty
			vm.catchError = ""
			vm.inTry = false
			vm.inCatch = false
		case code.OpFinally, code.OpListCompLiteral, code.OpSetCompLiteral, code.OpMapCompLiteral:
			// Do nothing
		case code.OpExecString:
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			execStr := vm.constants[constIndex]
			str, ok := execStr.(*object.ExecString)
			if !ok {
				return fmt.Errorf("expected ExecString, got=%T", execStr)
			}
			err := vm.push(object.ExecStringCommand(str.Value))
			if err != nil {
				return err
			}
		case code.OpEval:
			strToEval := vm.pop()
			err := vm.executeEvalOperation(strToEval)
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpMatchValue:
			right := vm.pop()
			left := vm.pop()
			vm.push(nativeToBooleanObject(matches(left, right)))
		case code.OpMatchAny:
			vm.push(object.VM_IGNORE)
		case code.OpDefaultArgs:
			numPairs := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			m := vm.buildDefaultArgs(vm.sp-numPairs, vm.sp)
			vm.sp = vm.sp - numPairs
			err := vm.push(m)
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpSlice:
			sliceIndexes := vm.pop()
			left := vm.pop()
			err := vm.push(vm.buildSliceFrom(left, sliceIndexes))
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpStruct:
			numFields := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			// -1 to account for the fields held in go object
			startIndex := vm.sp - numFields - 1
			bs, err := vm.buildStruct(startIndex, vm.sp)
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
			vm.sp = startIndex
			err = vm.push(bs)
			if err != nil {
				return err
			}
		case code.OpGetListIndex:
			index := int(code.ReadUint8(ins[ip+1:]))
			vm.currentFrame().ip += 1
			list := vm.peek() // Dont pop it off the stack
			if list.Type() != object.LIST_OBJ {
				return vm.PushAndReturnError(fmt.Errorf("OpGetListIndex did not find List on top of the stack. got=%T", list))
			}
			l := list.(*object.List).Elements
			if index > len(l) {
				return vm.PushAndReturnError(fmt.Errorf("OpGetListIndex index is greater than the length of the list. index=%d, len(list)=%d", index, len(l)))
			}
			err := vm.push(l[index])
			if err != nil {
				return err
			}
		case code.OpGetMapKey:
			key := vm.peek()
			if key.Type() != object.STRING_OBJ {
				return vm.PushAndReturnError(fmt.Errorf("OpGetMapKey did not find string key on top of the stack. got=%T", key))
			}
			m := vm.peekOffset(1)
			if m.Type() != object.MAP_OBJ {
				return vm.PushAndReturnError(fmt.Errorf("OpGetMapKey did not find map on top-1 of the stack. got=%T", m))
			}
			pair, ok := m.(*object.Map).Pairs.Get(object.HashKey{Type: object.STRING_OBJ, Value: object.HashObject(key)})
			if !ok {
				return vm.PushAndReturnError(fmt.Errorf("OpGetMapKey did not find value for name: `%s` in the map", key.(*object.Stringo).Value))
			}
			err := vm.push(pair.Value)
			if err != nil {
				return err
			}
		case code.OpDefer:
			numDefers := int(code.ReadUint8(ins[ip+1:]))
			for range numDefers {
				deferFun := vm.pop()
				if deferFun.Type() != object.CLOSURE {
					return vm.PushAndReturnError(fmt.Errorf("OpDefer did not find function on top of the stack. got=%T", deferFun))
				}
				vm.currentFrame().PushDeferFun(deferFun.(*object.Closure))
			}
		case code.OpSelf:
			var err error
			if p, ok := object.ProcessMap.Load(object.Pk(vm.NodeName, vm.PID)); ok {
				err = vm.push(p)
			} else {
				err = vm.push(newError("`self` error: process not found"))
			}
			if err != nil {
				return err
			}
		case code.OpSpawn:
			numArgs := int(code.ReadUint8(ins[ip+1:]))
			vm.currentFrame().ip += 1
			var args []object.Object
			funArgIndex := -1
			listArgIndex := -1
			if numArgs == 0 {
				args = []object.Object{vm.pop()}
				funArgIndex = 0
			} else {
				args = []object.Object{vm.pop(), vm.pop()}
				funArgIndex = 1
				listArgIndex = 0
			}
			vm.executeSpawn(args, funArgIndex, listArgIndex)
		case code.OpGetFunctionParameterSpecial:
			indexParam := int(code.ReadUint8(ins[ip+1:]))
			indexList := int(code.ReadUint8(ins[ip+2:]))
			vm.currentFrame().ip += 2
			err := vm.pushSpecialFunctionParameter(indexParam, indexList)
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpGetFunctionParameterSpecial2:
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			str, ok := vm.constants[constIndex].(*object.Stringo)
			if !ok {
				return vm.prepareStackTraceAndReturnError(fmt.Errorf("found non-string in constant for OpGetFunctionParameterSpecial2, got = %T", vm.constants[constIndex]))
			}
			err := vm.pushSpecialFunctionParameter2(str.Value)
			if err != nil {
				err = vm.PushAndReturnError(err)
				if err != nil {
					return err
				}
			}
		case code.OpSpecialIndexHelper:
			maybeJumpPos := int(code.ReadUint16(ins[ip+1:]))
			isSet := code.Opcode(ins[maybeJumpPos]) == code.OpIndexSet
			vm.currentFrame().ip += 2
			// Should be the string constant of the 'indx'
			indxStr := vm.peek().(*object.Stringo)
			// Should be the the object we are trying to figure out if its being pushed
			// to a function or if its a map and we are just indexing by this
			obj1 := vm.peekOffset(1)
			if obj1.Type() != object.MAP_OBJ {
				// If its not a map, just assume we are trying to push this to a function
				// pop the index string off the stack so it doesnt interfere
				vm.pop()
			} else {
				// If it is a map, need to determine if we need to index the map, or just pass to next function
				if obj1.(*object.Map).ContainsStringoKey(indxStr) || isSet {
					// Leave the string on the stack to be used for indexing and skip loading the global
					// If its an index set operation then we always want to use the string for setting
					// when it was done via a dot call
					vm.currentFrame().ip = maybeJumpPos - 1
				} else {
					// Otherwise pop off the stack and assume we are passing this map to a global function/index operation
					vm.pop()
				}
			}
		case code.OpNotInTry:
			vm.inTry = false
		case code.OpNotInCatch:
			vm.inCatch = false
		}
		if ip != 0 {
			vm.currentFrame().lastInstruction = op
		}
	}
	return nil
}

func (vm *VM) printMiniStack(slots int) {
	for i := range slots {
		obj := vm.stack[i]
		if obj != nil {
			s := obj.Inspect()
			if runeLen(s) > 10000 {
				s = s[:100] + "..."
			}
			log.Printf("stack[%d] = %q (%T)\n", i, s, obj)
		}
	}
}

func (vm *VM) printStackInfo() {
	normalItems := 0
	nilItems := 0
	typeMapCount := map[string]int{}
	for _, e := range vm.stack {
		if e != nil {
			normalItems++
			t := fmt.Sprintf("%T", e)
			_, ok := typeMapCount[t]
			if !ok {
				typeMapCount[t] = 1
			} else {
				typeMapCount[t]++
			}
		} else {
			nilItems++
		}
	}
	log.Printf("normalItems = %d, nilItems = %d", normalItems, nilItems)
	log.Printf("typeMapCount --------------------------")
	for k, v := range typeMapCount {
		log.Printf("%s %d", k, v)
	}
	log.Printf("------------ --------------------------")
}

func (vm *VM) prepareStackTraceAndReturnError(err error) error {
	vm.TokensForErrorTrace = []*token.Token{}
	ip := vm.lastNodePos
	// vm.printStackInfo()
	for vm.framesIndex >= 1 {
		ins := vm.currentFrame().Instructions()
		if ip > len(ins) {
			// Prevent panic when preparing stack trace
			// Note: I saw this occur when dealing with stack overflow
			break
		}
		op := code.Opcode(ins[ip])
		if op != code.OpNode {
			break
		}
		tokenPos := code.ReadUint16(ins[ip+1:])
		if int(tokenPos) > len(vm.tokens) {
			break
		}
		vm.TokensForErrorTrace = append(vm.TokensForErrorTrace, vm.tokens[tokenPos])
		if vm.framesIndex == 1 {
			// allow capture of call in main closure then exit (popFrame will return the same main)
			break
		}
		vm.popFrame()
		ip = vm.currentFrame().ip - 4 // Move back to OpNode of calling function
	}
	return err
}

func (vm *VM) push(o object.Object) error {
	if isError(o) {
		if vm.inTry || vm.inCatch {
			vm.gotoNextCatchOrFinally(o.(*object.Error).Message)
			return nil
		}
		return vm.prepareStackTraceAndReturnError(fmt.Errorf("%s", o.(*object.Error).Message))
	}
	if vm.sp >= StackSize {
		return vm.prepareStackTraceAndReturnError(fmt.Errorf("stack overflow when trying to push %+#v (%T)", o, o))
	}
	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pushNoErrorChecking(o object.Object) error {
	if vm.sp >= StackSize {
		return vm.prepareStackTraceAndReturnError(fmt.Errorf("stack overflow when trying to push %+#v (%T)", o, o))
	}
	vm.stack[vm.sp] = o
	vm.sp++
	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) safePeek() object.Object {
	if vm.sp == 0 {
		return object.NULL
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) peek() object.Object {
	return vm.stack[vm.sp-1]
}

func (vm *VM) peekOffset(offset int) object.Object {
	return vm.stack[vm.sp-1-offset]
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

func (vm *VM) gotoNextCatchOrFinally(errorMessage string) {
	vm.inTry = false
	wasInCatch := vm.inCatch && !vm.inTry
	vm.inCatch = false
	frameIndex := vm.framesIndex - 1
	for frameIndex >= 0 {
		frame := vm.frames[frameIndex]
		if newip, ok := vm.isOpCatchOrFinallyFoundInFrame(frame, errorMessage); ok {
			vm.framesIndex = frameIndex + 1
			vm.currentFrame().ip = newip
			break
		}
		if frameIndex-1 < 0 {
			// TODO: Error out here?
			break
		}
		frameIndex--
	}
	if wasInCatch {
		vm.pushNoErrorChecking(newError("%s", errorMessage))
	}
}

func (vm *VM) isOpCatchOrFinallyFoundInFrame(frame *Frame, errorMessage string) (int, bool) {
	if frame == nil {
		return -1, false
	}
	ins := frame.Instructions()
	for i := frame.ip - 1; i < len(ins); i++ {
		def, err := code.Lookup(ins[i])
		if err != nil {
			continue
		}
		_, read := code.ReadOperands(def, ins[i+1:])
		switch def.Name {
		case "OpCatch":
			vm.catchError = errorMessage
			vm.push(&object.Stringo{Value: errorMessage})
			return i - 1, true
		case "OpFinally":
			vm.catchError = errorMessage
			return i - 1, true
		}
		i += read
	}
	return -1, false
}

func (vm *VM) gotoCatchEnd() {
	frame := vm.currentFrame()
	ins := frame.Instructions()
	for i := frame.ip - 1; i < len(ins); i++ {
		def, err := code.Lookup(ins[i])
		if err != nil {
			continue
		}
		_, read := code.ReadOperands(def, ins[i+1:])
		if def.Name == "OpCatchEnd" {
			vm.currentFrame().ip = i
			break
		}
		i += read
	}
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()
	leftType := left.Type()
	rightType := right.Type()
	if leftType == rightType {
		if binFun, ok := binaryOperationFunctions[leftType]; ok {
			return binFun(vm, op, left, right)
		}
		return vm.executeDefaultBinaryOperation(op, left, right)
	} else if leftType != rightType {
		return vm.executeBinaryOperationDifferentTypes(op, left, right, leftType, rightType)
	}
	return nil
}

func (vm *VM) executeNotOperation() error {
	operand := vm.pop()
	isTrue := isTruthy(operand)
	if isTrue {
		return vm.push(object.FALSE)
	} else {
		return vm.push(object.TRUE)
	}
}

func (vm *VM) executeNotIfNotNullOperation() error {
	if vm.peek() == object.NULL {
		return nil
	}
	return vm.executeNotOperation()
}

func (vm *VM) executeNegOperation() error {
	operand := vm.pop()
	if operand.Type() == object.INTEGER_OBJ {
		value := operand.(*object.Integer).Value
		return vm.push(&object.Integer{Value: -value})
	}
	if operand.Type() == object.FLOAT_OBJ {
		value := operand.(*object.Float).Value
		return vm.push(&object.Float{Value: -value})
	}
	if operand.Type() == object.BIG_INTEGER_OBJ {
		value := operand.(*object.BigInteger).Value
		return vm.push(&object.BigInteger{Value: new(big.Int).Neg(value)})
	}
	if operand.Type() == object.BIG_FLOAT_OBJ {
		value := operand.(*object.BigFloat).Value
		return vm.push(&object.BigFloat{Value: value.Neg()})
	}

	return vm.push(newError("unknown operator: -%s", operand.Type()))
}

func (vm *VM) executeBitwiseNotOperation() error {
	operand := vm.pop()
	switch operand.Type() {
	case object.INTEGER_OBJ:
		value := operand.(*object.Integer).Value
		return vm.push(&object.Integer{Value: ^value})
	case object.UINTEGER_OBJ:
		value := operand.(*object.UInteger).Value
		return vm.push(&object.UInteger{Value: ^value})
	case object.BYTES_OBJ:
		value := operand.(*object.Bytes).Value
		buf := make([]byte, len(value))
		for i, b := range value {
			buf[i] = ^b
		}
		return vm.push(&object.Bytes{Value: buf})
	default:
		return vm.push(newError("unknown operator: ~%s", operand.Type()))
	}
}

func (vm *VM) executeLshiftPrefixOperation() error {
	operand := vm.pop()
	switch operand.Type() {
	case object.LIST_OBJ:
		l := operand.(*object.List)
		listLen := len(l.Elements)
		if listLen == 0 {
			return vm.push(object.NULL)
		}
		e := l.Elements[0]
		if listLen == 1 {
			l.Elements = []object.Object{}
		} else {
			l.Elements = l.Elements[1:listLen]
		}
		return vm.push(e)
	default:
		return vm.push(newError("unknown operator: << %s", operand.Type()))
	}
}

func (vm *VM) executeRshiftPostfixOperation() error {
	operand := vm.pop()
	switch operand.Type() {
	case object.LIST_OBJ:
		l := operand.(*object.List)
		listLen := len(l.Elements)
		if listLen == 0 {
			return vm.push(object.NULL)
		}
		e := l.Elements[listLen-1]
		if listLen == 1 {
			l.Elements = []object.Object{}
		} else {
			l.Elements = l.Elements[0 : listLen-1]
		}
		return vm.push(e)
	default:
		return vm.push(newError("unknown operator: %s >>", operand.Type()))
	}
}

func (vm *VM) buildList(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)
	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}
	return &object.List{Elements: elements}
}

func (vm *VM) buildSet(startIndex, endIndex int) object.Object {
	setMap := object.NewSetElementsWithSize(endIndex - startIndex)
	for i := startIndex; i < endIndex; i++ {
		elem := vm.stack[i]
		hashKey := object.HashObject(elem)
		setMap.Set(hashKey, object.SetPair{Value: elem, Present: struct{}{}})
	}
	return &object.Set{Elements: setMap}
}

func (vm *VM) buildStruct(startIndex, endIndex int) (object.Object, error) {
	index := startIndex
	maybeFields := vm.stack[index]
	fields, ok := maybeFields.(*object.GoObj[[]string])
	if !ok {
		return nil, fmt.Errorf("compilation error: struct did not have fields in index: %d", index)
	}
	index++
	bs, err := object.NewBlueStruct(fields.Value, object.CloneSlice(vm.stack[index:endIndex]))
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (vm *VM) buildMap(startIndex, endIndex int) object.Object {
	pairs := object.NewPairsMapWithSize((endIndex - startIndex) / 2)
	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]
		if isError(key) {
			return key
		} else if isError(value) {
			return value
		}
		// TODO: Missing some logic here from evaluator
		if ok := object.IsHashable(key); !ok {
			return newError("unusable as a map key: %s", key.Type())
		}
		hk := object.HashObject(key)
		hashed := object.HashKey{Type: key.Type(), Value: hk}

		pairs.Set(hashed, object.MapPair{Key: key, Value: value})
	}
	return &object.Map{Pairs: pairs}
}

func (vm *VM) buildDefaultArgs(startIndex, endIndex int) object.Object {
	m := make(map[string]object.Object)
	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]
		if isError(key) {
			return key
		} else if isError(value) {
			return value
		}
		m[key.(*object.Stringo).Value] = value
	}
	return &object.DefaultArgs{Value: m}
}

func (vm *VM) buildStringWithInterp(startIndex, endIndex, stringIndex int) object.Object {
	str := vm.constants[stringIndex]
	s, ok := str.(*object.Stringo)
	if !ok {
		return newError("string interpolation failed with non-string `%s`", str.Inspect())
	}
	newStr := s.Value
	for i := startIndex; i < endIndex; i += 2 {
		exp := vm.stack[i]
		orig := vm.stack[i+1]
		newStr = strings.Replace(newStr, orig.Inspect(), exp.Inspect(), 1)
	}
	return &object.Stringo{Value: newStr}
}

func (vm *VM) executeIndexExpression(left, indx object.Object) error {
	switch {
	case left.Type() == object.LIST_OBJ:
		return vm.executeListIndexExpression(left, indx)
	case left.Type() == object.SET_OBJ:
		return vm.executeSetIndexExpression(left, indx)
	case left.Type() == object.MAP_OBJ:
		return vm.executeMapIndexExpression(left, indx)
	case left.Type() == object.STRING_OBJ:
		return vm.executeStringIndexExpression(left, indx)
	// case left.Type() == object.MODULE_OBJ:
	// 	return e.evalModuleIndexExpression(left, indx)
	case left.Type() == object.PROCESS_OBJ && indx.Type() == object.STRING_OBJ:
		return vm.executeProcessIndexExpression(left.(*object.Process), indx.(*object.Stringo).Value)
	case left.Type() == object.GO_OBJ && indx.Type() == object.STRING_OBJ:
		return vm.executeGoObjIndexExpression(left, indx.(*object.Stringo).Value)
	// case left.Type() == object.BLUE_STRUCT_OBJ && indx.Type() == object.STRING_OBJ:
	// 	return e.evalBlueStructIndexExpression(left, indx)
	default:
		return vm.push(newError("index operator not supported: %s.%s (%s.%s)", left.Type(), indx.Type(), left.Inspect(), indx.Inspect()))
	}
}

// Set Frames[0] and Frames[1] to the same frame from callClosureFastFrame
// so that when calling Run on blue function via applyFunctionFast we can
// return the top of stack value internally and revert state
func (vm *VM) executeCallFastFrame(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	if closure, ok := callee.(*object.Closure); ok {
		return vm.callClosureFastFrame(closure, numArgs)
	}
	return nil
}

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *object.Closure:
		return vm.callClosure(callee, numArgs)
	case *object.Builtin:
		return vm.callBuiltin(callee, numArgs)
	default:
		return vm.prepareStackTraceAndReturnError(fmt.Errorf("calling non-closure and non-builtin %T", callee))
	}
}

func (vm *VM) callClosureFastFrame(cl *object.Closure, numArgs int) error {
	newArgCount, err := vm.interweaveArgsForCall(cl.Fun, numArgs)
	if err != nil {
		return err
	}

	frame := NewFrame(cl, vm.sp-newArgCount)
	// CurrentFrame looks at frameIndex-1
	vm.frames[vm.framesIndex-1] = frame
	vm.sp = frame.bp + cl.Fun.NumLocals
	return nil
}

func (vm *VM) callClosure(cl *object.Closure, numArgs int) error {
	newArgCount, err := vm.interweaveArgsForCall(cl.Fun, numArgs)
	if err != nil {
		return err
	}

	frame := NewFrame(cl, vm.sp-newArgCount)
	vm.pushFrame(frame)
	vm.sp = frame.bp + cl.Fun.NumLocals
	return nil
}

func (vm *VM) interweaveArgsForCall(cl *object.CompiledFunction, numArgs int) (int, error) {
	currentArgsOnStack := vm.stack[vm.sp-numArgs : vm.sp]
	vm.sp -= numArgs
	if len(currentArgsOnStack) > cl.NumParameters {
		return 0, fmt.Errorf("wrong number of arguments: want=%d, got=%d", cl.NumParameters, numArgs)
	}
	realNumArg := numArgs
	var defaultArgs map[string]object.Object = nil
	if len(currentArgsOnStack) != 0 {
		lastArg := currentArgsOnStack[len(currentArgsOnStack)-1]
		if da, ok := lastArg.(*object.DefaultArgs); ok {
			defaultArgs = da.Value
			realNumArg--
			currentArgsOnStack = currentArgsOnStack[:len(currentArgsOnStack)-1]
		}
	}
	currentArgOnStackIndex := 0
	args := make([]object.Object, 0, cl.NumParameters)
	potentialArgCount := realNumArg + cl.NumDefaultParams
	potentialArgToSwap := struct {
		paramIndex, argOnStackIndex int
	}{
		paramIndex:      -1,
		argOnStackIndex: -1,
	}
	for i := range cl.NumParameters {
		if realNumArg == cl.NumParameters {
			args = append(args, currentArgsOnStack[currentArgOnStackIndex])
			currentArgOnStackIndex++
			continue
		} else if defaultArgs != nil {
			if defaultArg, ok := defaultArgs[cl.Parameters[i]]; ok {
				args = append(args, defaultArg)
				continue
			}
		}
		useParamCond := (potentialArgCount == cl.NumParameters && cl.ParameterHasDefault[i]) ||
			(potentialArgCount > cl.NumParameters && currentArgOnStackIndex == len(currentArgsOnStack) && cl.ParameterHasDefault[i])
		if useParamCond {
			args = append(args, object.USE_PARAM_STR_OBJ)
		} else if potentialArgCount == cl.NumParameters && currentArgOnStackIndex < len(currentArgsOnStack) {
			args = append(args, currentArgsOnStack[currentArgOnStackIndex])
			currentArgOnStackIndex++
		} else if potentialArgCount > cl.NumParameters && currentArgOnStackIndex < len(currentArgsOnStack) {
			args = append(args, currentArgsOnStack[currentArgOnStackIndex])
			currentArgOnStackIndex++
			if cl.ParameterHasDefault[i] {
				potentialArgToSwap.argOnStackIndex = currentArgOnStackIndex - 1
				potentialArgToSwap.paramIndex = i
			}
		} else if realNumArg < cl.NumDefaultParams && potentialArgCount > cl.NumParameters && defaultArgs == nil && !cl.ParameterHasDefault[i] && potentialArgToSwap.argOnStackIndex != -1 {
			// Append the arg from the prior position that had a default parameter (and this one doesnt)
			args = append(args, currentArgsOnStack[potentialArgToSwap.argOnStackIndex])
			// Replace the prior position with the USE_PARAM object so that it will be used
			args[potentialArgToSwap.argOnStackIndex] = object.USE_PARAM_STR_OBJ
		}
	}
	argCountAfterWeave := len(args)
	if argCountAfterWeave != cl.NumParameters {
		return 0, fmt.Errorf("wrong number of arguments: want=%d, got=%d", cl.NumParameters, numArgs)
	}
	for _, arg := range args {
		vm.push(arg)
	}
	return argCountAfterWeave, nil
}

func (vm *VM) pushClosure(constIndex, numFree int) error {
	constant := vm.constants[constIndex]
	function, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}
	free := make([]object.Object, numFree)
	for i := range numFree {
		free[i] = vm.stack[vm.sp-numFree+i]
	}
	vm.sp = vm.sp - numFree
	closure := &object.Closure{Fun: function, Free: free}
	return vm.push(closure)
}

func (vm *VM) callBuiltin(builtin *object.Builtin, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]
	result := builtin.Fun(args...)
	vm.sp = vm.sp - numArgs - 1
	return vm.push(result)
}

func vmStr(s string) object.Object {
	l := lexer.New(s, "<internal: string>")
	p := parser.New(l)
	prog := p.ParseProgram()
	pErrors := p.Errors()
	if len(pErrors) != 0 {
		return newError("failed to `eval` string, found '%d' parser errors", len(pErrors))
	}
	c := compiler.New()
	err := c.Compile(prog)
	if err != nil {
		return newError("compiler error in `eval` string: %s", err.Error())
	}
	vm := New(c.Bytecode())
	err = vm.Run()
	if err != nil {
		return newError("vm error in `eval` string: %s", err.Error())
	}
	return vm.LastPoppedStackElem()
}

func (vm *VM) executeSpawn(args []object.Object, funArgIndex, listArgIndex int) {
	if args[funArgIndex].Type() != object.CLOSURE {
		vm.push(newError("`spawn` error: expected function got = %s", args[funArgIndex].Type()))
		return
	}
	if args[listArgIndex].Type() != object.LIST_OBJ {
		vm.push(newError("`spawn` error: expected list got = %s", args[listArgIndex].Type()))
		return
	}
	fun := args[funArgIndex].(*object.Closure)
	pid := object.PidCount.Add(1)
	process := &object.Process{
		Id: pid,
		// TODO: Eventually update to non-buffered and update send and recv as needed
		Ch: make(chan object.Object, 1),

		NodeName: vm.NodeName,
	}
	object.ProcessMap.Store(object.Pk(vm.NodeName, pid), process)
	clonedVm := vm.Clone(pid)
	// Dont clone args list so if processes are sent through then they will be usable by the process (channel must not be "cloned")
	go spawnFunction(clonedVm, vm.NodeName, fun.Clone().(*object.Closure), args[listArgIndex].(*object.List))
	vm.push(process)
}

func spawnFunction(vm *VM, nodeName string, fun *object.Closure, arg1 object.Object) {
	elems := arg1.(*object.List).Elements
	newObj := vm.applyFunctionFastWithMultipleArgs(fun, elems)
	if isError(newObj) {
		// err := newObj.(*object.Error)
		// var buf bytes.Buffer
		// buf.WriteString(err.Message)
		// buf.WriteByte('\n')
		// for newE.ErrorTokens.Len() > 0 {
		// 	tok := newE.ErrorTokens.PopBack()
		// 	fmt.Fprintf(&buf, "%s\n", lexer.GetErrorLineMessage(tok))
		// }
		fmt.Printf("%s%s\n", consts.PROCESS_ERROR_PREFIX, newObj.(*object.Error).Message)
	}
	// Delete from concurrent map and close channel (not 100% sure its necessary)
	if process, ok := object.ProcessMap.LoadAndDelete(object.Pk(nodeName, vm.PID)); ok {
		close(process.Ch)
	}
}
