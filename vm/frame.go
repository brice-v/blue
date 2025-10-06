package vm

import (
	"blue/code"
	"blue/object"
)

type Frame struct {
	cl *object.Closure
	ip int
	bp int
}

func NewFrame(fn *object.Closure, bp int) *Frame {
	return &Frame{cl: fn, ip: -1, bp: bp}
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fun.Instructions
}
