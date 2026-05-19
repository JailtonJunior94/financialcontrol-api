package usecase_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/usecase"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	portmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/idempotency"
	idempmocks "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/idempotency/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

func newCreateTransactionUC(
	t *testing.T,
	mgr *mockManager,
	txRepo *portmocks.TransactionRepository,
	invRepo *portmocks.InvoiceRepository,
	instRepo *portmocks.InstallmentRepository,
	idempRepo *idempmocks.IdempotencyRepository,
	cards *portmocks.CardProvider,
	cats *portmocks.CategoryProvider,
	clock *portmocks.Clock,
	ids *portmocks.IDGenerator,
) usecase.CreateTransaction {
	t.Helper()
	splitter := services.NewInstallmentSplitter()
	return usecase.NewCreateTransaction(mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, splitter, clock, ids, ports.NoopRecorder{})
}

func newCreateTransactionUCWithMetrics(
	t *testing.T,
	mgr *mockManager,
	txRepo *portmocks.TransactionRepository,
	invRepo *portmocks.InvoiceRepository,
	instRepo *portmocks.InstallmentRepository,
	idempRepo *idempmocks.IdempotencyRepository,
	cards *portmocks.CardProvider,
	cats *portmocks.CategoryProvider,
	clock *portmocks.Clock,
	ids *portmocks.IDGenerator,
	metrics ports.FinancialMetricsRecorder,
) usecase.CreateTransaction {
	t.Helper()
	splitter := services.NewInstallmentSplitter()
	return usecase.NewCreateTransaction(mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, splitter, clock, ids, metrics)
}

func validCreateRequest(catID string) dtos.CreateTransactionRequest {
	return dtos.CreateTransactionRequest{
		Description:     "Coffee",
		Amount:          "50.00",
		OccurredAt:      fixedNow,
		TransactionType: "expense",
		PaymentMethod:   "pix",
		CategoryID:      catID,
	}
}

func TestCreateTransaction_GoldenPath_Expense(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	txID := vos.NewTransactionID()
	key, _ := vos.NewIdempotencyKey("test-key-1")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	idempRepo := idempmocks.NewIdempotencyRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	clock.EXPECT().Now().Return(fixedNow)
	ids.On("NewTransactionID").Return(txID)

	idempRepo.EXPECT().Get(mock.Anything, userID, "POST /finance/transactions", key.Value()).Return(nil, nil)
	cats.EXPECT().GetByID(mock.Anything, userID, catID).Return(newActiveCategoryView(userID, catID), nil)
	txRepo.EXPECT().Add(mock.Anything, mock.AnythingOfType("*entities.Transaction")).Return(nil)
	idempRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("idempotency.IdempotencyRecord")).Return(nil)

	uc := newCreateTransactionUC(t, mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, clock, ids)
	req := validCreateRequest(catID.String())
	resp, err := uc.Execute(ctx, userID, key, req)

	require.NoError(t, err)
	assert.Equal(t, txID.String(), resp.ID)
	assert.Equal(t, "Coffee", resp.Description)
	assert.Equal(t, "expense", resp.TransactionType)
}

func TestCreateTransaction_IdempotencyHit_ReturnsCached(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	key, _ := vos.NewIdempotencyKey("test-key-hit")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	idempRepo := idempmocks.NewIdempotencyRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	req := validCreateRequest(catID.String())
	cachedResp := dtos.TransactionResponse{ID: "cached-id", Description: "Coffee"}
	cachedBody, _ := json.Marshal(cachedResp)

	reqBytes, _ := json.Marshal(req)
	reqHash := hashForTest(reqBytes)

	hit := &idempotency.IdempotencyHit{
		RequestHash:  reqHash,
		ResponseBody: string(cachedBody),
		StatusCode:   201,
	}

	idempRepo.EXPECT().Get(mock.Anything, userID, "POST /finance/transactions", key.Value()).Return(hit, nil)

	uc := newCreateTransactionUC(t, mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, clock, ids)
	resp, err := uc.Execute(ctx, userID, key, req)

	require.NoError(t, err)
	assert.Equal(t, "cached-id", resp.ID)
	txRepo.AssertNotCalled(t, "Add")
}

