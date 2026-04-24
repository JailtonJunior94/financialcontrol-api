// Package events provides the in-process event dispatcher used across modules.
// Modules depend on this package via ports; domain-specific events are defined
// inside each module that owns them.
package events

// Listener is implemented by any handler that wants to receive events from
// the dispatcher. SetData delivers the raw payload; Handle performs the action.
type Listener interface {
	SetData(data interface{})
	Handle() error
}

// Event is the message contract sent through the dispatcher.
type Event interface {
	GetKey() string
	GetData() interface{}
}

// Dispatcher is the in-process pub/sub hub. Listeners register for a named
// event key and are called synchronously when that event is dispatched.
type Dispatcher struct {
	listeners map[string][]Listener
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		listeners: make(map[string][]Listener),
	}
}

func (d *Dispatcher) AddListener(event string, listener Listener) {
	if d.listeners == nil {
		d.listeners = make(map[string][]Listener)
	}
	d.listeners[event] = append(d.listeners[event], listener)
}

func (d *Dispatcher) Dispatch(event Event) {
	if d.listeners == nil {
		return
	}

	for _, listener := range d.listeners[event.GetKey()] {
		listener.SetData(event.GetData())
		_ = listener.Handle()
	}
}
