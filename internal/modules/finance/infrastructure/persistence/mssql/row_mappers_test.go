package mssql_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/persistence/mssql"
)

var (
	validTxID   = "a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d"
	validUserID = "b2c3d4e5-f6a7-4b8c-9d0e-1f2a3b4c5d6e"
	validCardID = "c3d4e5f6-a7b8-4c9d-0e1f-2a3b4c5d6e7f"
	validCatID  = "d4e5f6a7-b8c9-4d0e-1f2a-3b4c5d6e7f8a"
	validInvID  = "e5f6a7b8-c9d0-4e1f-2a3b-4c5d6e7f8a9b"
	validInstID = "f6a7b8c9-d0e1-4f2a-3b4c-5d6e7f8a9b0c"
	now         = time.Now().UTC()
)

// TestRowToTransaction_ValidRow tests the golden path mapper.
func TestRowToTransaction_ValidRow(t *testing.T) {
	t.Parallel()
	row := &mssql.TransactionRow{
		ID:              validTxID,
		UserID:          validUserID,
		Description:     "Test transaction",
		Amount:          "150.0000",
		Currency:        "BRL",
		OccurredAt:      now,
		TransactionType: "expense",
		PaymentMethod:   "pix",
		CategoryID:      validCatID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	tx, err := mssql.RowToTransaction(row)
	require.NoError(t, err)
	assert.Equal(t, validTxID, tx.ID().String())
	assert.Equal(t, "Test transaction", tx.Description())
	assert.Nil(t, tx.CardID())
	assert.Nil(t, tx.DeletedAt())
}

func TestRowToTransaction_WithCardAndSubcategory(t *testing.T) {
	t.Parallel()
	subCatID := validInvID
	row := &mssql.TransactionRow{
		ID:              validTxID,
		UserID:          validUserID,
		Description:     "Credit buy",
		Amount:          "200.0000",
		Currency:        "BRL",
		OccurredAt:      now,
		TransactionType: "credit_purchase",
		PaymentMethod:   "credit_card",
		CardID:          &validCardID,
		CategoryID:      validCatID,
		SubcategoryID:   &subCatID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	tx, err := mssql.RowToTransaction(row)
	require.NoError(t, err)
	require.NotNil(t, tx.CardID())
	assert.Equal(t, validCardID, tx.CardID().String())
	require.NotNil(t, tx.SubcategoryID())
}

func TestRowToTransaction_InvalidUserID(t *testing.T) {
	t.Parallel()
	row := &mssql.TransactionRow{
		ID:              validTxID,
		UserID:          "not-a-uuid",
		Description:     "x",
		Amount:          "1.0000",
		Currency:        "BRL",
		OccurredAt:      now,
		TransactionType: "income",
		PaymentMethod:   "pix",
		CategoryID:      validCatID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	_, err := mssql.RowToTransaction(row)
	require.Error(t, err)
}

func TestRowToTransaction_InvalidAmount(t *testing.T) {
	t.Parallel()
	row := &mssql.TransactionRow{
		ID:              validTxID,
		UserID:          validUserID,
		Description:     "x",
		Amount:          "not-a-decimal",
		Currency:        "BRL",
		OccurredAt:      now,
		TransactionType: "income",
		PaymentMethod:   "pix",
		CategoryID:      validCatID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	_, err := mssql.RowToTransaction(row)
	require.Error(t, err)
}

func TestRowToTransaction_InvalidTransactionType(t *testing.T) {
	t.Parallel()
	row := &mssql.TransactionRow{
		ID:              validTxID,
		UserID:          validUserID,
		Description:     "x",
		Amount:          "1.0000",
		Currency:        "BRL",
		OccurredAt:      now,
		TransactionType: "unknown_type",
		PaymentMethod:   "pix",
		CategoryID:      validCatID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	_, err := mssql.RowToTransaction(row)
	require.Error(t, err)
}

func TestRowToTransaction_ZeroAmount(t *testing.T) {
	t.Parallel()
	row := &mssql.TransactionRow{
		ID:              validTxID,
		UserID:          validUserID,
		Description:     "x",
		Amount:          "0.0000",
		Currency:        "BRL",
		OccurredAt:      now,
		TransactionType: "income",
		PaymentMethod:   "pix",
		CategoryID:      validCatID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	_, err := mssql.RowToTransaction(row)
	require.Error(t, err)
}

func TestRowToInvoice_ValidRow(t *testing.T) {
	t.Parallel()
	row := &mssql.InvoiceRow{
		ID:          validInvID,
		UserID:      validUserID,
		CardID:      validCardID,
		State:       "open",
		CycleStart:  now.AddDate(0, -1, 0),
		CycleEnd:    now,
		ClosingDate: now,
		DueDate:     now.AddDate(0, 0, 10),
		Total:       "0.0000",
		Currency:    "BRL",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	inv, err := mssql.RowToInvoice(row)
	require.NoError(t, err)
	assert.Equal(t, validInvID, inv.ID().String())
	assert.True(t, inv.IsOpen())
}

func TestRowToInvoice_InvalidState(t *testing.T) {
	t.Parallel()
	row := &mssql.InvoiceRow{
		ID:          validInvID,
		UserID:      validUserID,
		CardID:      validCardID,
		State:       "unknown",
		CycleStart:  now,
		CycleEnd:    now,
		ClosingDate: now,
		DueDate:     now,
		Total:       "0.0000",
		Currency:    "BRL",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	_, err := mssql.RowToInvoice(row)
	require.Error(t, err)
}

func TestRowToInstallment_ValidRow(t *testing.T) {
	t.Parallel()
	row := &mssql.InstallmentRow{
		ID:            validInstID,
		TransactionID: validTxID,
		InvoiceID:     validInvID,
		Number:        1,
		Total:         3,
		Amount:        "33.3400",
		Currency:      "BRL",
		Status:        "scheduled",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	inst, err := mssql.RowToInstallment(row)
	require.NoError(t, err)
	assert.Equal(t, 1, inst.Number().Value())
	assert.Equal(t, 3, inst.Total().Value())
}

func TestRowToInstallment_InvalidStatus(t *testing.T) {
	t.Parallel()
	row := &mssql.InstallmentRow{
		ID:            validInstID,
		TransactionID: validTxID,
		InvoiceID:     validInvID,
		Number:        1,
		Total:         1,
		Amount:        "10.0000",
		Currency:      "BRL",
		Status:        "bad_status",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	_, err := mssql.RowToInstallment(row)
	require.Error(t, err)
}

func TestRowToInstallment_InvalidNumber(t *testing.T) {
	t.Parallel()
	row := &mssql.InstallmentRow{
		ID:            validInstID,
		TransactionID: validTxID,
		InvoiceID:     validInvID,
		Number:        0,
		Total:         1,
		Amount:        "10.0000",
		Currency:      "BRL",
		Status:        "scheduled",
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	_, err := mssql.RowToInstallment(row)
	require.Error(t, err)
}
