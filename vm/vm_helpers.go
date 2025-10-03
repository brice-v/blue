package vm

import (
	"blue/object"
	"fmt"
	"unicode/utf8"
)

func executeIntegerRangeOperator(leftVal, rightVal int64) object.Object {
	var i int64

	if leftVal < rightVal {
		size := rightVal - leftVal
		listElems := make([]object.Object, 0, size)
		for i = leftVal; i <= rightVal; i++ {
			listElems = append(listElems, &object.Integer{Value: i})
		}
		return &object.List{Elements: listElems}
	} else if rightVal < leftVal {
		size := leftVal - rightVal
		listElems := make([]object.Object, 0, size)
		for i = leftVal; i >= rightVal; i-- {
			listElems = append(listElems, &object.Integer{Value: i})
		}
		return &object.List{Elements: listElems}
	}
	// When they are equal just return a value (leftVal in this case)
	return &object.List{Elements: []object.Object{&object.Integer{Value: leftVal}}}
}

func executeIntegerNonInclusiveRangeOperator(leftVal, rightVal int64) object.Object {
	var i int64

	if leftVal < rightVal {
		size := rightVal - leftVal
		listElems := make([]object.Object, 0, size-1)
		for i = leftVal; i < rightVal; i++ {
			listElems = append(listElems, &object.Integer{Value: i})
		}
		return &object.List{Elements: listElems}
	} else if rightVal < leftVal {
		size := leftVal - rightVal
		listElems := make([]object.Object, 0, size-1)
		for i = leftVal; i > rightVal; i-- {
			listElems = append(listElems, &object.Integer{Value: i})
		}
		return &object.List{Elements: listElems}
	}
	return &object.List{Elements: []object.Object{}}
}

func newError(format string, a ...any) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}

func nativeToBooleanObject(b bool) *object.Boolean {
	if b {
		return object.TRUE
	}
	return object.FALSE
}
