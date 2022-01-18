package stream

import "sync"

func NewSemaphore(size int) Semaphore {
	return Semaphore{
		size: size,
		ch:   make(chan func(), size),
	}
}

type Semaphore struct {
	size int
	ch   chan func()
}

func (w Semaphore) Wait() <-chan func() { return w.ch }

func (w Semaphore) Unblock() {
	for i := 0; i < w.size; i++ {
		w.unblockOne()
	}
}

func (w Semaphore) unblockOne() {
	var once sync.Once
	w.ch <- func() { once.Do(w.unblockOne) }
}
