package evaluator

import (
	"blue/ast"
	"blue/object"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

// for now everything that is not null or false returns true
// TODO: Update this list to include non truthy for empty objects, lists, etc.
func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

// newError is the wrapper function to add an error to the evaluator
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// isError is the helper function to determine if an object is an error
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func execCommand(arg0 string, args ...string) *exec.Cmd {
	if args == nil {
		if runtime.GOOS == "windows" {
			winArgs := []string{"/c"}
			winArgs = append(winArgs, arg0)
			return exec.Command("cmd", winArgs...)
		}
		return exec.Command(arg0)
	}
	if runtime.GOOS == "windows" {
		winArgs := []string{"/c"}
		winArgs = append(winArgs, arg0)
		winArgs = append(winArgs, args...)
		return exec.Command("cmd", winArgs...)
	}
	return exec.Command(arg0, args...)
}

func twoListsEqual(leftList, rightList *object.List) bool {
	// This is a deep equality expensive function
	return object.HashObject(leftList) == object.HashObject(rightList)
}

func nativeToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

// ArgToPassToBuiltin is the argument to be given to the builtin function
var ArgToPassToBuiltin object.Object = nil

func tryCreateValidBuiltinForDotCall(left, indx object.Object, leftNode ast.Expression) object.Object {
	// Try to see if the index being used is a builtin function
	if indx.Type() != object.STRING_OBJ {
		return nil
	}
	_, isBuiltin := builtins[indx.Inspect()]
	_, isStringBuiltin := stringbuiltins[indx.Inspect()]
	if !isBuiltin && !isStringBuiltin {
		return nil
	}
	// Allow either a string object or identifier to be passed to the builtin
	_, ok1 := left.(*object.Stringo)
	_, ok2 := leftNode.(*ast.Identifier)
	if !ok1 && !ok2 {
		return nil
	}

	ArgToPassToBuiltin = left
	// Return the builtin function object so that it can be used in the call
	// expression
	if isBuiltin {
		return &object.Builtin{
			Fun: builtins[indx.Inspect()].Fun,
		}
	}
	return &object.Builtin{
		Fun: stringbuiltins[indx.Inspect()].Fun,
	}
}

func applyFunction(fun object.Object, args []object.Object, defaultArgs map[string]object.Object) object.Object {
	switch function := fun.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(function, args, defaultArgs)
		evaluated := Eval(function.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		if ArgToPassToBuiltin != nil {
			// prepend the argument to pass in to the front
			args = append([]object.Object{ArgToPassToBuiltin}, args...)
			// Unset the argument to pass in so itll be free next time we come to it
			ArgToPassToBuiltin = nil
		}
		return function.Fun(args...)
	default:
		return newError("not a function %s", function.Type())
	}
}

func extendFunctionEnv(fun *object.Function, args []object.Object, defaultArgs map[string]object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fun.Env)

	// If the arguments slice is the same length as the parameter list, then we have them all
	// so set them as normal
	if len(args) == len(fun.Parameters) {
		for paramIndx, param := range fun.Parameters {
			env.Set(param.Value, args[paramIndx])
		}
		setDefaultCallExpressionParameters(defaultArgs, env)
	} else if len(args) < len(fun.Parameters) {
		// loop and while less than the total parameters set environment variables accordingly
		argsIndx := 0
		for paramIndx, param := range fun.Parameters {
			if fun.DefaultParameters[paramIndx] == nil {
				if argsIndx < len(args) {
					env.Set(param.Value, args[argsIndx])
					argsIndx++
					continue
				}
			}
			env.Set(param.Value, fun.DefaultParameters[paramIndx])
		}
		setDefaultCallExpressionParameters(defaultArgs, env)
	}
	return env
}

func setDefaultCallExpressionParameters(defaultArgs map[string]object.Object, env *object.Environment) {
	for k, v := range defaultArgs {
		_, ok := env.Get(k)
		if ok {
			env.Set(k, v)
		}
	}
}

func createFilePathFromImportPath(importPath string) string {
	var fpath bytes.Buffer
	if EvalBasePath != "." {
		fpath.WriteString(EvalBasePath)
		fpath.WriteString(string(os.PathSeparator))
	}
	importPath = strings.ReplaceAll(importPath, ".", string(os.PathSeparator))
	fpath.WriteString(importPath)
	fpath.WriteString(".b")
	return fpath.String()
}

func doCondAndMatchExpEqual(condVal, matchVal object.Object) bool {
	condValPairs := condVal.(*object.Map).Pairs
	matchValPairs := matchVal.(*object.Map).Pairs
	condValLen := len(condValPairs)
	matchValLen := len(matchValPairs)
	if condValLen != matchValLen {
		return false
	}
	for condKey, condValue := range condValPairs {
		_, ok := matchValPairs[condKey]
		if !ok {
			return false
		}
		if condValue.Value == IGNORE {
			continue
		}
		val, ok := matchValPairs[condKey]
		if !ok {
			return false
		}
		if object.HashObject(val.Value) != object.HashObject(condValue.Value) {
			return false
		}
	}

	return true
}
