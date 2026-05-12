package idempotency_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	devkitmocks "github.com/JailtonJunior94/devkit-go/pkg/database/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/idempotency"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// anyArgs returns n mock.Anything matchers — required because devkit-go mocks spread
// variadic args individually, so testify needs one matcher per argument position.
func anyArgs(n int) []interface{} {
	args := make([]interface{}, n)
	for i := range args {
		args[i] = mock.Anything
	}
	return args
}

func newIdempotencyRepo(t *testing.T) (*idempotency.MSSQLRepository, *devkitmocks.MockDBTX) {
	t.Helper()
	db := devkitmocks.NewMockDBTX(t)
	return idempotency.NewMSSQLRepository(db), db
}

func newTestRecord(userID identityvo.UserID) idempotency.IdempotencyRecord {
	return idempotency.IdempotencyRecord{
		UserID:       userID.String(),
		Endpoint:     "/api/v1/transactions",
		Key:          "idem-key-123",
		RequestHash:  "abc123hash",
		ResponseBody: `{"id":"tx-1"}`,
		StatusCode:   201,
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(24 * time.Hour),
	}
}

// ---- Get ----

func TestMSSQLRepository_Get_Hit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newIdempotencyRepo(t)
	userID := identityvo.NewUserID()

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			*(args[0].(*string)) = "abc123hash"
			*(args[1].(*string)) = `{"id":"tx-1"}`
			*(args[2].(*int)) = 201
		}).Return(nil)
	// ctx + query + userId + endpoint + key + now = 6 total
	db.On("QueryRowContext", anyArgs(6)...).Return(row)

	hit, err := repo.Get(ctx, userID, "/api/v1/transactions", "idem-key-123")
	require.NoError(t, err)
	require.NotNil(t, hit)
	assert.Equal(t, "abc123hash", hit.RequestHash)
	assert.Equal(t, 201, hit.StatusCode)
}

func TestMSSQLRepository_Get_Miss(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newIdempotencyRepo(t)
	userID := identityvo.NewUserID()

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything, mock.Anything, mock.Anything).Return(sql.ErrNoRows)
	db.On("QueryRowContext", anyArgs(6)...).Return(row)

	hit, err := repo.Get(ctx, userID, "/api/v1/transactions", "missing-key")
	require.NoError(t, err)
	assert.Nil(t, hit)
}

func TestMSSQLRepository_Get_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newIdempotencyRepo(t)
	userID := identityvo.NewUserID()

	row := devkitmocks.NewMockRow(t)
	row.On("Scan", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("timeout"))
	db.On("QueryRowContext", anyArgs(6)...).Return(row)

	_, err := repo.Get(ctx, userID, "/api/v1/transactions", "key")
	require.Error(t, err)
	assert.ErrorContains(t, err, "mssql: get idempotency key")
}

// ---- Save ----

func TestMSSQLRepository_Save_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newIdempotencyRepo(t)
	userID := identityvo.NewUserID()
	rec := newTestRecord(userID)

	result := devkitmocks.NewMockResult(t)
	// ctx + query + userId + endpoint + key + requestHash + responseBody + statusCode + createdAt + expiresAt = 10 total
	db.On("ExecContext", anyArgs(10)...).Return(result, nil)

	require.NoError(t, repo.Save(ctx, rec))
}

func TestMSSQLRepository_Save_PKViolation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newIdempotencyRepo(t)
	userID := identityvo.NewUserID()
	rec := newTestRecord(userID)

	// simulate PK violation (race detection for duplicate key)
	db.On("ExecContext", anyArgs(10)...).Return(nil, errors.New("duplicate key"))

	err := repo.Save(ctx, rec)
	require.Error(t, err)
	assert.ErrorContains(t, err, "mssql: save idempotency key")
}

func TestMSSQLRepository_Save_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newIdempotencyRepo(t)
	userID := identityvo.NewUserID()
	rec := newTestRecord(userID)

	db.On("ExecContext", anyArgs(10)...).Return(nil, errors.New("duplicate key"))

	err := repo.Save(ctx, rec)
	require.Error(t, err)
	assert.ErrorContains(t, err, "mssql: save idempotency key")
}

// ---- PurgeExpired ----

func TestMSSQLRepository_PurgeExpired_Success(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newIdempotencyRepo(t)

	result := devkitmocks.NewMockResult(t)
	result.On("RowsAffected").Return(int64(3), nil)
	// ctx + query + now = 3 total
	db.On("ExecContext", anyArgs(3)...).Return(result, nil)

	n, err := repo.PurgeExpired(ctx, time.Now().UTC())
	require.NoError(t, err)
	assert.Equal(t, int64(3), n)
}

func TestMSSQLRepository_PurgeExpired_DBError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newIdempotencyRepo(t)

	db.On("ExecContext", anyArgs(3)...).Return(nil, errors.New("lock timeout"))

	_, err := repo.PurgeExpired(ctx, time.Now().UTC())
	require.Error(t, err)
	assert.ErrorContains(t, err, "mssql: purge expired idempotency keys")
}

func TestMSSQLRepository_PurgeExpired_RowsAffectedError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo, db := newIdempotencyRepo(t)

	result := devkitmocks.NewMockResult(t)
	result.On("RowsAffected").Return(int64(0), errors.New("driver err"))
	db.On("ExecContext", anyArgs(3)...).Return(result, nil)

	_, err := repo.PurgeExpired(ctx, time.Now().UTC())
	require.Error(t, err)
	assert.ErrorContains(t, err, "mssql: purge expired rows affected")
}