func TestCreateTransaction_IdempotencyMismatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	key, _ := vos.NewIdempotencyKey("test-key-mismatch")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	idempRepo := idempmocks.NewIdempotencyRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	hit := &idempotency.IdempotencyHit{
		RequestHash:  "different-hash",
		ResponseBody: `{"id":"old"}`,
		StatusCode:   201,
	}
	idempRepo.EXPECT().Get(mock.Anything, userID, "POST /finance/transactions", key.Value()).Return(hit, nil)

	uc := newCreateTransactionUC(t, mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, clock, ids)
	req := validCreateRequest(catID.String())
	_, err := uc.Execute(ctx, userID, key, req)

	assert.ErrorIs(t, err, domain.ErrIdempotencyMismatch)
}

func TestCreateTransaction_CategoryNotActive(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	key, _ := vos.NewIdempotencyKey("test-key-cat")

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
	inactiveView := newActiveCategoryView(userID, catID)
	inactiveView.Active = false
	cats.EXPECT().GetByID(mock.Anything, userID, catID).Return(inactiveView, nil)

	uc := newCreateTransactionUC(t, mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, clock, ids)
	req := validCreateRequest(catID.String())
	_, err := uc.Execute(ctx, userID, key, req)

	assert.ErrorIs(t, err, domain.ErrCategoryNotActive)
}

func TestCreateTransaction_CardNotActive(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	cardID := vos.NewCardID()
	key, _ := vos.NewIdempotencyKey("test-key-card-inactive")

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
	inactiveCard := newActiveCardView(userID, cardID)
	inactiveCard.Active = false
	cards.EXPECT().GetByID(mock.Anything, userID, cardID).Return(inactiveCard, nil)

	uc := newCreateTransactionUC(t, mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, clock, ids)
	cid := cardID.String()
	req := dtos.CreateTransactionRequest{
		Description:     "Card purchase",
		Amount:          "100.00",
		OccurredAt:      fixedNow,
		TransactionType: "credit_purchase",
		PaymentMethod:   "credit_card",
		CardID:          &cid,
		CategoryID:      catID.String(),
	}
	_, err := uc.Execute(ctx, userID, key, req)

	assert.ErrorIs(t, err, domain.ErrCardNotActive)
}

func TestCreateTransaction_UoWRollback_WhenInstallmentBatchFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	cardID := vos.NewCardID()
	txID := vos.NewTransactionID()
	instID := vos.NewInstallmentID()
	invoiceID := vos.NewInvoiceID()
	key, _ := vos.NewIdempotencyKey("test-key-rollback")

	tx := &mockTx{}
	mgr := &mockManager{}
	mgr.On("BeginTx", mock.Anything, mock.Anything).Return(tx, nil)
	tx.On("Rollback", mock.Anything).Return(nil)

	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	idempRepo := idempmocks.NewIdempotencyRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	clock.EXPECT().Now().Return(fixedNow)
	ids.On("NewTransactionID").Return(txID)
	ids.On("NewInstallmentID").Return(instID)

	idempRepo.EXPECT().Get(mock.Anything, userID, "POST /finance/transactions", key.Value()).Return(nil, nil)
	cats.EXPECT().GetByID(mock.Anything, userID, catID).Return(newActiveCategoryView(userID, catID), nil)
	cards.EXPECT().GetByID(mock.Anything, userID, cardID).Return(newActiveCardView(userID, cardID), nil)

	openInv := newOpenInvoiceForCard(userID, cardID, invoiceID)
	invRepo.EXPECT().AssignOrCreateOpen(mock.Anything, userID, cardID, mock.AnythingOfType("time.Time"), mock.Anything, mock.Anything).Return(openInv, nil)

	batchErr := errTest("batch failed")
	txRepo.EXPECT().Add(mock.Anything, mock.Anything).Return(nil)
	instRepo.EXPECT().AddBatch(mock.Anything, mock.Anything).Return(batchErr)

	uc := newCreateTransactionUC(t, mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, clock, ids)
	cid := cardID.String()
	req := dtos.CreateTransactionRequest{
		Description:     "Credit purchase",
		Amount:          "100.00",
		OccurredAt:      fixedNow,
		TransactionType: "credit_purchase",
		PaymentMethod:   "credit_card",
		CardID:          &cid,
		CategoryID:      catID.String(),
	}
	_, err := uc.Execute(ctx, userID, key, req)

	assert.ErrorIs(t, err, batchErr)
	idempRepo.AssertNotCalled(t, "Save")
	tx.AssertCalled(t, "Rollback", mock.Anything)
}

