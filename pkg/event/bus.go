package event

import (
	"context"
	"sync"
	"time"

	"github.com/samber/mo"
)

type Event struct {
	AggregateType string
	AggregateID   string
	EventType     string
	Version       int
	Payload       map[string]interface{}
	Metadata      map[string]interface{}
}

type Handler func(Event) mo.Result[struct{}]

type FailedEvent struct {
	Event   Event
	Handler Handler
	Err     error
}

type Bus struct {
	handlers    map[string][]Handler
	ch          chan Event
	deadLetter  chan FailedEvent
	workerCount int
	bufferSize  int
	maxRetries  int
	backoff     func(attempt int) time.Duration
	wg          sync.WaitGroup
	cancel      context.CancelFunc
	mu          sync.RWMutex
}

func OkHandle() mo.Result[struct{}] {
	return mo.Ok[struct{}](struct{}{})
}

func NewBus(opts ...Option) *Bus {
	b := &Bus{
		handlers:    make(map[string][]Handler),
		ch:          make(chan Event, 1024),
		deadLetter:  make(chan FailedEvent, 256),
		workerCount: 4,
		bufferSize:  1024,
		maxRetries:  3,
		backoff:     defaultBackoff,
	}
	for _, opt := range opts {
		opt(b)
	}
	b.ch = make(chan Event, b.bufferSize)
	return b
}

func (b *Bus) Subscribe(eventType string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *Bus) Publish(evt Event) mo.Result[struct{}] {
	b.ch <- evt
	return OkHandle()
}

func (b *Bus) Start(ctx context.Context) {
	ctx, b.cancel = context.WithCancel(ctx)
	for i := 0; i < b.workerCount; i++ {
		b.wg.Add(1)
		go b.worker(ctx)
	}
}

func (b *Bus) Stop() {
	if b.cancel != nil {
		b.cancel()
	}
	close(b.ch)
	b.wg.Wait()
	close(b.deadLetter)
}

func (b *Bus) DeadLetters() <-chan FailedEvent {
	return b.deadLetter
}

func (b *Bus) worker(ctx context.Context) {
	defer b.wg.Done()
	for evt := range b.ch {
		select {
		case <-ctx.Done():
			return
		default:
		}
		b.dispatch(ctx, evt)
	}
}

func (b *Bus) dispatch(_ context.Context, evt Event) {
	b.mu.RLock()
	handlers := b.handlers[evt.EventType]
	wildcardHandlers := b.handlers["*"]
	b.mu.RUnlock()

	allHandlers := make([]Handler, 0, len(handlers)+len(wildcardHandlers))
	allHandlers = append(allHandlers, handlers...)
	allHandlers = append(allHandlers, wildcardHandlers...)

	for _, handler := range allHandlers {
		b.executeWithRetry(evt, handler)
	}
}

func (b *Bus) executeWithRetry(evt Event, handler Handler) {
	var lastErr error
	for attempt := 0; attempt <= b.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(b.backoff(attempt))
		}
		if r := handler(evt); r.IsError() {
			lastErr = r.Error()
			continue
		}
		return
	}
	select {
	case b.deadLetter <- FailedEvent{Event: evt, Handler: handler, Err: lastErr}:
	default:
	}
}
