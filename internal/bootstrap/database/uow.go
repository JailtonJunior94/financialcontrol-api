package database

import (
	"context"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"
)

// Do executes fn inside a single transaction managed by mgr.
// The active tx is propagated implicitly via devkitdb.WithTx so that any
// repository calling mgr.DBTX(ctx) inside fn will automatically use the tx.
// Rollback is attempted on fn error; on success the tx is committed.
func Do(ctx context.Context, mgr manager.Manager, fn func(ctx context.Context) error) error {
	tx, err := mgr.BeginTx(ctx, devkitdb.TxOptions{})
	if err != nil {
		return err
	}
	ctx = devkitdb.WithTx(ctx, tx)
	if err := fn(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}
