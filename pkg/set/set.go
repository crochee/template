package set

type exists struct{}

type Set struct {
	m map[interface{}]exists
}

func (s *Set) Add(items ...interface{}) {
	for _, item := range items {
		s.m[item] = exists{}
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
	s.m = make(map[interface{}]exists)
	s.Add(items...)
	return s
}

func IsContains(item string, items []string) bool {
	s := &Set{}
	s.m = make(map[interface{}]exists)
	for _, item := range items {
		s.Add(item)
	}
	return s.Contains(item)
}
