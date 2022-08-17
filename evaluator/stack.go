package evaluator

import (
	"blue/object"
	"container/list"
)

type Stack struct {
	s *list.List
}

func NewStack() *Stack {
	return &Stack{s: list.New()}
}

func (s *Stack) Push(item *object.Object) {
	s.s.PushFront(item)
}

func (s *Stack) Pop() *object.Object {
	if s.s.Len() == 0 {
		return nil
	}
	e := s.s.Front()
	s.s.Remove(e)
	return e.Value.(*object.Object)
}

func (s *Stack) Len() int {
	return s.s.Len()
}
