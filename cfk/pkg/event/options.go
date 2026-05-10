package event

import (
	"time"
)

type Option func(*Bus)

func WithWorkerPool(size int) Option {
	return func(b *Bus) {
		if size > 0 {
			b.workerCount = size
		}
	}
}

func WithBufferSize(n int) Option {
	return func(b *Bus) {
		if n > 0 {
			b.bufferSize = n
		}
	}
}

func WithMaxRetries(n int) Option {
	return func(b *Bus) {
		if n >= 0 {
			b.maxRetries = n
		}
	}
}

func WithBackoff(fn func(attempt int) time.Duration) Option {
	return func(b *Bus) {
		b.backoff = fn
	}
}

func defaultBackoff(attempt int) time.Duration {
	base := 100 * time.Millisecond
	maxDur := 5 * time.Second
	dur := base
	for i := 0; i < attempt; i++ {
		dur *= 2
		if dur > maxDur {
			dur = maxDur
			break
		}
	}
	return dur
}
