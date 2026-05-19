package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/filters"
	portmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/projections"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	idempmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/idempotency/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// newRefundTransaction creates a refund transaction pointing to originalID.
func newRefundTransaction(t *testing.T, userID identityvo.UserID, originalID vos.TransactionID) *entities.Transaction {
	t.Helper()
	catID := vos.NewCategoryID()
	a, err := vos.NewMoney("50.00")
	require.NoError(t, err)
	amount, err := vos.NewAmountFromMoney(a)
	require.NoError(t, err)
	oid := originalID
	tx, err := entities.NewTransaction(
		vos.NewTransactionID(),
		userID,
		"Estorno",
		amount,
		fixedNow,
		vos.TransactionTypeRefund,
		vos.PaymentMethodPix,
		nil,
		catID,
		nil,
		&oid,
		&fixedClockImpl{fixedNow},
	)
	require.NoError(t, err)
	return tx
}

// --- RefundTransaction additional coverage ---

func TestRefundTransaction_RefundOfRefund_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	originalID := vos.NewTransactionID()
	txID := vos.NewTransactionID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	refundTx := newRefundTransaction(t, userID, originalID)
	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(refundTx, nil)
	txRepo.EXPECT().HasActiveRefundFor(mock.Anything, userID, txID).Return(false, nil)

	uc := newRefundTransactionUC(t, mgr, txRepo, clock, ids)
	_, err := uc.Execute(ctx, userID, txID, dtos.RefundTransactionRequest{})

	assert.ErrorIs(t, err, domain.ErrRefundOfRefundNotAllowed)
}

func TestRefundTransaction_HasActiveRefundForError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	txID := vos.NewTransactionID()
	repoErr := errors.New("db error")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	original := newExpenseTransaction(t, userID)
	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(original, nil)
	txRepo.EXPECT().HasActiveRefundFor(mock.Anything, userID, txID).Return(false, repoErr)

	uc := newRefundTransactionUC(t, mgr, txRepo, clock, ids)
	_, err := uc.Execute(ctx, userID, txID, dtos.RefundTransactionRequest{})

	assert.ErrorIs(t, err, repoErr)
}

func TestRefundTransaction_WithPaymentMethodOverride(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	txID := vos.NewTransactionID()
	refundID := vos.NewTransactionID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	original := newExpenseTransaction(t, userID)
	clock.EXPECT().Now().Return(fixedNow)
	ids.On("NewTransactionID").Return(refundID)

	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(original, nil)
	txRepo.EXPECT().HasActiveRefundFor(mock.Anything, userID, txID).Return(false, nil)
	txRepo.EXPECT().Add(mock.Anything, mock.AnythingOfType("*entities.Transaction")).Return(nil)

	uc := newRefundTransactionUC(t, mgr, txRepo, clock, ids)
	pm := "boleto"
	req := dtos.RefundTransactionRequest{PaymentMethod: &pm}
	resp, err := uc.Execute(ctx, userID, txID, req)

	require.NoError(t, err)
	assert.Equal(t, "refund", resp.TransactionType)
	assert.Equal(t, "boleto", resp.PaymentMethod)
}

func TestRefundTransaction_InvalidPaymentMethodOverride(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	txID := vos.NewTransactionID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	original := newExpenseTransaction(t, userID)
	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(original, nil)
	txRepo.EXPECT().HasActiveRefundFor(mock.Anything, userID, txID).Return(false, nil)

	uc := newRefundTransactionUC(t, mgr, txRepo, clock, ids)
	invalid := "invalid_method"
	req := dtos.RefundTransactionRequest{PaymentMethod: &invalid}
	_, err := uc.Execute(ctx, userID, txID, req)

	assert.ErrorIs(t, err, domain.ErrInvalidPaymentMethod)
}

// --- UpdateTransaction additional coverage ---

