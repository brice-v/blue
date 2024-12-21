package evaluator

import (
	"blue/token"
	"blue/util"
	"sync"
)

type TokenStackSet struct {
	s    *util.List[token.Token]
	m    map[token.Token]struct{}
	lock sync.RWMutex
}

func NewTokenStackSet() *TokenStackSet {
	return &TokenStackSet{s: util.NewList[token.Token](), m: make(map[token.Token]struct{}), lock: sync.RWMutex{}}
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
	s.s = util.NewList[token.Token]()
}
