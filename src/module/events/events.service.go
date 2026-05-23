// Package events provides an in-process EventEmitter for Nexgou applications.
//
// Handlers are invoked synchronously in the order they were registered.
// For async dispatch, pass a handler that launches a goroutine.
//
// Usage:
//
//	func (s *OrderService) PlaceOrder(order Order) error {
//	    // ... business logic ...
//	    s.events.Emit("order.placed", order)
//	    return nil
//	}
//
//	func (s *NotificationService) OnStart() {
//	    s.events.On("order.placed", func(payload any) {
//	        order := payload.(Order)
//	        s.sendEmail(order)
//	    })
//	}
package events

import (
	"sync"

	"github.com/nexgou/server/src/logger"
)

// Handler is the function signature for event listeners.
type Handler func(payload any)

// EventEmitter is a thread-safe in-process pub/sub event bus.
// Listeners are called synchronously in registration order.
type EventEmitter struct {
	mu       sync.RWMutex
	handlers map[string][]namedHandler
	log      *logger.ScopedLogger
}

type namedHandler struct {
	id   string
	fn   Handler
	once bool
}

// NewEventEmitter creates a new EventEmitter.
// Depends on *logger.LoggerService.
func NewEventEmitter(log *logger.LoggerService) *EventEmitter {
	return &EventEmitter{
		handlers: make(map[string][]namedHandler),
		log:      log.WithContext("EventEmitter"),
	}
}

// On registers a persistent listener for the given event name.
// Returns a subscription ID that can be used to remove the listener.
func (e *EventEmitter) On(event string, fn Handler) string {
	return e.register(event, fn, false)
}

// Once registers a one-time listener that is automatically removed after first invocation.
func (e *EventEmitter) Once(event string, fn Handler) string {
	return e.register(event, fn, true)
}

// Off removes a listener by the subscription ID returned from On or Once.
func (e *EventEmitter) Off(event, id string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	handlers := e.handlers[event]
	filtered := handlers[:0]
	for _, h := range handlers {
		if h.id != id {
			filtered = append(filtered, h)
		}
	}
	e.handlers[event] = filtered
}

// Emit dispatches payload to all listeners registered for the given event name.
// Once-listeners are removed after being called.
func (e *EventEmitter) Emit(event string, payload any) {
	e.mu.Lock()
	handlers := make([]namedHandler, len(e.handlers[event]))
	copy(handlers, e.handlers[event])
	var persistent []namedHandler
	for _, h := range e.handlers[event] {
		if !h.once {
			persistent = append(persistent, h)
		}
	}
	e.handlers[event] = persistent
	e.mu.Unlock()

	for _, h := range handlers {
		h.fn(payload)
	}
	e.log.Debug("event emitted", "event", event, "listeners", len(handlers))
}

// RemoveAll removes all listeners for the given event name.
func (e *EventEmitter) RemoveAll(event string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.handlers, event)
}

// ListenerCount returns the number of listeners registered for an event.
func (e *EventEmitter) ListenerCount(event string) int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.handlers[event])
}

// EventNames returns all event names that have at least one listener.
func (e *EventEmitter) EventNames() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	names := make([]string, 0, len(e.handlers))
	for k := range e.handlers {
		names = append(names, k)
	}
	return names
}

func (e *EventEmitter) register(event string, fn Handler, once bool) string {
	e.mu.Lock()
	defer e.mu.Unlock()
	id := generateID()
	e.handlers[event] = append(e.handlers[event], namedHandler{id: id, fn: fn, once: once})
	return id
}
