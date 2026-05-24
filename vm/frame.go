package vm

import (
	"blue/code"
	"blue/object"
)

type Frame struct {
	cl *object.Closure
	ip int
	bp int

	deferFuns []*object.Closure

	lastInstruction code.Opcode

	// per-frame context
	inTry      bool
	inCatch    bool
	catchError string
	// last node position for error trace
	callerLastNodePos int
}

func (f *Frame) Clone() *Frame {
	if f == nil {
		return nil
	}
	var newDeferFuns []*object.Closure
	if f.deferFuns != nil {
		newDeferFuns = make([]*object.Closure, len(f.deferFuns))
		for i, df := range f.deferFuns {
			newDeferFuns[i] = df.Clone().(*object.Closure)
		}
	}
	return &Frame{
		cl:                f.cl.Clone().(*object.Closure),
		ip:                f.ip,
		bp:                f.bp,
		deferFuns:         newDeferFuns,
		lastInstruction:   f.lastInstruction,
		inTry:             f.inTry,
		inCatch:           f.inCatch,
		catchError:        f.catchError,
		callerLastNodePos: f.callerLastNodePos,
	}
}

func NewFrame(fn *object.Closure, bp int) *Frame {
	return &Frame{cl: fn, ip: -1, bp: bp, lastInstruction: code.OpInvalid, callerLastNodePos: -1}
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fun.Instructions
}

func (f *Frame) PushDeferFun(cl *object.Closure) {
	if f.deferFuns == nil {
		f.deferFuns = []*object.Closure{}
	}
	f.deferFuns = append([]*object.Closure{cl}, f.deferFuns...)
}

func (f *Frame) PopDeferFun() *object.Closure {
	if f.deferFuns == nil {
		return nil
	}
	if len(f.deferFuns) < 1 {
		return nil
	}
	deferFun := f.deferFuns[len(f.deferFuns)-1]
	f.deferFuns = f.deferFuns[:len(f.deferFuns)-1]
	return deferFun
}
