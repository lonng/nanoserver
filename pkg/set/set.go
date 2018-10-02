package set

import "sync"

type Set struct {
	mu sync.Mutex
	m  map[string]struct{}
}

func (s *Set) Contains(val string) bool {
	if val == "" {
		return true
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.m[val]; ok {
		return true
	}
	return false
}

func (s *Set) Add(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[val] = struct{}{}
}

func (s *Set) Remove(val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, val)
}

func New() *Set {
	return &Set{
		m: make(map[string]struct{}),
	}
}
