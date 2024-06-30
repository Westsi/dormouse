package util

import "fmt"

var OBJSIZE int = 8

type Armstack[T any] struct {
	Elements []T
	size     int
}

func NewAStack[T any](size int) *Armstack[T] {
	if size%OBJSIZE != 0 {
		fmt.Printf("Size needs to be a multiple of %d\n", OBJSIZE)
	}
	return &Armstack[T]{
		size:     size,
		Elements: make([]T, size/OBJSIZE),
	}
}

func (s *Armstack[T]) Set(element T, index int) {
	if index%OBJSIZE != 0 {
		fmt.Println("UH OH!!")
	}
	s.Elements[(index/OBJSIZE)-1] = element
}

func (s *Armstack[T]) Get(index int) T {
	element := s.Elements[(index/OBJSIZE)-1]
	return element
}

func (s *Armstack[T]) Size() int {
	return s.size
}
