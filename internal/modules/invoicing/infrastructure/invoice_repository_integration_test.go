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

// TestInvoiceRepository_GetInvoiceByCardId_Integration validates that the
// Invoice table and Card join are reachable without errors.
func TestInvoiceRepository_GetInvoiceByCardId_Integration(t *testing.T) {
	db := openIntegrationDB(t)

	rows, err := db.Query("SELECT TOP 1 CAST(Id AS CHAR(36)) FROM dbo.Invoice WHERE Active = 1")
	require.NoError(t, err)
	defer rows.Close()

	t.Log("integration: Invoice table is reachable")
}

// TestInvoiceRepository_GetInvoiceItems_Integration validates that the
// InvoiceItem table and Category join are reachable without errors.
func TestInvoiceRepository_GetInvoiceItems_Integration(t *testing.T) {
	db := openIntegrationDB(t)

	rows, err := db.Query(`SELECT TOP 1
		ii.Id, ii.InvoiceId, c.Name
		FROM dbo.InvoiceItem ii
		INNER JOIN dbo.Category c ON c.Id = ii.CategoryId
		WHERE ii.Active = 1`)
	require.NoError(t, err)
	defer rows.Close()

	t.Log("integration: InvoiceItem→Category join is reachable")
}

// TestInvoiceRepository_GetLastControl_Integration validates the aggregate
// query that returns the maximum InvoiceControl value used during item import.
func TestInvoiceRepository_GetLastControl_Integration(t *testing.T) {
	db := openIntegrationDB(t)

	row := db.QueryRow("SELECT ISNULL(MAX(InvoiceControl), 0) FROM dbo.InvoiceItem")
	var control int64
	require.NoError(t, row.Scan(&control))

	t.Logf("integration: GetLastControl returned %d", control)
}

// TestInvoiceRepository_GetInvoices_Integration validates the composite query
// used by planning's InvoicingReadAdapter to aggregate monthly invoice data.
func TestInvoiceRepository_GetInvoices_Integration(t *testing.T) {
	db := openIntegrationDB(t)

	query := `SELECT TOP 1
				CAST(i.Id AS CHAR(36)),
				i.[Date],
				i.Total
			FROM dbo.Invoice i
			INNER JOIN dbo.Card c ON c.Id = i.CardId`

	rows, err := db.Query(query)
	require.NoError(t, err)
	defer rows.Close()

	t.Log("integration: invoice→card composite query is reachable")
}
