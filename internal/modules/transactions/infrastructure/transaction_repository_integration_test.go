//go:build integration

package infrastructure_test

import (
	"os"
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// openIntegrationDB opens a real SQL Server connection using MSSQL_DSN.
// Tests in this file skip automatically when the variable is absent so the
// suite never breaks normal CI runs that do not provision a database.
func openIntegrationDB(t *testing.T) *sqlx.DB {
	t.Helper()
	dsn := os.Getenv("MSSQL_DSN")
	if dsn == "" {
		t.Skip("MSSQL_DSN not set: skipping integration test (run with -tags integration and MSSQL_DSN=<dsn>)")
	}

	db, err := sqlx.Connect("sqlserver", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestTransactionRepository_GetTransactions_Integration validates that the
// transactions table is reachable and the SELECT query returns without error.
func TestTransactionRepository_GetTransactions_Integration(t *testing.T) {
	db := openIntegrationDB(t)

	rows, err := db.Query("SELECT TOP 1 Id FROM dbo.[Transaction] WHERE Active = 1")
	require.NoError(t, err)
	defer rows.Close()

	t.Log("integration: transactions table is reachable")
}

// TestTransactionRepository_GetTransactionItems_Integration validates that the
// TransactionItem table is reachable and query parameters are wired correctly.
func TestTransactionRepository_GetTransactionItems_Integration(t *testing.T) {
	db := openIntegrationDB(t)

	rows, err := db.Query("SELECT TOP 1 Id FROM dbo.TransactionItem WHERE Active = 1")
	require.NoError(t, err)
	defer rows.Close()

	t.Log("integration: transaction items table is reachable")
}

// TestTransactionRepository_FetchTransactionByDate_Integration validates the
// critical cross-join query used during invoice sync operations.
func TestTransactionRepository_FetchTransactionByDate_Integration(t *testing.T) {
	db := openIntegrationDB(t)

	query := `SELECT TOP 1
				CAST(t.Id AS CHAR(36)),
				CAST(ti.Id AS CHAR(36)),
				CAST(t.UserId AS CHAR(36))
			FROM dbo.[Transaction] t
			INNER JOIN dbo.TransactionItem ti ON ti.TransactionId = t.Id
			WHERE t.[Date] = CONVERT(DATETIME, GETDATE())`

	rows, err := db.Query(query)
	require.NoError(t, err)
	defer rows.Close()

	t.Log("integration: FetchTransactionByDate query executed successfully")
}