func TestCreateTransaction_WithSubcategory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	subcatID := vos.NewCategoryID()
	txID := vos.NewTransactionID()
	key, _ := vos.NewIdempotencyKey("test-key-subcat")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	idempRepo := idempmocks.NewIdempotencyRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	clock.EXPECT().Now().Return(fixedNow)
	ids.On("NewTransactionID").Return(txID)

	idempRepo.EXPECT().Get(mock.Anything, userID, "POST /finance/transactions", key.Value()).Return(nil, nil)

	// parent category
	catView := newActiveCategoryView(userID, catID)
	cats.EXPECT().GetByID(mock.Anything, userID, catID).Return(catView, nil)

	// subcategory - child of catID
	subView := newActiveCategoryView(userID, subcatID)
	subView.ParentID = &catID
	cats.EXPECT().GetByID(mock.Anything, userID, subcatID).Return(subView, nil)

	txRepo.EXPECT().Add(mock.Anything, mock.AnythingOfType("*entities.Transaction")).Return(nil)
	idempRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("idempotency.IdempotencyRecord")).Return(nil)

	uc := newCreateTransactionUC(t, mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, clock, ids)
	scid := subcatID.String()
	req := dtos.CreateTransactionRequest{
		Description:     "Grocery",
		Amount:          "30.00",
		OccurredAt:      fixedNow,
		TransactionType: "expense",
		PaymentMethod:   "pix",
		CategoryID:      catID.String(),
		SubcategoryID:   &scid,
	}
	resp, err := uc.Execute(ctx, userID, key, req)

	require.NoError(t, err)
	assert.Equal(t, txID.String(), resp.ID)
	assert.Equal(t, scid, *resp.SubcategoryID)
}

func TestCreateTransaction_CreditPurchase_SingleInstallment(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	cardID := vos.NewCardID()
	txID := vos.NewTransactionID()
	instID := vos.NewInstallmentID()
	invoiceID := vos.NewInvoiceID()
	key, _ := vos.NewIdempotencyKey("test-key-credit")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	idempRepo := idempmocks.NewIdempotencyRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	clock.EXPECT().Now().Return(fixedNow)
	ids.On("NewTransactionID").Return(txID)
	ids.On("NewInstallmentID").Return(instID)

	idempRepo.EXPECT().Get(mock.Anything, userID, "POST /finance/transactions", key.Value()).Return(nil, nil)
	cats.EXPECT().GetByID(mock.Anything, userID, catID).Return(newActiveCategoryView(userID, catID), nil)
	cards.EXPECT().GetByID(mock.Anything, userID, cardID).Return(newActiveCardView(userID, cardID), nil)

	openInv := newOpenInvoiceForCard(userID, cardID, invoiceID)
	invRepo.EXPECT().AssignOrCreateOpen(mock.Anything, userID, cardID, mock.AnythingOfType("time.Time"), mock.Anything, mock.Anything).Return(openInv, nil)
	txRepo.EXPECT().Add(mock.Anything, mock.AnythingOfType("*entities.Transaction")).Return(nil)
	instRepo.EXPECT().AddBatch(mock.Anything, mock.Anything).Return(nil)
	idempRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("idempotency.IdempotencyRecord")).Return(nil)

	uc := newCreateTransactionUC(t, mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, clock, ids)
	cid := cardID.String()
	req := dtos.CreateTransactionRequest{
		Description:     "Credit buy",
		Amount:          "100.00",
		OccurredAt:      fixedNow,
		TransactionType: "credit_purchase",
		PaymentMethod:   "credit_card",
		CardID:          &cid,
		CategoryID:      catID.String(),
	}
	resp, err := uc.Execute(ctx, userID, key, req)

	require.NoError(t, err)
	assert.Equal(t, txID.String(), resp.ID)
	assert.Equal(t, "credit_purchase", resp.TransactionType)
	assert.Len(t, resp.Installments, 1)
}

