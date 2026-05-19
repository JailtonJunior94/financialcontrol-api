package health

import (
	"context"
	"errors"
	"testing"

	"github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeManager is a minimal test double for manager.Manager.
// Only Ping has meaningful behaviour; the other methods satisfy the interface.
type fakeManager struct {
	pingErr error
}

func (f *fakeManager) Driver() database.Driver                                  { return "" }
func (f *fakeManager) DBTX(_ context.Context) database.DBTX                    { return nil }
func (f *fakeManager) BeginTx(_ context.Context, _ database.TxOptions) (database.Tx, error) {
	return nil, nil
}
func (f *fakeManager) Ping(_ context.Context) error    { return f.pingErr }
func (f *fakeManager) Shutdown(_ context.Context) error { return nil }

func TestNew_RegistersExactlyOneMSSQLEntry(t *testing.T) {
	mgr := &fakeManager{}
	checks := New(mgr)

	m := checks.Map()
	assert.Len(t, m, 1, "expected exactly one health check entry")
	assert.Contains(t, m, "mssql", "expected 'mssql' key to be present")
}

func TestMSSQLCheck_DelegatesToManager(t *testing.T) {
	pingErr := errors.New("connection refused")

	tests := []struct {
		name    string
		pingErr error
		wantErr error
	}{
		{
			name:    "ping ok returns nil",
			pingErr: nil,
			wantErr: nil,
		},
		{
			name:    "ping failure propagates original error",
			pingErr: pingErr,
			wantErr: pingErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := &fakeManager{pingErr: tt.pingErr}
			checks := New(mgr)

			checkFn, ok := checks.Map()["mssql"]
			require.True(t, ok, "mssql check must be registered")

			err := checkFn(context.Background())
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
