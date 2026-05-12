package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"

	database "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/database"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	domain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/projections"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/idempotency"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

const createTransactionEndpoint = "POST /finance/transactions"

// CreateTransaction is the use case interface for creating a new transaction (RF-01..04, RF-26).
type CreateTransaction interface {
	Execute(ctx context.Context, userID identityvo.UserID, key vos.IdempotencyKey, req dtos.CreateTransactionRequest) (dtos.TransactionResponse, error)
}

type createTransaction struct {
	mgr      manager.Manager
	txRepo   ports.TransactionRepository
	invRepo  ports.InvoiceRepository
	instRepo ports.InstallmentRepository
	idemp    idempotency.IdempotencyRepository
	cards    ports.CardProvider
	cats     ports.CategoryProvider
	splitter *services.InstallmentSplitter
	clock    ports.Clock
	ids      ports.IDGenerator
}

// NewCreateTransaction constructs the CreateTransaction use case with all required collaborators.
func NewCreateTransaction(
	mgr manager.Manager,
	txRepo ports.TransactionRepository,
	invRepo ports.InvoiceRepository,
	instRepo ports.InstallmentRepository,
	idemp idempotency.IdempotencyRepository,
	cards ports.CardProvider,
	cats ports.CategoryProvider,
	splitter *services.InstallmentSplitter,
	clock ports.Clock,
	ids ports.IDGenerator,
) CreateTransaction {
	return &createTransaction{
		mgr:      mgr,
		txRepo:   txRepo,
		invRepo:  invRepo,
		instRepo: instRepo,
		idemp:    idemp,
		cards:    cards,
		cats:     cats,
		splitter: splitter,
		clock:    clock,
		ids:      ids,
	}
}

func (uc *createTransaction) Execute(ctx context.Context, userID identityvo.UserID, key vos.IdempotencyKey, req dtos.CreateTransactionRequest) (dtos.TransactionResponse, error) {
	reqHash, err := hashRequest(req)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	hit, err := uc.idemp.Get(ctx, userID, createTransactionEndpoint, key.Value())
	if err != nil {
		return dtos.TransactionResponse{}, err
	}
	if hit != nil {
		if hit.RequestHash != reqHash {
			return dtos.TransactionResponse{}, domain.ErrIdempotencyMismatch
		}
		var cached dtos.TransactionResponse
		if jsonErr := json.Unmarshal([]byte(hit.ResponseBody), &cached); jsonErr != nil {
			return dtos.TransactionResponse{}, jsonErr
		}
		return cached, nil
	}

	amount, err := vos.NewAmount(req.Amount)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}
	txType, err := vos.ParseTransactionType(req.TransactionType)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}
	pm, err := vos.ParsePaymentMethod(req.PaymentMethod)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}
	catID, err := vos.ParseCategoryID(req.CategoryID)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	cardID, cardView, err := uc.resolveCard(ctx, userID, req.CardID, pm)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	if err := uc.validateCategory(ctx, userID, catID, req.SubcategoryID); err != nil {
		return dtos.TransactionResponse{}, err
	}

	subcatID, err := parseOptionalCategoryID(req.SubcategoryID)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	tx, err := entities.NewTransaction(
		uc.ids.NewTransactionID(),
		userID,
		req.Description,
		amount,
		req.OccurredAt,
		txType,
		pm,
		cardID,
		catID,
		subcatID,
		nil,
		uc.clock,
	)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	installments, err := uc.buildInstallments(ctx, userID, tx, txType, cardID, amount, req.InstallmentCount, cardView, req.OccurredAt)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}
	tx.SetInstallments(installments)

	resp := toTransactionResponse(tx)
	respBody, err := json.Marshal(resp)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	now := uc.clock.Now().UTC()
	err = database.Do(ctx, uc.mgr, func(ctx context.Context) error {
		if addErr := uc.txRepo.Add(ctx, tx); addErr != nil {
			return addErr
		}
		if len(installments) > 0 {
			if batchErr := uc.instRepo.AddBatch(ctx, installments); batchErr != nil {
				return batchErr
			}
		}
		return uc.idemp.Save(ctx, idempotency.IdempotencyRecord{
			UserID:       userID.String(),
			Endpoint:     createTransactionEndpoint,
			Key:          key.Value(),
			RequestHash:  reqHash,
			ResponseBody: string(respBody),
			StatusCode:   201,
			CreatedAt:    now,
			ExpiresAt:    now.Add(time.Hour),
		})
	})
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	return resp, nil
}

func (uc *createTransaction) resolveCard(ctx context.Context, userID identityvo.UserID, rawCardID *string, pm vos.PaymentMethod) (*vos.CardID, projections.CardView, error) {
	if !pm.RequiresCard() {
		return nil, projections.CardView{}, nil
	}
	if rawCardID == nil {
		return nil, projections.CardView{}, domain.ErrCardRequiredForPaymentMethod
	}
	cid, err := vos.ParseCardID(*rawCardID)
	if err != nil {
		return nil, projections.CardView{}, err
	}
	cv, err := uc.cards.GetByID(ctx, userID, cid)
	if err != nil {
		return nil, projections.CardView{}, err
	}
	if !cv.Active {
		return nil, projections.CardView{}, domain.ErrCardNotActive
	}
	return &cid, cv, nil
}

