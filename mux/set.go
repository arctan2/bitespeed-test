package mux

type StringSet map[string]struct{}

func (s StringSet) Add(element string) {
	s[element] = struct{}{}
}

func (s StringSet) Contains(element string) bool {
	_, found := s[element]
	return found
}

func (s StringSet) Remove(element string) {
	delete(s, element)
}
