package vm

import (
	"blue/code"
	"blue/object"
)

type Frame struct {
	fun *object.CompiledFunction
	ip  int
}

func NewFrame(fn *object.CompiledFunction) *Frame {
	return &Frame{fun: fn, ip: -1}
}
func (f *Frame) Instructions() code.Instructions {
	return f.fun.Instructions
}
