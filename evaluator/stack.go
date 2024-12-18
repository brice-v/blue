package evaluator

import (
	"blue/evaluator/list"
	"blue/token"
	"bytes"
	"fmt"
	"sync"
)

type Stack[T any] struct {
	s *list.List[T]
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{s: list.New[T]()}
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

type TokenStackSet struct {
	s    *list.List[token.Token]
	m    map[token.Token]struct{}
	lock sync.RWMutex
}

func NewTokenStackSet() *TokenStackSet {
	return &TokenStackSet{s: list.New[token.Token](), m: make(map[token.Token]struct{}), lock: sync.RWMutex{}}
}

func (s *TokenStackSet) Push(tok token.Token) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.m[tok]; !ok {
		s.s.PushFront(tok)
		s.m[tok] = struct{}{}
	}
}

func (s *TokenStackSet) Pop() token.Token {
	s.lock.Lock()
	defer s.lock.Unlock()
	var result token.Token
	if s.s.Len() == 0 {
		return result
	}
	e := s.s.Front()
	s.s.Remove(e)
	tok := e.Value
	delete(s.m, tok)
	return tok
}

func (s *TokenStackSet) PopBack() token.Token {
	s.lock.Lock()
	defer s.lock.Unlock()
	var result token.Token
	if s.s.Len() == 0 {
		return result
	}
	e := s.s.Back()
	s.s.Remove(e)
	tok := e.Value
	delete(s.m, tok)
	return tok
}

func (s *TokenStackSet) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.s.Len()
}

func (s *TokenStackSet) RemoveAllEntries() {
	s.lock.Lock()
	defer s.lock.Unlock()
	// Just instantiate new instances of the map and list for 'removing everything'
	s.m = make(map[token.Token]struct{})
	s.s = list.New[token.Token]()
}