func (uc *createTransaction) validateCategory(ctx context.Context, userID identityvo.UserID, catID vos.CategoryID, rawSubcatID *string) error {
	catView, err := uc.cats.GetByID(ctx, userID, catID)
	if err != nil {
		return err
	}
	if !catView.Active {
		return domain.ErrCategoryNotActive
	}
	if rawSubcatID == nil {
		return nil
	}
	subcatID, err := vos.ParseCategoryID(*rawSubcatID)
	if err != nil {
		return err
	}
	subView, err := uc.cats.GetByID(ctx, userID, subcatID)
	if errors.Is(err, domain.ErrCategoryNotFound) {
		return domain.ErrSubcategoryNotFound
	}
	if err != nil {
		return err
	}
	if !subView.Active {
		return domain.ErrSubcategoryNotActive
	}
	if subView.ParentID == nil || *subView.ParentID != catID {
		return domain.ErrSubcategoryNotChildOfCategory
	}
	return nil
}

func (uc *createTransaction) buildInstallments(
	ctx context.Context,
	userID identityvo.UserID,
	tx *entities.Transaction,
	txType vos.TransactionType,
	cardID *vos.CardID,
	amount vos.Amount,
	installmentCount int,
	cardView projections.CardView,
	occurredAt time.Time,
) ([]*entities.Installment, error) {
	if !txType.IsCardBased() {
		return nil, nil
	}
	saoPaulo, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return nil, err
	}
	occurredAtLocal := occurredAt.In(saoPaulo)

	if txType != vos.TransactionTypeInstallmentPurchase {
		return uc.buildSingleInstallment(ctx, userID, tx, cardID, amount, cardView, occurredAtLocal)
	}
	return uc.buildMultipleInstallments(ctx, userID, tx, cardID, amount, installmentCount, cardView, occurredAtLocal)
}

func (uc *createTransaction) buildSingleInstallment(
	ctx context.Context,
	userID identityvo.UserID,
	tx *entities.Transaction,
	cardID *vos.CardID,
	amount vos.Amount,
	cardView projections.CardView,
	occurredAtLocal time.Time,
) ([]*entities.Installment, error) {
	invoice, err := uc.invRepo.AssignOrCreateOpen(ctx, userID, *cardID, occurredAtLocal, cardView, uc.clock)
	if err != nil {
		return nil, err
	}
	count, _ := vos.NewInstallmentCount(1)
	num, _ := vos.NewInstallmentNumber(1)
	now := uc.clock.Now().UTC()
	inst := entities.RehydrateInstallment(
		uc.ids.NewInstallmentID(),
		tx.ID(),
		invoice.ID(),
		num,
		count,
		amount.Money(),
		vos.InstallmentStatusScheduled,
		nil,
		now,
		now,
		nil,
	)
	return []*entities.Installment{inst}, nil
}

func (uc *createTransaction) buildMultipleInstallments(
	ctx context.Context,
	userID identityvo.UserID,
	tx *entities.Transaction,
	cardID *vos.CardID,
	amount vos.Amount,
	installmentCount int,
	cardView projections.CardView,
	occurredAtLocal time.Time,
) ([]*entities.Installment, error) {
	count, err := vos.NewInstallmentCount(installmentCount)
	if err != nil {
		return nil, err
	}
	amounts, instIDs := uc.splitter.Split(amount.Money(), count, uc.ids)
	installments := make([]*entities.Installment, len(amounts))
	now := uc.clock.Now().UTC()
	for i, amt := range amounts {
		instOccurredAt := addMonthsClamped(occurredAtLocal, i)
		invoice, invErr := uc.invRepo.AssignOrCreateOpen(ctx, userID, *cardID, instOccurredAt, cardView, uc.clock)
		if invErr != nil {
			return nil, invErr
		}
		num, _ := vos.NewInstallmentNumber(i + 1)
		installments[i] = entities.RehydrateInstallment(
			instIDs[i],
			tx.ID(),
			invoice.ID(),
			num,
			count,
			amt,
			vos.InstallmentStatusScheduled,
			nil,
			now,
			now,
			nil,
		)
	}
	return installments, nil
}

// hashRequest computes a stable SHA-256 of the JSON-encoded request for idempotency comparison.
func hashRequest(req any) (string, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("hash request: %w", err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(b)), nil
}

// parseOptionalCategoryID parses a nullable category ID pointer.
func parseOptionalCategoryID(raw *string) (*vos.CategoryID, error) {
	if raw == nil {
		return nil, nil
	}
	cid, err := vos.ParseCategoryID(*raw)
	if err != nil {
		return nil, err
	}
	return &cid, nil
}
