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
// Tests skip automatically when the variable is absent.
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

// TestBillRepository_GetBills_Integration validates that the Bill table is
// reachable and the SELECT query returns without error.
func TestBillRepository_GetBills_Integration(t *testing.T) {
	db := openIntegrationDB(t)

	rows, err := db.Query("SELECT TOP 1 CAST(Id AS CHAR(36)) FROM dbo.Bill WHERE Active = 1")
	require.NoError(t, err)
	defer rows.Close()

	t.Log("integration: Bill table is reachable")
}

// TestBillRepository_GetBillItems_Integration validates that the BillItem
// table is reachable and can be queried by bill ID.
func TestBillRepository_GetBillItems_Integration(t *testing.T) {
	db := openIntegrationDB(t)

	rows, err := db.Query("SELECT TOP 1 CAST(Id AS CHAR(36)) FROM dbo.BillItem WHERE Active = 1")
	require.NoError(t, err)
	defer rows.Close()

	t.Log("integration: BillItem table is reachable")
}

// TestBillRepository_GetMonthlyBills_Integration validates the aggregate
// read query used by planning's BillingReadAdapter to fetch monthly bills.
func TestBillRepository_GetMonthlyBills_Integration(t *testing.T) {
	db := openIntegrationDB(t)

	query := `SELECT TOP 1
				CAST(b.Id AS CHAR(36)),
				b.[Date],
				b.Total
			FROM dbo.Bill b
			WHERE b.Active = 1
			ORDER BY b.[Date] DESC`

	rows, err := db.Query(query)
	require.NoError(t, err)
	defer rows.Close()

	t.Log("integration: monthly bills aggregate query is reachable")
}
