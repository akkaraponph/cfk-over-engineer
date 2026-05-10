package event

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/samber/mo"
)

func TestPublish_EventReachesHandler(t *testing.T) {
	bus := NewBus(WithWorkerPool(1), WithBufferSize(64))

	var mu sync.Mutex
	received := []Event{}
	bus.Subscribe("user.created", func(evt Event) mo.Result[struct{}] {
		mu.Lock()
		received = append(received, evt)
		mu.Unlock()
		return OkHandle()
	})

	ctx := context.Background()
	bus.Start(ctx)

	evt := Event{
		AggregateType: "user",
		AggregateID:   "u-1",
		EventType:     "user.created",
		Version:       1,
		Payload:       map[string]interface{}{"id": "u-1"},
	}
	bus.Publish(evt)

	time.Sleep(50 * time.Millisecond)

	bus.Stop()

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 1 {
		t.Fatalf("expected 1 event, got %d", len(received))
	}
	if received[0].EventType != "user.created" {
		t.Errorf("expected event type 'user.created', got '%s'", received[0].EventType)
	}
	if received[0].AggregateID != "u-1" {
		t.Errorf("expected aggregate ID 'u-1', got '%s'", received[0].AggregateID)
	}
}

func TestSubscribe_WildcardReceivesAllEvents(t *testing.T) {
	bus := NewBus(WithWorkerPool(1), WithBufferSize(64))

	var mu sync.Mutex
	received := []Event{}
	bus.Subscribe("*", func(evt Event) mo.Result[struct{}] {
		mu.Lock()
		received = append(received, evt)
		mu.Unlock()
		return OkHandle()
	})

	ctx := context.Background()
	bus.Start(ctx)

	bus.Publish(Event{AggregateType: "user", EventType: "user.created", Version: 1})
	bus.Publish(Event{AggregateType: "pocket", EventType: "pocket.created", Version: 1})
	bus.Publish(Event{AggregateType: "transfer", EventType: "transfer.initiated", Version: 1})

	time.Sleep(100 * time.Millisecond)

	bus.Stop()

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 3 {
		t.Fatalf("expected 3 events, got %d", len(received))
	}
}

func TestSubscribe_SpecificAndWildcardBothReceive(t *testing.T) {
	bus := NewBus(WithWorkerPool(1), WithBufferSize(64))

	var mu sync.Mutex
	specificReceived := []Event{}
	wildcardReceived := []Event{}

	bus.Subscribe("user.created", func(evt Event) mo.Result[struct{}] {
		mu.Lock()
		specificReceived = append(specificReceived, evt)
		mu.Unlock()
		return OkHandle()
	})
	bus.Subscribe("*", func(evt Event) mo.Result[struct{}] {
		mu.Lock()
		wildcardReceived = append(wildcardReceived, evt)
		mu.Unlock()
		return OkHandle()
	})

	ctx := context.Background()
	bus.Start(ctx)

	bus.Publish(Event{EventType: "user.created", Version: 1})
	bus.Publish(Event{EventType: "pocket.created", Version: 1})

	time.Sleep(100 * time.Millisecond)

	bus.Stop()

	mu.Lock()
	defer mu.Unlock()
	if len(specificReceived) != 1 {
		t.Errorf("expected 1 specific event, got %d", len(specificReceived))
	}
	if len(wildcardReceived) != 2 {
		t.Errorf("expected 2 wildcard events, got %d", len(wildcardReceived))
	}
}

func TestRetry_HandlerFailsThenSucceeds(t *testing.T) {
	var mu sync.Mutex
	callCount := 0
	errHandlerFailed := &handlerError{msg: "handler failed"}

	bus := NewBus(
		WithWorkerPool(1),
		WithBufferSize(64),
		WithMaxRetries(3),
		WithBackoff(func(attempt int) time.Duration {
			return 5 * time.Millisecond
		}),
	)

	bus.Subscribe("retry.test", func(evt Event) mo.Result[struct{}] {
		mu.Lock()
		callCount++
		if callCount < 3 {
			mu.Unlock()
			return mo.Err[struct{}](errHandlerFailed)
		}
		mu.Unlock()
		return OkHandle()
	})

	ctx := context.Background()
	bus.Start(ctx)

	bus.Publish(Event{EventType: "retry.test", Version: 1})

	time.Sleep(200 * time.Millisecond)

	bus.Stop()

	mu.Lock()
	defer mu.Unlock()
	if callCount < 3 {
		t.Errorf("expected at least 3 calls, got %d", callCount)
	}
}

type handlerError struct {
	msg string
}

func (e *handlerError) Error() string {
	return e.msg
}

