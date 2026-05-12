package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	portmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

func newUpdateTransactionUC(
	t *testing.T,
	mgr *mockManager,
	txRepo *portmocks.TransactionRepository,
	invRepo *portmocks.InvoiceRepository,
	instRepo *portmocks.InstallmentRepository,
	cards *portmocks.CardProvider,
	cats *portmocks.CategoryProvider,
	clock *portmocks.Clock,
	ids *portmocks.IDGenerator,
) usecase.UpdateTransaction {
	t.Helper()
	splitter := services.NewInstallmentSplitter()
	return usecase.NewUpdateTransaction(mgr, txRepo, invRepo, instRepo, cards, cats, splitter, clock, ids)
}

func TestUpdateTransaction_GoldenPath_Expense(t *testing.T) {
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
	clock.On("Now").Return(fixedNow)

	txRepo.On("GetByID", mock.Anything, userID, txID).Return(tx, nil)
	instRepo.On("ListByTransaction", mock.Anything, txID).Return([]entities.Installment{}, nil)
	cats.On("GetByID", mock.Anything, userID, catID).Return(newActiveCategoryView(userID, catID), nil)
	txRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Transaction")).Return(nil)

	uc := newUpdateTransactionUC(t, mgr, txRepo, invRepo, instRepo, cards, cats, clock, ids)
	req := dtos.UpdateTransactionRequest{
		Description:     "Updated coffee",
		Amount:          "75.00",
		OccurredAt:      fixedNow,
		TransactionType: "expense",
		PaymentMethod:   "pix",
		CategoryID:      catID.String(),
	}
	resp, err := uc.Execute(ctx, userID, txID, req)

	require.NoError(t, err)
	assert.Equal(t, "Updated coffee", resp.Description)
	assert.Equal(t, "expense", resp.TransactionType)
}

func TestUpdateTransaction_InstallmentPurchase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	cardID := vos.NewCardID()
	txID := vos.NewTransactionID()
	instID1 := vos.NewInstallmentID()
	instID2 := vos.NewInstallmentID()
	invoiceID := vos.NewInvoiceID()

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	tx := newCreditPurchaseTransaction(t, userID, cardID)
	clock.On("Now").Return(fixedNow)
	ids.On("NewInstallmentID").Return(instID1).Once()
	ids.On("NewInstallmentID").Return(instID2).Once()

	txRepo.On("GetByID", mock.Anything, userID, txID).Return(tx, nil)
	instRepo.On("ListByTransaction", mock.Anything, txID).Return([]entities.Installment{}, nil)
	cats.On("GetByID", mock.Anything, userID, catID).Return(newActiveCategoryView(userID, catID), nil)
	cards.On("GetByID", mock.Anything, userID, cardID).Return(newActiveCardView(userID, cardID), nil)

	openInv := newOpenInvoice(t, userID, cardID)
	_ = invoiceID
	invRepo.On("AssignOrCreateOpen", mock.Anything, userID, cardID, mock.AnythingOfType("time.Time"), mock.Anything, mock.Anything).Return(openInv, nil)

	txRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Transaction")).Return(nil)
	instRepo.On("AddBatch", mock.Anything, mock.Anything).Return(nil)

	uc := newUpdateTransactionUC(t, mgr, txRepo, invRepo, instRepo, cards, cats, clock, ids)
	cid := cardID.String()
	req := dtos.UpdateTransactionRequest{
		Description:      "Installment buy 2x",
		Amount:           "200.00",
		OccurredAt:       fixedNow,
		TransactionType:  "installment_purchase",
		PaymentMethod:    "credit_card",
		CardID:           &cid,
		CategoryID:       catID.String(),
		InstallmentCount: 2,
	}
	resp, err := uc.Execute(ctx, userID, txID, req)

	require.NoError(t, err)
	assert.Equal(t, "installment_purchase", resp.TransactionType)
	assert.Len(t, resp.Installments, 2)
}

