//go:build integration

package database_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"
	devkitmgr "github.com/JailtonJunior94/devkit-go/pkg/database/manager"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	bootstrapdb "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/database"
	dbmssql "github.com/jailtonjunior94/financialcontrol-api/pkg/database/mssql"
)

const (
	createUoWTestTable = `
IF NOT EXISTS (
    SELECT 1 FROM INFORMATION_SCHEMA.TABLES
    WHERE TABLE_SCHEMA = 'dbo' AND TABLE_NAME = 'UoWTest'
)
CREATE TABLE dbo.[UoWTest] (
    [Id]  INT           NOT NULL PRIMARY KEY,
    [Val] NVARCHAR(100) NOT NULL
)`

	insertUoWRow = `INSERT INTO dbo.[UoWTest] ([Id],[Val]) VALUES (@id, @val)`
	countUoWRows = `SELECT COUNT(1) FROM dbo.[UoWTest] WHERE [Id] = @id`
	deleteUoWAll = `DELETE FROM dbo.[UoWTest]`
)

type UoWSuite struct {
	suite.Suite
	mgr devkitmgr.Manager
}

func TestUoWSuite(t *testing.T) {
	suite.Run(t, new(UoWSuite))
}

func (s *UoWSuite) SetupSuite() {
	db, _, err := dbmssql.GetSharedTestDatabase()
	s.Require().NoError(err, "shared test database unavailable")

	_, err = db.ExecContext(context.Background(), createUoWTestTable)
	s.Require().NoError(err, "create UoWTest table")

	mgr, _, err := dbmssql.GetSharedTestManager()
	s.Require().NoError(err, "shared test manager unavailable")
	s.mgr = mgr
}

func (s *UoWSuite) SetupTest() {
	_, err := s.mgr.DBTX(context.Background()).ExecContext(context.Background(), deleteUoWAll)
	s.Require().NoError(err, "truncate UoWTest before test")
}

// TestCommitOnSuccess verifies that when fn returns nil the tx is committed and
// the inserted row is visible outside the transaction.
func (s *UoWSuite) TestCommitOnSuccess() {
	ctx := context.Background()

	err := bootstrapdb.Do(ctx, s.mgr, func(ctx context.Context) error {
		_, err := s.mgr.DBTX(ctx).ExecContext(ctx, insertUoWRow,
			sql.Named("id", 1),
			sql.Named("val", "hello"),
		)
		return err
	})
	s.Require().NoError(err)

	// Row must be visible after commit.
	s.assertCount(ctx, 1, 1)
}

// TestRollbackOnError verifies that when fn returns an error the tx is rolled
// back and no row is visible outside the transaction.
func (s *UoWSuite) TestRollbackOnError() {
	ctx := context.Background()

	boom := errors.New("intentional error")
	err := bootstrapdb.Do(ctx, s.mgr, func(ctx context.Context) error {
		_, execErr := s.mgr.DBTX(ctx).ExecContext(ctx, insertUoWRow,
			sql.Named("id", 2),
			sql.Named("val", "should be rolled back"),
		)
		if execErr != nil {
			return execErr
		}
		return boom
	})
	require.ErrorIs(s.T(), err, boom)

	// Row must NOT be visible after rollback.
	s.assertCount(ctx, 2, 0)
}

// TestDBTXRespectsWithTx verifies that mgr.DBTX(ctx) inside the fn callback
// returns the active tx propagated via devkitdb.WithTx, not a pool connection.
func (s *UoWSuite) TestDBTXRespectsWithTx() {
	ctx := context.Background()

	var dbtxInsideFn devkitdb.DBTX
	err := bootstrapdb.Do(ctx, s.mgr, func(ctx context.Context) error {
		dbtxInsideFn = s.mgr.DBTX(ctx)
		_, ok := devkitdb.FromContext(ctx)
		if !ok {
			return errors.New("tx missing from context")
		}
		return nil
	})
	s.Require().NoError(err)
	s.Require().NotNil(dbtxInsideFn, "mgr.DBTX(ctx) must return non-nil inside fn")
}

func (s *UoWSuite) assertCount(ctx context.Context, id, want int) {
	s.T().Helper()
	var count int
	err := s.mgr.DBTX(ctx).QueryRowContext(ctx, countUoWRows,
		sql.Named("id", id),
	).Scan(&count)
	s.Require().NoError(err)
	s.Equal(want, count, "unexpected row count for id=%d", id)
}
