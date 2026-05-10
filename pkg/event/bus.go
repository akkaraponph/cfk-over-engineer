package event

import "fmt"

type Event struct {
	AggregateType string
	AggregateID   string
	EventType     string
	Version       int
	Payload       map[string]interface{}
	Metadata      map[string]interface{}
}

type Handler func(Event) error

type Bus struct {
	handlers map[string][]Handler
}

func NewBus() *Bus {
	return &Bus{
		handlers: make(map[string][]Handler),
	}
}

func (b *Bus) Subscribe(eventType string, handler Handler) {
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *Bus) Publish(event Event) error {
	if handlers, ok := b.handlers[event.EventType]; ok {
		for _, handler := range handlers {
			if err := handler(event); err != nil {
				return fmt.Errorf("event handler failed for %s: %w", event.EventType, err)
			}
		}
	}
	if wildcardHandlers, ok := b.handlers["*"]; ok {
		for _, handler := range wildcardHandlers {
			if err := handler(event); err != nil {
				return fmt.Errorf("wildcard handler failed for %s: %w", event.EventType, err)
			}
		}
	}
	return nil
}