func TestCreateTransaction_InstallmentPurchase_MultipleInstallments(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	cardID := vos.NewCardID()
	txID := vos.NewTransactionID()
	instID1 := vos.NewInstallmentID()
	instID2 := vos.NewInstallmentID()
	invoiceID := vos.NewInvoiceID()
	key, _ := vos.NewIdempotencyKey("test-key-installment")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	idempRepo := idempmocks.NewIdempotencyRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	clock.EXPECT().Now().Return(fixedNow)
	ids.On("NewTransactionID").Return(txID)
	ids.On("NewInstallmentID").Return(instID1).Once()
	ids.On("NewInstallmentID").Return(instID2).Once()

	idempRepo.EXPECT().Get(mock.Anything, userID, "POST /finance/transactions", key.Value()).Return(nil, nil)
	cats.EXPECT().GetByID(mock.Anything, userID, catID).Return(newActiveCategoryView(userID, catID), nil)
	cards.EXPECT().GetByID(mock.Anything, userID, cardID).Return(newActiveCardView(userID, cardID), nil)

	openInv := newOpenInvoiceForCard(userID, cardID, invoiceID)
	invRepo.EXPECT().AssignOrCreateOpen(mock.Anything, userID, cardID, mock.AnythingOfType("time.Time"), mock.Anything, mock.Anything).Return(openInv, nil)
	txRepo.EXPECT().Add(mock.Anything, mock.AnythingOfType("*entities.Transaction")).Return(nil)
	instRepo.EXPECT().AddBatch(mock.Anything, mock.Anything).Return(nil)
	idempRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("idempotency.IdempotencyRecord")).Return(nil)

	uc := newCreateTransactionUC(t, mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, clock, ids)
	cid := cardID.String()
	req := dtos.CreateTransactionRequest{
		Description:      "Buy in 2x",
		Amount:           "100.00",
		OccurredAt:       fixedNow,
		TransactionType:  "installment_purchase",
		PaymentMethod:    "credit_card",
		CardID:           &cid,
		CategoryID:       catID.String(),
		InstallmentCount: 2,
	}
	resp, err := uc.Execute(ctx, userID, key, req)

	require.NoError(t, err)
	assert.Equal(t, txID.String(), resp.ID)
	assert.Equal(t, "installment_purchase", resp.TransactionType)
	assert.Len(t, resp.Installments, 2)
}

func TestCreateTransaction_RecordsMetrics_AfterSuccessfulCommit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	txID := vos.NewTransactionID()
	key, _ := vos.NewIdempotencyKey("test-key-metrics")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	idempRepo := idempmocks.NewIdempotencyRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)
	spy := &spyRecorder{}

	clock.EXPECT().Now().Return(fixedNow)
	ids.On("NewTransactionID").Return(txID)
	idempRepo.EXPECT().Get(mock.Anything, userID, "POST /finance/transactions", key.Value()).Return(nil, nil)
	cats.EXPECT().GetByID(mock.Anything, userID, catID).Return(newActiveCategoryView(userID, catID), nil)
	txRepo.EXPECT().Add(mock.Anything, mock.AnythingOfType("*entities.Transaction")).Return(nil)
	idempRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("idempotency.IdempotencyRecord")).Return(nil)

	uc := newCreateTransactionUCWithMetrics(t, mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, clock, ids, spy)
	req := validCreateRequest(catID.String())
	_, err := uc.Execute(ctx, userID, key, req)

	require.NoError(t, err)
	require.Len(t, spy.calls, 1)
	assert.Equal(t, vos.TransactionTypeExpense, spy.calls[0].txType)
	assert.Equal(t, vos.PaymentMethodPix, spy.calls[0].method)
}