func TestUpdateTransaction_CardRequiredButNil_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	txID := vos.NewTransactionID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	tx := newExpenseTransaction(t, userID)
	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(tx, nil)
	instRepo.EXPECT().ListByTransaction(mock.Anything, txID).Return([]entities.Installment{}, nil)

	uc := newUpdateTransactionUC(t, mgr, txRepo, invRepo, instRepo, cards, cats, clock, ids)
	req := dtos.UpdateTransactionRequest{
		Description:     "test",
		Amount:          "10.00",
		OccurredAt:      fixedNow,
		TransactionType: "credit_purchase",
		PaymentMethod:   "credit_card",
		CardID:          nil,
		CategoryID:      catID.String(),
	}
	_, err := uc.Execute(ctx, userID, txID, req)

	assert.ErrorIs(t, err, domain.ErrCardRequiredForPaymentMethod)
}

func TestUpdateTransaction_CardNotActive_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	cardID := vos.NewCardID()
	txID := vos.NewTransactionID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	tx := newExpenseTransaction(t, userID)
	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(tx, nil)
	instRepo.EXPECT().ListByTransaction(mock.Anything, txID).Return([]entities.Installment{}, nil)

	inactiveCard := newActiveCardView(userID, cardID)
	inactiveCard.Active = false
	cards.EXPECT().GetByID(mock.Anything, userID, cardID).Return(inactiveCard, nil)

	uc := newUpdateTransactionUC(t, mgr, txRepo, invRepo, instRepo, cards, cats, clock, ids)
	cid := cardID.String()
	req := dtos.UpdateTransactionRequest{
		Description:     "test",
		Amount:          "10.00",
		OccurredAt:      fixedNow,
		TransactionType: "credit_purchase",
		PaymentMethod:   "credit_card",
		CardID:          &cid,
		CategoryID:      catID.String(),
	}
	_, err := uc.Execute(ctx, userID, txID, req)

	assert.ErrorIs(t, err, domain.ErrCardNotActive)
}

func TestUpdateTransaction_SubcategoryNotFound_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	subcatID := vos.NewCategoryID()
	txID := vos.NewTransactionID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	tx := newExpenseTransaction(t, userID)
	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(tx, nil)
	instRepo.EXPECT().ListByTransaction(mock.Anything, txID).Return([]entities.Installment{}, nil)
	cats.EXPECT().GetByID(mock.Anything, userID, catID).Return(newActiveCategoryView(userID, catID), nil)
	cats.EXPECT().GetByID(mock.Anything, userID, subcatID).Return(projections.CategoryView{}, errors.New("not found"))

	uc := newUpdateTransactionUC(t, mgr, txRepo, invRepo, instRepo, cards, cats, clock, ids)
	scid := subcatID.String()
	req := dtos.UpdateTransactionRequest{
		Description:     "test",
		Amount:          "10.00",
		OccurredAt:      fixedNow,
		TransactionType: "expense",
		PaymentMethod:   "pix",
		CategoryID:      catID.String(),
		SubcategoryID:   &scid,
	}
	_, err := uc.Execute(ctx, userID, txID, req)

	assert.ErrorIs(t, err, domain.ErrSubcategoryNotFound)
}

func TestUpdateTransaction_SubcategoryNotActive_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	subcatID := vos.NewCategoryID()
	txID := vos.NewTransactionID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	tx := newExpenseTransaction(t, userID)
	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(tx, nil)
	instRepo.EXPECT().ListByTransaction(mock.Anything, txID).Return([]entities.Installment{}, nil)
	cats.EXPECT().GetByID(mock.Anything, userID, catID).Return(newActiveCategoryView(userID, catID), nil)

	inactiveSub := newActiveCategoryView(userID, subcatID)
	inactiveSub.Active = false
	inactiveSub.ParentID = &catID
	cats.EXPECT().GetByID(mock.Anything, userID, subcatID).Return(inactiveSub, nil)

	uc := newUpdateTransactionUC(t, mgr, txRepo, invRepo, instRepo, cards, cats, clock, ids)
	scid := subcatID.String()
	req := dtos.UpdateTransactionRequest{
		Description:     "test",
		Amount:          "10.00",
		OccurredAt:      fixedNow,
		TransactionType: "expense",
		PaymentMethod:   "pix",
		CategoryID:      catID.String(),
		SubcategoryID:   &scid,
	}
	_, err := uc.Execute(ctx, userID, txID, req)

	assert.ErrorIs(t, err, domain.ErrSubcategoryNotActive)
}

