package util

type Stack[T any] struct {
	Elements []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{
		Elements: make([]T, 0),
	}
}

func (s *Stack[T]) Push(element T) {
	s.Elements = append(s.Elements, element)
}

func (s *Stack[T]) Pop() T {
	var element T
	element, s.Elements = s.Elements[len(s.Elements)-1], s.Elements[:len(s.Elements)-1]
	return element
}

func (s *Stack[T]) Size() int {
	return len(s.Elements)
}
