// Package events provides the in-process event dispatcher used across modules.
// Modules depend on this package via ports; domain-specific events are defined
// inside each module that owns them.
package events

// Listener is implemented by any handler that wants to receive events from
// the dispatcher. SetData delivers the raw payload; Handle performs the action.
type Listener interface {
	SetData(data any)
	Handle() error
}

// Event is the message contract sent through the dispatcher.
type Event interface {
	GetKey() string
	GetData() any
}

// EventDispatcher is the port consumed by application-layer code that needs to
// publish or subscribe to domain events. It is intentionally small so that
// both synchronous and future asynchronous implementations can satisfy it.
type EventDispatcher interface {
	Dispatch(event Event)
	AddListener(event string, listener Listener)
}

// InProcessDispatcher is the synchronous, in-process pub/sub hub. Listeners
// register for a named event key and are called synchronously when that event
// is dispatched. Use NewInProcessDispatcher to construct.
type InProcessDispatcher struct {
	listeners map[string][]Listener
}

// NewInProcessDispatcher returns a ready-to-use InProcessDispatcher.
func NewInProcessDispatcher() *InProcessDispatcher {
	return &InProcessDispatcher{
		listeners: make(map[string][]Listener),
	}
}

func (d *InProcessDispatcher) AddListener(event string, listener Listener) {
	if d.listeners == nil {
		d.listeners = make(map[string][]Listener)
	}
	d.listeners[event] = append(d.listeners[event], listener)
}

func (d *InProcessDispatcher) Dispatch(event Event) {
	if d.listeners == nil {
		return
	}

	for _, listener := range d.listeners[event.GetKey()] {
		listener.SetData(event.GetData())
		_ = listener.Handle()
	}
}
