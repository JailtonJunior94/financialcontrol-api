package database_test

import (
	"context"
	"errors"
	"testing"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	bootstrapdb "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/database"
)

// mockTx implements database.Tx for unit testing.
type mockTx struct {
	mock.Mock
}

func (m *mockTx) ExecContext(ctx context.Context, query string, args ...any) (devkitdb.Result, error) {
	ret := m.Called(ctx, query, args)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(devkitdb.Result), ret.Error(1)
}

func (m *mockTx) QueryContext(ctx context.Context, query string, args ...any) (devkitdb.Rows, error) {
	ret := m.Called(ctx, query, args)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(devkitdb.Rows), ret.Error(1)
}

func (m *mockTx) QueryRowContext(ctx context.Context, query string, args ...any) devkitdb.Row {
	ret := m.Called(ctx, query, args)
	return ret.Get(0).(devkitdb.Row)
}

func (m *mockTx) Commit(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *mockTx) Rollback(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

// mockManager implements a minimal manager.Manager for unit testing.
type mockManager struct {
	mock.Mock
}

func (m *mockManager) Driver() devkitdb.Driver {
	return m.Called().Get(0).(devkitdb.Driver)
}

func (m *mockManager) DBTX(ctx context.Context) devkitdb.DBTX {
	ret := m.Called(ctx)
	if ret.Get(0) == nil {
		return nil
	}
	return ret.Get(0).(devkitdb.DBTX)
}

func (m *mockManager) BeginTx(ctx context.Context, opts devkitdb.TxOptions) (devkitdb.Tx, error) {
	ret := m.Called(ctx, opts)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(devkitdb.Tx), ret.Error(1)
}

func (m *mockManager) Ping(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *mockManager) Shutdown(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func TestDo(t *testing.T) {
	errFn := errors.New("fn failed")
	errBegin := errors.New("begin failed")

	tests := []struct {
		name         string
		setupMgr     func(*mockManager, *mockTx)
		fn           func(ctx context.Context) error
		wantErr      error
		wantCommit   bool
		wantRollback bool
	}{
		{
			name: "commits when fn returns nil",
			setupMgr: func(mgr *mockManager, tx *mockTx) {
				mgr.On("BeginTx", mock.Anything, devkitdb.TxOptions{}).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
			},
			fn:         func(ctx context.Context) error { return nil },
			wantCommit: true,
		},
		{
			name: "rolls back and propagates fn error",
			setupMgr: func(mgr *mockManager, tx *mockTx) {
				mgr.On("BeginTx", mock.Anything, devkitdb.TxOptions{}).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
			},
			fn:           func(ctx context.Context) error { return errFn },
			wantErr:      errFn,
			wantRollback: true,
		},
		{
			name: "returns begin error immediately without tx operations",
			setupMgr: func(mgr *mockManager, _ *mockTx) {
				mgr.On("BeginTx", mock.Anything, devkitdb.TxOptions{}).Return(nil, errBegin)
			},
			fn:      func(ctx context.Context) error { return nil },
			wantErr: errBegin,
		},
		{
			name: "commit error is propagated",
			setupMgr: func(mgr *mockManager, tx *mockTx) {
				mgr.On("BeginTx", mock.Anything, devkitdb.TxOptions{}).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(errors.New("commit failed"))
			},
			fn:      func(ctx context.Context) error { return nil },
			wantErr: errors.New("commit failed"),
		},
		{
			name: "tx is available via devkitdb.FromContext inside fn",
			setupMgr: func(mgr *mockManager, tx *mockTx) {
				mgr.On("BeginTx", mock.Anything, devkitdb.TxOptions{}).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
			},
			fn: func(ctx context.Context) error {
				got, ok := devkitdb.FromContext(ctx)
				if !ok {
					return errors.New("tx not in context")
				}
				if got == nil {
					return errors.New("tx is nil")
				}
				return nil
			},
			wantCommit: true,
		},
		{
			name: "rollback error is ignored and fn error is returned",
			setupMgr: func(mgr *mockManager, tx *mockTx) {
				mgr.On("BeginTx", mock.Anything, devkitdb.TxOptions{}).Return(tx, nil)
				tx.On("Rollback", mock.Anything).Return(errors.New("rollback failed"))
			},
			fn:           func(ctx context.Context) error { return errFn },
			wantErr:      errFn,
			wantRollback: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tx := &mockTx{}
			mgr := &mockManager{}
			if tc.setupMgr != nil {
				tc.setupMgr(mgr, tx)
			}

			err := bootstrapdb.Do(context.Background(), mgr, tc.fn)

			if tc.wantErr != nil {
				require.Error(t, err)
				// compare message for cases where we create a new error instance
				assert.Equal(t, tc.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			if tc.wantCommit {
				tx.AssertCalled(t, "Commit", mock.Anything)
				tx.AssertNotCalled(t, "Rollback")
			}
			if tc.wantRollback {
				tx.AssertCalled(t, "Rollback", mock.Anything)
				tx.AssertNotCalled(t, "Commit")
			}

			mgr.AssertExpectations(t)
			tx.AssertExpectations(t)
		})
	}
}
