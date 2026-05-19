package usecase

import (
	"context"
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
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

type updateTransaction struct {
	mgr      manager.Manager
	txRepo   ports.TransactionRepository
	invRepo  ports.InvoiceRepository
	instRepo ports.InstallmentRepository
	cards    ports.CardProvider
	cats     ports.CategoryProvider
	splitter *services.InstallmentSplitter
	clock    ports.Clock
	ids      ports.IDGenerator
}

// NewUpdateTransaction constructs the UpdateTransaction use case.
func NewUpdateTransaction(
	mgr manager.Manager,
	txRepo ports.TransactionRepository,
	invRepo ports.InvoiceRepository,
	instRepo ports.InstallmentRepository,
	cards ports.CardProvider,
	cats ports.CategoryProvider,
	splitter *services.InstallmentSplitter,
	clock ports.Clock,
	ids ports.IDGenerator,
) UpdateTransaction {
	return &updateTransaction{
		mgr:      mgr,
		txRepo:   txRepo,
		invRepo:  invRepo,
		instRepo: instRepo,
		cards:    cards,
		cats:     cats,
		splitter: splitter,
		clock:    clock,
		ids:      ids,
	}
}

func (uc *updateTransaction) Execute(ctx context.Context, userID identityvo.UserID, txID vos.TransactionID, req dtos.UpdateTransactionRequest) (dtos.TransactionResponse, error) {
	tx, err := uc.txRepo.GetByID(ctx, userID, txID)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	existingInsts, err := uc.instRepo.ListByTransaction(ctx, txID)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}
	ptrs := make([]*entities.Installment, len(existingInsts))
	for i := range existingInsts {
		ptrs[i] = &existingInsts[i]
	}
	tx.SetInstallments(ptrs)

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

	cardID, cardView, err := uc.resolveCardUpdate(ctx, userID, req.CardID, pm)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	if err := uc.validateCategoryUpdate(ctx, userID, catID, req.SubcategoryID); err != nil {
		return dtos.TransactionResponse{}, err
	}

	subcatID, err := parseOptionalCategoryID(req.SubcategoryID)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	now := uc.clock.Now()

	hasClosedOrPaid, err := uc.instRepo.HasClosedOrPaidForTransaction(ctx, txID)
	if err != nil {
		return dtos.TransactionResponse{}, err
	}
	// Short-circuit before buildUpdateAdapters: AssignOrCreateOpen runs outside
	// the database.Do wrapper and may persist invoices that would be discarded
	// once tx.Replace rejects the request with ErrInstallmentInClosedOrPaidInvoice.
	if hasClosedOrPaid {
		return dtos.TransactionResponse{}, domain.ErrInstallmentInClosedOrPaidInvoice
	}

	var splitterAdapt entities.Splitter
	var assignerAdapt entities.Assigner
	if txType == vos.TransactionTypeInstallmentPurchase {
		splitterAdapt, assignerAdapt, err = uc.buildUpdateAdapters(ctx, userID, cardID, amount, req.InstallmentCount, cardView, req.OccurredAt, now)
		if err != nil {
			return dtos.TransactionResponse{}, err
		}
	}

	input := entities.ReplaceInput{
		Description:            req.Description,
		Amount:                 amount,
		OccurredAt:             req.OccurredAt,
		TransactionType:        txType,
		PaymentMethod:          pm,
		CardID:                 cardID,
		CategoryID:             catID,
		SubcategoryID:          subcatID,
		InstallmentCount:       req.InstallmentCount,
		Now:                    now,
		HasClosedOrPaidInvoice: hasClosedOrPaid,
	}
	if err := tx.Replace(input, splitterAdapt, assignerAdapt); err != nil {
		return dtos.TransactionResponse{}, err
	}

	err = database.Do(ctx, uc.mgr, func(ctx context.Context) error {
		if updErr := uc.txRepo.Update(ctx, tx); updErr != nil {
			return updErr
		}
		var oldInsts, newInsts []*entities.Installment
		for _, inst := range tx.Installments() {
			if inst.IsDeleted() {
				oldInsts = append(oldInsts, inst)
			} else {
				newInsts = append(newInsts, inst)
			}
		}
		if len(oldInsts) > 0 {
			if updErr := uc.instRepo.UpdateBatch(ctx, oldInsts); updErr != nil {
				return updErr
			}
		}
		if len(newInsts) > 0 {
			if addErr := uc.instRepo.AddBatch(ctx, newInsts); addErr != nil {
				return addErr
			}
		}
		return nil
	})
	if err != nil {
		return dtos.TransactionResponse{}, err
	}

	return toTransactionResponse(tx), nil
}

func (uc *updateTransaction) resolveCardUpdate(ctx context.Context, userID identityvo.UserID, rawCardID *string, pm vos.PaymentMethod) (*vos.CardID, projections.CardView, error) {
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

func (uc *updateTransaction) validateCategoryUpdate(ctx context.Context, userID identityvo.UserID, catID vos.CategoryID, rawSubcatID *string) error {
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
	if err != nil {
		return domain.ErrSubcategoryNotFound
	}
	if !subView.Active {
		return domain.ErrSubcategoryNotActive
	}
	if subView.ParentID == nil || *subView.ParentID != catID {
		return domain.ErrSubcategoryNotChildOfCategory
	}
	return nil
}

func (uc *updateTransaction) buildUpdateAdapters(
	ctx context.Context,
	userID identityvo.UserID,
	cardID *vos.CardID,
	amount vos.Amount,
	installmentCount int,
	cardView projections.CardView,
	occurredAt time.Time,
	now time.Time,
) (entities.Splitter, entities.Assigner, error) {
	count, err := vos.NewInstallmentCount(installmentCount)
	if err != nil {
		return nil, nil, err
	}
	amounts, instIDs := uc.splitter.Split(amount.Money(), count, uc.ids)

	saoPaulo, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return nil, nil, err
	}
	occurredAtLocal := occurredAt.In(saoPaulo)

	invoiceIDs := make(map[int]vos.InvoiceID, len(amounts))
	for i := range amounts {
		instOccurredAt := addMonthsClamped(occurredAtLocal, i)
		invoice, invErr := uc.invRepo.AssignOrCreateOpen(ctx, userID, *cardID, instOccurredAt, cardView, uc.clock)
		if invErr != nil {
			return nil, nil, invErr
		}
		invoiceIDs[i+1] = invoice.ID()
	}

	splAdapt := &precomputedSplitter{amounts: amounts, ids: instIDs}
	assAdapt := &precomputedAssigner{ids: invoiceIDs}
	return splAdapt, assAdapt, nil
}

// precomputedSplitter satisfies entities.Splitter using pre-computed values.
type precomputedSplitter struct {
	amounts []vos.Money
	ids     []vos.InstallmentID
}

func (s *precomputedSplitter) Split(_ vos.Money, _ vos.InstallmentCount) ([]vos.Money, []vos.InstallmentID) {
	return s.amounts, s.ids
}

// precomputedAssigner satisfies entities.Assigner using a pre-computed map.
type precomputedAssigner struct {
	ids map[int]vos.InvoiceID
}

func (a *precomputedAssigner) InvoiceForNumber(n int) (vos.InvoiceID, error) {
	id, ok := a.ids[n]
	if !ok {
		return vos.InvoiceID(""), domain.ErrInvoiceNotFound
	}
	return id, nil
}
