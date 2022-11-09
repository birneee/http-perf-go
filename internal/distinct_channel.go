package internal

import (
	"sync"
)

type DistinctChannel[T comparable] interface {
	// Add returns true if element was added
	// false when element is not distinct
	Add(elem T) bool
	Next() T
}

type distinctChannel[T comparable] struct {
	internal     chan T
	historyMutex sync.Mutex
	history      map[T]bool /* use as set; value is always true */
}

// NewDistinctChannel provides a channel with only unique elements
func NewDistinctChannel[T comparable](size int) DistinctChannel[T] {

	c := &distinctChannel[T]{
		internal: make(chan T, size),
		history:  make(map[T]bool),
	}

	return c
}

func (c *distinctChannel[T]) Add(elem T) bool {
	c.historyMutex.Lock()
	_, found := c.history[elem]
	c.history[elem] = true
	c.historyMutex.Unlock()
	if found {
		return false // element is not unique
	}
	c.internal <- elem
	return true
}

func (c *distinctChannel[T]) Next() T {
	elem := <-c.internal
	return elem
}