func TestUpdateTransaction_WithSubcategory(t *testing.T) {
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
	clock.On("Now").Return(fixedNow)

	txRepo.On("GetByID", mock.Anything, userID, txID).Return(tx, nil)
	instRepo.On("ListByTransaction", mock.Anything, txID).Return([]entities.Installment{}, nil)

	catView := newActiveCategoryView(userID, catID)
	cats.On("GetByID", mock.Anything, userID, catID).Return(catView, nil)

	subView := newActiveCategoryView(userID, subcatID)
	subView.ParentID = &catID
	cats.On("GetByID", mock.Anything, userID, subcatID).Return(subView, nil)

	txRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Transaction")).Return(nil)

	uc := newUpdateTransactionUC(t, mgr, txRepo, invRepo, instRepo, cards, cats, clock, ids)
	scid := subcatID.String()
	req := dtos.UpdateTransactionRequest{
		Description:     "Updated grocery",
		Amount:          "50.00",
		OccurredAt:      fixedNow,
		TransactionType: "expense",
		PaymentMethod:   "pix",
		CategoryID:      catID.String(),
		SubcategoryID:   &scid,
	}
	resp, err := uc.Execute(ctx, userID, txID, req)

	require.NoError(t, err)
	assert.Equal(t, "Updated grocery", resp.Description)
}

func TestUpdateTransaction_CategoryNotActive_ReturnsError(t *testing.T) {
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

	txRepo.On("GetByID", mock.Anything, userID, txID).Return(tx, nil)
	instRepo.On("ListByTransaction", mock.Anything, txID).Return([]entities.Installment{}, nil)

	inactiveView := newActiveCategoryView(userID, catID)
	inactiveView.Active = false
	cats.On("GetByID", mock.Anything, userID, catID).Return(inactiveView, nil)

	uc := newUpdateTransactionUC(t, mgr, txRepo, invRepo, instRepo, cards, cats, clock, ids)
	req := dtos.UpdateTransactionRequest{
		Description:     "x",
		Amount:          "10.00",
		OccurredAt:      fixedNow,
		TransactionType: "expense",
		PaymentMethod:   "pix",
		CategoryID:      catID.String(),
	}
	_, err := uc.Execute(ctx, userID, txID, req)

	assert.ErrorIs(t, err, domain.ErrCategoryNotActive)
}

func TestUpdateTransaction_CreditCard_GoldenPath(t *testing.T) {
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

	tx := newCreditPurchaseTransaction(t, userID, cardID)
	clock.On("Now").Return(fixedNow)

	txRepo.On("GetByID", mock.Anything, userID, txID).Return(tx, nil)
	instRepo.On("ListByTransaction", mock.Anything, txID).Return([]entities.Installment{}, nil)
	cats.On("GetByID", mock.Anything, userID, catID).Return(newActiveCategoryView(userID, catID), nil)
	cards.On("GetByID", mock.Anything, userID, cardID).Return(newActiveCardView(userID, cardID), nil)
	txRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Transaction")).Return(nil)

	uc := newUpdateTransactionUC(t, mgr, txRepo, invRepo, instRepo, cards, cats, clock, ids)
	cid := cardID.String()
	req := dtos.UpdateTransactionRequest{
		Description:     "Updated credit",
		Amount:          "150.00",
		OccurredAt:      fixedNow,
		TransactionType: "credit_purchase",
		PaymentMethod:   "credit_card",
		CardID:          &cid,
		CategoryID:      catID.String(),
	}
	resp, err := uc.Execute(ctx, userID, txID, req)

	require.NoError(t, err)
	assert.Equal(t, "credit_purchase", resp.TransactionType)
	assert.Equal(t, "Updated credit", resp.Description)
}

func TestUpdateTransaction_NotFound_ReturnsError(t *testing.T) {
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

	txRepo.On("GetByID", mock.Anything, userID, txID).Return(nil, domain.ErrTransactionNotFound)

	uc := newUpdateTransactionUC(t, mgr, txRepo, invRepo, instRepo, cards, cats, clock, ids)
	req := dtos.UpdateTransactionRequest{
		Description:     "x",
		Amount:          "10.00",
		OccurredAt:      fixedNow,
		TransactionType: "expense",
		PaymentMethod:   "pix",
		CategoryID:      catID.String(),
	}
	_, err := uc.Execute(ctx, userID, txID, req)

	assert.ErrorIs(t, err, domain.ErrTransactionNotFound)
}