func TestCreateTransaction_DoesNotRecordMetrics_OnCommitFailure(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	txID := vos.NewTransactionID()
	key, _ := vos.NewIdempotencyKey("test-key-no-metrics")

	tx := &mockTx{}
	mgr := &mockManager{}
	mgr.On("BeginTx", mock.Anything, mock.Anything).Return(tx, nil)
	tx.On("Rollback", mock.Anything).Return(nil)

	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	idempRepo := idempmocks.NewIdempotencyRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)
	spy := &spyRecorder{}

	clock.EXPECT().Now().Return(fixedNow)
	ids.On("NewTransactionID").Return(txID)
	idempRepo.EXPECT().Get(mock.Anything, userID, "POST /finance/transactions", key.Value()).Return(nil, nil)
	cats.EXPECT().GetByID(mock.Anything, userID, catID).Return(newActiveCategoryView(userID, catID), nil)
	txRepo.EXPECT().Add(mock.Anything, mock.AnythingOfType("*entities.Transaction")).Return(errTest("repo failure"))

	uc := newCreateTransactionUCWithMetrics(t, mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, clock, ids, spy)
	req := validCreateRequest(catID.String())
	_, err := uc.Execute(ctx, userID, key, req)

	require.Error(t, err)
	assert.Empty(t, spy.calls, "recorder must not be called on commit failure")
}

func TestCreateTransaction_PanicInRecorder_DoesNotPropagate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := identityvo.NewUserID()
	catID := vos.NewCategoryID()
	txID := vos.NewTransactionID()
	key, _ := vos.NewIdempotencyKey("test-key-panic")

	mgr, _ := newMockMgr(t)
	txRepo := portmocks.NewTransactionRepository(t)
	invRepo := portmocks.NewInvoiceRepository(t)
	instRepo := portmocks.NewInstallmentRepository(t)
	idempRepo := idempmocks.NewIdempotencyRepository(t)
	cards := portmocks.NewCardProvider(t)
	cats := portmocks.NewCategoryProvider(t)
	clock := portmocks.NewClock(t)
	ids := portmocks.NewIDGenerator(t)

	clock.EXPECT().Now().Return(fixedNow)
	ids.On("NewTransactionID").Return(txID)
	idempRepo.EXPECT().Get(mock.Anything, userID, "POST /finance/transactions", key.Value()).Return(nil, nil)
	cats.EXPECT().GetByID(mock.Anything, userID, catID).Return(newActiveCategoryView(userID, catID), nil)
	txRepo.EXPECT().Add(mock.Anything, mock.AnythingOfType("*entities.Transaction")).Return(nil)
	idempRepo.EXPECT().Save(mock.Anything, mock.AnythingOfType("idempotency.IdempotencyRecord")).Return(nil)

	uc := newCreateTransactionUCWithMetrics(t, mgr, txRepo, invRepo, instRepo, idempRepo, cards, cats, clock, ids, &panicRecorder{})
	req := validCreateRequest(catID.String())
	resp, err := uc.Execute(ctx, userID, key, req)

	require.NoError(t, err, "panic in recorder must not propagate to caller")
	assert.Equal(t, txID.String(), resp.ID)
}

// --- helpers local to this file ---

type testError string

func (e testError) Error() string { return string(e) }

func errTest(s string) error { return testError(s) }

func hashForTest(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func newOpenInvoiceForCard(userID identityvo.UserID, cardID vos.CardID, _ vos.InvoiceID) *entities.Invoice {
	card := newActiveCardView(userID, cardID)
	cycleStart := fixedNow.AddDate(0, 0, -5)
	cycleEnd := fixedNow.AddDate(0, 1, 0)
	closingDate := fixedNow.AddDate(0, 1, 10)
	dueDate := fixedNow.AddDate(0, 1, 20)
	inv, _ := entities.NewInvoice(card, cycleStart, cycleEnd, closingDate, dueDate, fixedNow)
	return inv
}

// spyRecorder records calls to RecordAmountProcessed for assertion in tests.
type spyRecorder struct {
	calls []spyCall
}

type spyCall struct {
	txType vos.TransactionType
	method vos.PaymentMethod
}

func (s *spyRecorder) RecordAmountProcessed(_ context.Context, _ vos.Money, txType vos.TransactionType, method vos.PaymentMethod) {
	s.calls = append(s.calls, spyCall{txType: txType, method: method})
}

// panicRecorder panics on every call to test safeRecord recovery.
type panicRecorder struct{}

func (panicRecorder) RecordAmountProcessed(_ context.Context, _ vos.Money, _ vos.TransactionType, _ vos.PaymentMethod) {
	panic("simulated recorder panic")
}