func TestUpdateTransaction_SubcategoryNotChild_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	subcatID := vos.NewCategoryID()
	otherCatID := vos.NewCategoryID()
	txID := vos.NewTransactionID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	tx := newExpenseTransaction(t, userID)
	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(tx, nil)
	instRepo.EXPECT().ListByTransaction(mock.Anything, txID).Return([]entities.Installment{}, nil)
	cats.EXPECT().GetByID(mock.Anything, userID, catID).Return(newActiveCategoryView(userID, catID), nil)

	subView := newActiveCategoryView(userID, subcatID)
	subView.ParentID = &otherCatID
	cats.EXPECT().GetByID(mock.Anything, userID, subcatID).Return(subView, nil)

	uc := newUpdateTransactionUC(t, mgr, txRepo, invRepo, instRepo, cards, cats, clock, ids)
	scid := subcatID.String()
	req := dtos.UpdateTransactionRequest{
		Description:     "test",
		Amount:          "10.00",
		OccurredAt:      fixedNow,
		TransactionType: "expense",
		PaymentMethod:   "pix",
		CategoryID:      catID.String(),
		SubcategoryID:   &scid,
	}
	_, err := uc.Execute(ctx, userID, txID, req)

	assert.ErrorIs(t, err, domain.ErrSubcategoryNotChildOfCategory)
}

func TestUpdateTransaction_ListByTransactionError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	txID := vos.NewTransactionID()
	repoErr := errors.New("list installments error")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	tx := newExpenseTransaction(t, userID)
	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(tx, nil)
	instRepo.EXPECT().ListByTransaction(mock.Anything, txID).Return(nil, repoErr)

	uc := newUpdateTransactionUC(t, mgr, txRepo, invRepo, instRepo, cards, cats, clock, ids)
	req := dtos.UpdateTransactionRequest{
		Description:     "test",
		Amount:          "10.00",
		OccurredAt:      fixedNow,
		TransactionType: "expense",
		PaymentMethod:   "pix",
		CategoryID:      catID.String(),
	}
	_, err := uc.Execute(ctx, userID, txID, req)

	assert.ErrorIs(t, err, repoErr)
}

// --- CreateTransaction additional coverage ---

func TestCreateTransaction_SubcategoryNotFound_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	subcatID := vos.NewCategoryID()
	key, _ := vos.NewIdempotencyKey("test-key-subcat-notfound")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	idempRepo := idempmocks.NewIdempotencyRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	idempRepo.EXPECT().Get(mock.Anything, userID, "POST /finance/transactions", key.Value()).Return(nil, nil)
	cats.EXPECT().GetByID(mock.Anything, userID, catID).Return(newActiveCategoryView(userID, catID), nil)
	cats.EXPECT().GetByID(mock.Anything, userID, subcatID).Return(projections.CategoryView{}, domain.ErrCategoryNotFound)

	uc := newCreateTransactionUC(t, mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, clock, ids)
	scid := subcatID.String()
	req := dtos.CreateTransactionRequest{
		Description:     "test",
		Amount:          "10.00",
		OccurredAt:      fixedNow,
		TransactionType: "expense",
		PaymentMethod:   "pix",
		CategoryID:      catID.String(),
		SubcategoryID:   &scid,
	}
	_, err := uc.Execute(ctx, userID, key, req)

	assert.ErrorIs(t, err, domain.ErrSubcategoryNotFound)
}

func TestCreateTransaction_SubcategoryNotActive_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	subcatID := vos.NewCategoryID()
	key, _ := vos.NewIdempotencyKey("test-key-subcat-inactive")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	idempRepo := idempmocks.NewIdempotencyRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	idempRepo.EXPECT().Get(mock.Anything, userID, "POST /finance/transactions", key.Value()).Return(nil, nil)
	cats.EXPECT().GetByID(mock.Anything, userID, catID).Return(newActiveCategoryView(userID, catID), nil)

	inactiveSub := newActiveCategoryView(userID, subcatID)
	inactiveSub.Active = false
	inactiveSub.ParentID = &catID
	cats.EXPECT().GetByID(mock.Anything, userID, subcatID).Return(inactiveSub, nil)

	uc := newCreateTransactionUC(t, mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, clock, ids)
	scid := subcatID.String()
	req := dtos.CreateTransactionRequest{
		Description:     "test",
		Amount:          "10.00",
		OccurredAt:      fixedNow,
		TransactionType: "expense",
		PaymentMethod:   "pix",
		CategoryID:      catID.String(),
		SubcategoryID:   &scid,
	}
	_, err := uc.Execute(ctx, userID, key, req)

	assert.ErrorIs(t, err, domain.ErrSubcategoryNotActive)
}

