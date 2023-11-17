package set

type Set struct {
	m map[interface{}]struct{}
}

func (s *Set) Add(items ...interface{}) {
	for _, item := range items {
		s.m[item] = struct{}{}
	}
}

func (s *Set) Remove(item interface{}) {
	delete(s.m, item)
}

func (s *Set) Contains(item interface{}) bool {
	_, ok := s.m[item]
	return ok
}

func (s *Set) Size() int {
	return len(s.m)
}

func NewSet(items ...interface{}) *Set {
	s := &Set{}
	s.m = make(map[interface{}]struct{})
	s.Add(items...)
	return s
}

func IsContains(item string, items []string) bool {
	s := &Set{}
	s.m = make(map[interface{}]struct{})
	for _, item := range items {
		s.Add(item)
	}
	return s.Contains(item)
}
