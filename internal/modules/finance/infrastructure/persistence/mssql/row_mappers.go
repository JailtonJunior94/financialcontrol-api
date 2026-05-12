package mssql

import (
	"time"

	"github.com/shopspring/decimal"

	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// TransactionRow holds columns scanned from dbo.FinanceTransactions.
type TransactionRow struct {
	ID                    string
	UserID                string
	Description           string
	Amount                string
	Currency              string
	OccurredAt            time.Time
	TransactionType       string
	PaymentMethod         string
	CardID                *string
	CategoryID            string
	SubcategoryID         *string
	OriginalTransactionID *string
	LegacyOrigin          *string
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             *time.Time
}

// InvoiceRow holds columns scanned from dbo.FinanceInvoices.
type InvoiceRow struct {
	ID           string
	UserID       string
	CardID       string
	State        string
	CycleStart   time.Time
	CycleEnd     time.Time
	ClosingDate  time.Time
	DueDate      time.Time
	Total        string
	Currency     string
	PaidAt       *time.Time
	LegacyOrigin *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// InstallmentRow holds columns scanned from dbo.FinanceInstallments.
type InstallmentRow struct {
	ID            string
	TransactionID string
	InvoiceID     string
	Number        int
	Total         int
	Amount        string
	Currency      string
	Status        string
	LegacyOrigin  *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// RowToTransaction maps a TransactionRow into a domain Transaction via RehydrateTransaction.
func RowToTransaction(r *TransactionRow) (*entities.Transaction, error) {
	id, err := vos.ParseTransactionID(r.ID)
	if err != nil {
		return nil, err
	}
	userID, err := identityvo.ParseUserID(r.UserID)
	if err != nil {
		return nil, err
	}
	d, err := decimal.NewFromString(r.Amount)
	if err != nil {
		return nil, domain.ErrInvalidMoneyFormat
	}
	amountMoney := vos.NewMoneyFromDecimal(d)
	amount, err := vos.NewAmountFromMoney(amountMoney)
	if err != nil {
		return nil, err
	}
	txType, err := vos.ParseTransactionType(r.TransactionType)
	if err != nil {
		return nil, err
	}
	pm, err := vos.ParsePaymentMethod(r.PaymentMethod)
	if err != nil {
		return nil, err
	}
	catID, err := vos.ParseCategoryID(r.CategoryID)
	if err != nil {
		return nil, err
	}

	var cardID *vos.CardID
	if r.CardID != nil {
		c, cerr := vos.ParseCardID(*r.CardID)
		if cerr != nil {
			return nil, cerr
		}
		cardID = &c
	}

	var subCatID *vos.CategoryID
	if r.SubcategoryID != nil {
		s, serr := vos.ParseCategoryID(*r.SubcategoryID)
		if serr != nil {
			return nil, serr
		}
		subCatID = &s
	}

	var origTxID *vos.TransactionID
	if r.OriginalTransactionID != nil {
		o, oerr := vos.ParseTransactionID(*r.OriginalTransactionID)
		if oerr != nil {
			return nil, oerr
		}
		origTxID = &o
	}

	return entities.RehydrateTransaction(
		id, userID, r.Description, amount, r.OccurredAt,
		txType, pm, cardID, catID, subCatID, origTxID,
		r.LegacyOrigin, r.CreatedAt, r.UpdatedAt, r.DeletedAt,
	), nil
}

// RowToInvoice maps an InvoiceRow into a domain Invoice via RehydrateInvoice.
func RowToInvoice(r *InvoiceRow) (*entities.Invoice, error) {
	id, err := vos.ParseInvoiceID(r.ID)
	if err != nil {
		return nil, err
	}
	userID, err := identityvo.ParseUserID(r.UserID)
	if err != nil {
		return nil, err
	}
	cardID, err := vos.ParseCardID(r.CardID)
	if err != nil {
		return nil, err
	}
	state, err := vos.ParseInvoiceState(r.State)
	if err != nil {
		return nil, err
	}
	d, err := decimal.NewFromString(r.Total)
	if err != nil {
		return nil, domain.ErrInvalidMoneyFormat
	}
	total := vos.NewMoneyFromDecimal(d)

	return entities.RehydrateInvoice(
		id, userID, cardID, state,
		r.CycleStart, r.CycleEnd, r.ClosingDate, r.DueDate,
		total, r.PaidAt, r.LegacyOrigin,
		r.CreatedAt, r.UpdatedAt, r.DeletedAt,
	), nil
}

// RowToInstallment maps an InstallmentRow into a domain Installment via RehydrateInstallment.
func RowToInstallment(r *InstallmentRow) (*entities.Installment, error) {
	id, err := vos.ParseInstallmentID(r.ID)
	if err != nil {
		return nil, err
	}
	txID, err := vos.ParseTransactionID(r.TransactionID)
	if err != nil {
		return nil, err
	}
	invID, err := vos.ParseInvoiceID(r.InvoiceID)
	if err != nil {
		return nil, err
	}
	number, err := vos.NewInstallmentNumber(r.Number)
	if err != nil {
		return nil, err
	}
	total, err := vos.NewInstallmentCount(r.Total)
	if err != nil {
		return nil, err
	}
	d, err := decimal.NewFromString(r.Amount)
	if err != nil {
		return nil, domain.ErrInvalidMoneyFormat
	}
	amount := vos.NewMoneyFromDecimal(d)
	status, err := vos.ParseInstallmentStatus(r.Status)
	if err != nil {
		return nil, err
	}

	return entities.RehydrateInstallment(
		id, txID, invID, number, total, amount,
		status, r.LegacyOrigin,
		r.CreatedAt, r.UpdatedAt, r.DeletedAt,
	), nil
}