func TestCreateTransaction_SubcategoryNotChild_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	subcatID := vos.NewCategoryID()
	otherCatID := vos.NewCategoryID()
	key, _ := vos.NewIdempotencyKey("test-key-subcat-notchild")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	idempRepo := idempmocks.NewIdempotencyRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	idempRepo.EXPECT().Get(mock.Anything, userID, "POST /finance/transactions", key.Value()).Return(nil, nil)
	cats.EXPECT().GetByID(mock.Anything, userID, catID).Return(newActiveCategoryView(userID, catID), nil)

	subView := newActiveCategoryView(userID, subcatID)
	subView.ParentID = &otherCatID
	cats.EXPECT().GetByID(mock.Anything, userID, subcatID).Return(subView, nil)

	uc := newCreateTransactionUC(t, mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, clock, ids)
	scid := subcatID.String()
	req := dtos.CreateTransactionRequest{
		Description:     "test",
		Amount:          "10.00",
		OccurredAt:      fixedNow,
		TransactionType: "expense",
		PaymentMethod:   "pix",
		CategoryID:      catID.String(),
		SubcategoryID:   &scid,
	}
	_, err := uc.Execute(ctx, userID, key, req)

	assert.ErrorIs(t, err, domain.ErrSubcategoryNotChildOfCategory)
}

// --- ListTransactions additional coverage ---

func TestListTransactions_InvRepoListError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	repoErr := errors.New("invoice list error")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	clock := portmocks.NewClock(t)

	clock.EXPECT().Now().Return(fixedNow)
	invRepo.EXPECT().List(mock.Anything, userID, mock.AnythingOfType("filters.InvoiceFilter")).Return([]entities.Invoice{}, int64(0), repoErr)

	pag, _ := vos.NewPagination(1, 10)
	f, err := filters.NewTransactionFilter(userID, nil, nil, nil, nil, vos.InvoiceStatusFilterNone, "", nil, nil, pag)
	require.NoError(t, err)

	uc := newListTransactionsUC(t, mgr, txRepo, invRepo, clock)
	_, execErr := uc.Execute(ctx, userID, f)

	assert.ErrorIs(t, execErr, repoErr)
}

func TestListTransactions_LazyClose_UpdateError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	cardID := vos.NewCardID()
	updateErr := errors.New("update invoice error")

	tx := &mockTx{}
	mgr := &mockManager{}
	mgr.On("BeginTx", mock.Anything, mock.Anything).Return(tx, nil)
	tx.On("Rollback", mock.Anything).Return(nil)

	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	clock := portmocks.NewClock(t)

	card := newActiveCardView(userID, cardID)
	pastClosing := fixedNow.AddDate(0, -1, -5)
	pastEnd := fixedNow.AddDate(0, -1, 0)
	pastStart := fixedNow.AddDate(0, -2, 0)
	dueDate := fixedNow.AddDate(0, -1, 10)
	openInv, err := entities.NewInvoice(card, pastStart, pastEnd, pastClosing, dueDate, fixedNow.AddDate(0, -2, 0))
	require.NoError(t, err)

	clock.EXPECT().Now().Return(fixedNow)
	invRepo.EXPECT().List(mock.Anything, userID, mock.AnythingOfType("filters.InvoiceFilter")).Return([]entities.Invoice{*openInv}, int64(1), nil)
	invRepo.EXPECT().Update(mock.Anything, mock.AnythingOfType("*entities.Invoice")).Return(updateErr)

	pag, _ := vos.NewPagination(1, 10)
	f, ferr := filters.NewTransactionFilter(userID, nil, nil, nil, nil, vos.InvoiceStatusFilterNone, "", nil, nil, pag)
	require.NoError(t, ferr)

	closer := services.NewInvoiceCloser()
	uc := usecase.NewListTransactions(mgr, txRepo, invRepo, closer, clock)
	_, execErr := uc.Execute(ctx, userID, f)

	assert.ErrorIs(t, execErr, updateErr)
}

