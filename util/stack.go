package util

import (
	"bytes"
	"fmt"
)

type Stack[T any] struct {
	s *List[T]
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{s: NewList[T]()}
}

func (s *Stack[T]) Push(item T) {
	s.s.PushFront(item)
}

func (s *Stack[T]) Pop() T {
	var result T
	if s.s.Len() == 0 {
		return result
	}
	e := s.s.Front()
	s.s.Remove(e)
	return e.Value
}

func (s *Stack[T]) Peek() T {
	var result T
	if s.s.Len() == 0 {
		return result
	}
	e := s.s.Front()
	return e.Value
}

func (s *Stack[T]) PopBack() T {
	var result T
	if s.s.Len() == 0 {
		return result
	}
	e := s.s.Back()
	s.s.Remove(e)
	return e.Value
}

func (s *Stack[T]) Len() int {
	return s.s.Len()
}

func (s *Stack[T]) String() string {
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("%T{", s))
	for e := s.s.Front(); e != nil; e = e.Next() {
		out.WriteString(fmt.Sprintf("%#v,", e.Value))
	}
	out.WriteString("}")
	return out.String()
}
