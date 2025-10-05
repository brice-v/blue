package vm

import (
	"blue/code"
	"blue/object"
)

type Frame struct {
	fun *object.CompiledFunction
	ip  int
	bp  int
}

func NewFrame(fn *object.CompiledFunction, bp int) *Frame {
	return &Frame{fun: fn, ip: -1, bp: bp}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fun.Instructions
}