// --- AnticipateInstallment additional coverage ---

func TestAnticipateInstallment_ListByTransactionError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	cardID := vos.NewCardID()
	txID := vos.NewTransactionID()
	instID := vos.NewInstallmentID()
	repoErr := errors.New("list installments error")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	clock := portmocks.NewClock(t)

	tx := newCreditPurchaseTransaction(t, userID, cardID)
	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(tx, nil)
	instRepo.EXPECT().ListByTransaction(mock.Anything, txID).Return(nil, repoErr)

	uc := newAnticipateInstallmentUC(t, mgr, txRepo, invRepo, instRepo, clock)
	_, err := uc.Execute(ctx, userID, txID, instID)

	assert.ErrorIs(t, err, repoErr)
}

// --- CreateTransaction: idempotency get error and card-required tests ---

func TestCreateTransaction_IdempotencyGetError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	key, _ := vos.NewIdempotencyKey("test-key-idemp-err")
	repoErr := errors.New("idemp get error")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	idempRepo := idempmocks.NewIdempotencyRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	idempRepo.EXPECT().Get(mock.Anything, userID, "POST /finance/transactions", key.Value()).Return(nil, repoErr)

	uc := newCreateTransactionUC(t, mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, clock, ids)
	req := validCreateRequest(catID.String())
	_, err := uc.Execute(ctx, userID, key, req)

	assert.ErrorIs(t, err, repoErr)
}

func TestCreateTransaction_CardRequiredButNil_ReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	key, _ := vos.NewIdempotencyKey("test-key-card-nil")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	idempRepo := idempmocks.NewIdempotencyRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	idempRepo.EXPECT().Get(mock.Anything, userID, "POST /finance/transactions", key.Value()).Return(nil, nil)

	uc := newCreateTransactionUC(t, mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, clock, ids)
	req := dtos.CreateTransactionRequest{
		Description:     "Credit purchase",
		Amount:          "100.00",
		OccurredAt:      fixedNow,
		TransactionType: "credit_purchase",
		PaymentMethod:   "credit_card",
		CardID:          nil,
		CategoryID:      catID.String(),
	}
	_, err := uc.Execute(ctx, userID, key, req)

	assert.ErrorIs(t, err, domain.ErrCardRequiredForPaymentMethod)
}

// --- DeleteTransaction additional coverage ---

func TestDeleteTransaction_HasActiveRefundForError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	txID := vos.NewTransactionID()
	repoErr := errors.New("has active refund error")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	clock := portmocks.NewClock(t)

	tx := newExpenseTransaction(t, userID)
	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(tx, nil)
	txRepo.EXPECT().HasActiveRefundFor(mock.Anything, userID, txID).Return(false, repoErr)

	uc := newDeleteTransactionUC(t, mgr, txRepo, instRepo, clock)
	err := uc.Execute(ctx, userID, txID)

	assert.ErrorIs(t, err, repoErr)
}

func TestDeleteTransaction_HasClosedOrPaidInstallmentError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	txID := vos.NewTransactionID()
	repoErr := errors.New("has closed or paid error")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	clock := portmocks.NewClock(t)

	tx := newExpenseTransaction(t, userID)
	txRepo.EXPECT().GetByID(mock.Anything, userID, txID).Return(tx, nil)
	txRepo.EXPECT().HasActiveRefundFor(mock.Anything, userID, txID).Return(false, nil)
	instRepo.EXPECT().HasClosedOrPaidForTransaction(mock.Anything, txID).Return(false, repoErr)

	uc := newDeleteTransactionUC(t, mgr, txRepo, instRepo, clock)
	err := uc.Execute(ctx, userID, txID)

	assert.ErrorIs(t, err, repoErr)
}