func TestDeadLetter_MaxRetriesExceeded(t *testing.T) {
	bus := NewBus(
		WithWorkerPool(1),
		WithBufferSize(64),
		WithMaxRetries(2),
		WithBackoff(func(attempt int) time.Duration {
			return 5 * time.Millisecond
		}),
	)

	bus.Subscribe("deadletter.test", func(evt Event) mo.Result[struct{}] {
		return mo.Err[struct{}](&handlerError{msg: "always fails"})
	})

	ctx := context.Background()
	bus.Start(ctx)

	bus.Publish(Event{
		AggregateType: "test",
		AggregateID:   "dl-1",
		EventType:     "deadletter.test",
		Version:       1,
	})

	time.Sleep(200 * time.Millisecond)

	bus.Stop()

	var deadLetters []FailedEvent
	for {
		select {
		case fe, ok := <-bus.DeadLetters():
			if !ok {
				goto done
			}
			deadLetters = append(deadLetters, fe)
		default:
			goto done
		}
	}
done:

	if len(deadLetters) != 1 {
		t.Errorf("expected 1 dead letter, got %d", len(deadLetters))
	}
	if len(deadLetters) > 0 {
		if deadLetters[0].Event.EventType != "deadletter.test" {
			t.Errorf("expected dead letter event type 'deadletter.test', got '%s'", deadLetters[0].Event.EventType)
		}
		if deadLetters[0].Err == nil {
			t.Error("expected dead letter to have an error")
		}
	}
}

func TestStop_GracefulShutdown(t *testing.T) {
	bus := NewBus(WithWorkerPool(2), WithBufferSize(64))

	var mu sync.Mutex
	count := 0
	bus.Subscribe("*", func(evt Event) mo.Result[struct{}] {
		mu.Lock()
		count++
		mu.Unlock()
		return OkHandle()
	})

	ctx := context.Background()
	bus.Start(ctx)

	for i := 0; i < 10; i++ {
		bus.Publish(Event{EventType: "test.event", Version: 1})
	}

	time.Sleep(100 * time.Millisecond)

	bus.Stop()

	mu.Lock()
	defer mu.Unlock()
	if count != 10 {
		t.Errorf("expected 10 events processed, got %d", count)
	}
}

func TestNewBus_DefaultOptions(t *testing.T) {
	bus := NewBus()
	if bus == nil {
		t.Fatal("expected non-nil bus")
	}
	if bus.workerCount != 4 {
		t.Errorf("expected default worker count 4, got %d", bus.workerCount)
	}
	if bus.maxRetries != 3 {
		t.Errorf("expected default max retries 3, got %d", bus.maxRetries)
	}
}

func TestNewBus_CustomOptions(t *testing.T) {
	bus := NewBus(
		WithWorkerPool(8),
		WithBufferSize(256),
		WithMaxRetries(5),
	)
	if bus.workerCount != 8 {
		t.Errorf("expected worker count 8, got %d", bus.workerCount)
	}
	if bus.bufferSize != 256 {
		t.Errorf("expected buffer size 256, got %d", bus.bufferSize)
	}
	if bus.maxRetries != 5 {
		t.Errorf("expected max retries 5, got %d", bus.maxRetries)
	}
}

func TestPublish_ReturnsOkHandle(t *testing.T) {
	bus := NewBus(WithWorkerPool(1), WithBufferSize(64))

	ctx := context.Background()
	bus.Start(ctx)

	result := bus.Publish(Event{EventType: "test", Version: 1})
	if result.IsError() {
		t.Errorf("expected Publish to return Ok, got error: %v", result.Error())
	}

	bus.Stop()
}

func TestMultipleHandlers_SameEventType(t *testing.T) {
	bus := NewBus(WithWorkerPool(1), WithBufferSize(64))

	var mu sync.Mutex
	handler1Count := 0
	handler2Count := 0

	bus.Subscribe("multi.test", func(evt Event) mo.Result[struct{}] {
		mu.Lock()
		handler1Count++
		mu.Unlock()
		return OkHandle()
	})
	bus.Subscribe("multi.test", func(evt Event) mo.Result[struct{}] {
		mu.Lock()
		handler2Count++
		mu.Unlock()
		return OkHandle()
	})

	ctx := context.Background()
	bus.Start(ctx)

	bus.Publish(Event{EventType: "multi.test", Version: 1})

	time.Sleep(50 * time.Millisecond)

	bus.Stop()

	mu.Lock()
	defer mu.Unlock()
	if handler1Count != 1 {
		t.Errorf("expected handler1 to be called once, got %d", handler1Count)
	}
	if handler2Count != 1 {
		t.Errorf("expected handler2 to be called once, got %d", handler2Count)
	}
}
