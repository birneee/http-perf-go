package internal

import "sync"

type CondHelper[T any] interface {
	SetState(T)
	UpdateState(func(T) T)
	// Wait until the condition is true
	Wait(cond func(T) bool)
}

type condHelper[T any] struct {
	cond  *sync.Cond
	state T
}

// NewCondHelper returns works similar to sync.Cond but the state is wrapped by the data structure
func NewCondHelper[T any](initialState T) CondHelper[T] {
	return &condHelper[T]{
		cond:  sync.NewCond(&sync.Mutex{}),
		state: initialState,
	}
}

func (c *condHelper[T]) SetState(state T) {
	c.cond.L.Lock()
	c.state = state
	c.cond.Broadcast()
	c.cond.L.Unlock()
}

func (c *condHelper[T]) UpdateState(update func(T) T) {
	c.SetState(update(c.state))
}

func (c *condHelper[T]) Wait(condition func(T) bool) {
	c.cond.L.Lock()
	for !condition(c.state) {
		c.cond.Wait()
	}
	c.cond.L.Unlock()
}
