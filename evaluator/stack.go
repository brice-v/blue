package evaluator

import (
	"blue/token"
	"bytes"
	"container/list"
	"fmt"
)

type Stack[T any] struct {
	s *list.List
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{s: list.New()}
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
	return e.Value.(T)
}

func (s *Stack[T]) PopBack() T {
	var result T
	if s.s.Len() == 0 {
		return result
	}
	e := s.s.Back()
	s.s.Remove(e)
	return e.Value.(T)
}

func (s *Stack[T]) Len() int {
	return s.s.Len()
}

func (s *Stack[T]) String() string {
	var out bytes.Buffer
	out.WriteString("Stack{")
	for e := s.s.Front(); e != nil; e = e.Next() {
		out.WriteString(fmt.Sprintf("%#v,", e.Value))
	}
	out.WriteString("}")
	return out.String()
}

type TokenStackSet struct {
	s *list.List
	m map[token.Token]struct{}
}

func NewTokenStackSet() *TokenStackSet {
	return &TokenStackSet{s: list.New(), m: make(map[token.Token]struct{})}
}

func (s *TokenStackSet) Push(tok token.Token) {
	if _, ok := s.m[tok]; !ok {
		s.s.PushFront(tok)
		s.m[tok] = struct{}{}
	}
}

func (s *TokenStackSet) Pop() token.Token {
	var result token.Token
	if s.s.Len() == 0 {
		return result
	}
	e := s.s.Front()
	s.s.Remove(e)
	tok := e.Value.(token.Token)
	delete(s.m, tok)
	return tok
}

func (s *TokenStackSet) PopBack() token.Token {
	var result token.Token
	if s.s.Len() == 0 {
		return result
	}
	e := s.s.Back()
	s.s.Remove(e)
	tok := e.Value.(token.Token)
	delete(s.m, tok)
	return tok
}

func (s *TokenStackSet) Len() int {
	return s.s.Len()
}

func (s *TokenStackSet) RemoveAllEntries() {
	// Just instantiate new instances of the map and list for 'removing everything'
	s.m = make(map[token.Token]struct{})
	s.s = list.New()
}
