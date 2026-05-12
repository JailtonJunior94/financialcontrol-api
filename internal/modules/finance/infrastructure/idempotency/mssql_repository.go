package idempotency

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

const (
	getIdempotencyKey = `SELECT [RequestHash],[ResponseBody],[StatusCode]
	FROM finance.IdempotencyKeys (NOLOCK)
	WHERE [UserId] = @userId AND [Endpoint] = @endpoint AND [Key] = @key
	  AND [ExpiresAt] > @now`

	saveIdempotencyKey = `INSERT INTO finance.IdempotencyKeys
		([UserId],[Endpoint],[Key],[RequestHash],[ResponseBody],[StatusCode],[CreatedAt],[ExpiresAt])
		VALUES
		(@userId,@endpoint,@key,@requestHash,@responseBody,@statusCode,@createdAt,@expiresAt)`

	purgeExpiredIdempotencyKeys = `DELETE FROM finance.IdempotencyKeys WHERE [ExpiresAt] <= @now`
)

var _ IdempotencyRepository = (*MSSQLRepository)(nil)

// MSSQLRepository implements IdempotencyRepository against MSSQL.
type MSSQLRepository struct {
	db devkitdb.DBTX
}

func NewMSSQLRepository(db devkitdb.DBTX) *MSSQLRepository {
	return &MSSQLRepository{db: db}
}

func (r *MSSQLRepository) Get(ctx context.Context, userID identityvo.UserID, endpoint, key string) (*IdempotencyHit, error) {
	var hit IdempotencyHit
	err := r.db.QueryRowContext(ctx, getIdempotencyKey,
		sql.Named("userId", userID.String()),
		sql.Named("endpoint", endpoint),
		sql.Named("key", key),
		sql.Named("now", time.Now().UTC()),
	).Scan(&hit.RequestHash, &hit.ResponseBody, &hit.StatusCode)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("mssql: get idempotency key: %w", err)
	}
	return &hit, nil
}

func (r *MSSQLRepository) Save(ctx context.Context, rec IdempotencyRecord) error {
	_, err := r.db.ExecContext(ctx, saveIdempotencyKey,
		sql.Named("userId", rec.UserID),
		sql.Named("endpoint", rec.Endpoint),
		sql.Named("key", rec.Key),
		sql.Named("requestHash", rec.RequestHash),
		sql.Named("responseBody", rec.ResponseBody),
		sql.Named("statusCode", rec.StatusCode),
		sql.Named("createdAt", rec.CreatedAt),
		sql.Named("expiresAt", rec.ExpiresAt),
	)
	if err != nil {
		return fmt.Errorf("mssql: save idempotency key: %w", err)
	}
	return nil
}

func (r *MSSQLRepository) PurgeExpired(ctx context.Context, now time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, purgeExpiredIdempotencyKeys,
		sql.Named("now", now.UTC()),
	)
	if err != nil {
		return 0, fmt.Errorf("mssql: purge expired idempotency keys: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("mssql: purge expired rows affected: %w", err)
	}
	return affected, nil
}
