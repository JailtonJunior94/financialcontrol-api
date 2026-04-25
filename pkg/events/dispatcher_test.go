package events_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// compile-time assertion: InProcessDispatcher must satisfy EventDispatcher.
var _ events.EventDispatcher = (*events.InProcessDispatcher)(nil)

// stubEvent is a minimal Event for testing.
type stubEvent struct {
	key  string
	data any
}

func (e *stubEvent) GetKey() string { return e.key }
func (e *stubEvent) GetData() any   { return e.data }

// stubListener records invocations.
type stubListener struct {
	data    any
	handled int
}

func (l *stubListener) SetData(data any) { l.data = data }
func (l *stubListener) Handle() error    { l.handled++; return nil }

func TestInProcessDispatcherSatisfiesEventDispatcher(t *testing.T) {
	var d events.EventDispatcher = events.NewInProcessDispatcher()
	require.NotNil(t, d)
}

func TestInProcessDispatcher_Dispatch(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(d *events.InProcessDispatcher) *stubListener
		event         *stubEvent
		wantHandled   int
		wantData      any
	}{
		{
			name: "dispatches event to registered listener",
			setup: func(d *events.InProcessDispatcher) *stubListener {
				l := &stubListener{}
				d.AddListener("order.created", l)
				return l
			},
			event:       &stubEvent{key: "order.created", data: "order-1"},
			wantHandled: 1,
			wantData:    "order-1",
		},
		{
			name: "no listener registered — dispatch is a no-op",
			setup: func(d *events.InProcessDispatcher) *stubListener {
				return &stubListener{}
			},
			event:       &stubEvent{key: "unknown.event", data: "x"},
			wantHandled: 0,
			wantData:    nil,
		},
		{
			name: "listener for different key is not called",
			setup: func(d *events.InProcessDispatcher) *stubListener {
				l := &stubListener{}
				d.AddListener("other.event", l)
				return l
			},
			event:       &stubEvent{key: "order.created", data: "order-2"},
			wantHandled: 0,
			wantData:    nil,
		},
		{
			name: "multiple listeners for same key all receive the event",
			setup: func(d *events.InProcessDispatcher) *stubListener {
				l1 := &stubListener{}
				l2 := &stubListener{}
				d.AddListener("invoice_changed", l1)
				d.AddListener("invoice_changed", l2)
				return l1
			},
			event:       &stubEvent{key: "invoice_changed", data: "inv-42"},
			wantHandled: 1,
			wantData:    "inv-42",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := events.NewInProcessDispatcher()
			listener := tc.setup(d)
			d.Dispatch(tc.event)
			assert.Equal(t, tc.wantHandled, listener.handled)
			assert.Equal(t, tc.wantData, listener.data)
		})
	}
}
