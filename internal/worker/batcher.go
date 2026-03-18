package worker

import "sync"

type Batcher struct {
	mu    sync.Mutex
	size  int
	count int
}

func NewBatcher(size int) *Batcher {
	if size <= 0 {
		size = 100
	}

	return &Batcher{size: size}
}

func (b *Batcher) Add() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.count++
}

func (b *Batcher) Flush() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	flushed := b.count
	b.count = 0
	return flushed
}

func (b *Batcher) Size() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.count
}
