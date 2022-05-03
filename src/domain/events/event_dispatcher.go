package events

type Listener interface {
	SetData(data interface{})
	Handle() error
}

type Event interface {
	GetKey() string
	GetData() interface{}
}

type EventDispatcher struct {
	Listeners map[string][]Listener
}

func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		Listeners: make(map[string][]Listener),
	}
}

func (e *EventDispatcher) AddListener(event string, listener Listener) {
	if e.Listeners == nil {
		e.Listeners = make(map[string][]Listener)
	}
	e.Listeners[event] = append(e.Listeners[event], listener)
}

func (e *EventDispatcher) Dispatch(event Event) {
	if e.Listeners == nil {
		return
	}

	for _, listener := range e.Listeners[event.GetKey()] {
		listener.SetData(event.GetData())
		listener.Handle()
	}
}
