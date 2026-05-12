package finance

import (
	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"
	"github.com/gofiber/fiber/v2"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/usecase"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http/routes"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/idempotency"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
)

// Deps holds the external dependencies required to build the finance module.
type Deps struct {
	Manager          manager.Manager
	JwtParser        pkgjwt.Parser
	CardProvider     ports.CardProvider
	CategoryProvider ports.CategoryProvider
	TxRepo           ports.TransactionRepository
	InvRepo          ports.InvoiceRepository
	InstRepo         ports.InstallmentRepository
	IdempotencyRepo  idempotency.IdempotencyRepository
	Clock            ports.Clock
	IDGen            ports.IDGenerator
}

// Module holds all wired use cases and handlers for the finance domain.
type Module struct {
	CreateTransaction     usecase.CreateTransaction
	ListTransactions      usecase.ListTransactions
	GetTransaction        usecase.GetTransaction
	UpdateTransaction     usecase.UpdateTransaction
	DeleteTransaction     usecase.DeleteTransaction
	RefundTransaction     usecase.RefundTransaction
	ListInvoices          usecase.ListInvoices
	GetInvoice            usecase.GetInvoice
	PayInvoice            usecase.PayInvoice
	AnticipateInstallment usecase.AnticipateInstallment
	MonthlySummary        usecase.MonthlySummary

	TransactionHandler *handlers.TransactionHandler
	InvoiceHandler     *handlers.InvoiceHandler
	InstallmentHandler *handlers.InstallmentHandler
	SummaryHandler     *handlers.SummaryHandler
}

// NewModule builds the finance module from its external dependencies.
// Domain services are stateless and constructed here; use cases and handlers receive them via DI.
func NewModule(deps Deps) *Module {
	splitter := services.NewInstallmentSplitter()
	closer := services.NewInvoiceCloser()
	factory := services.NewRefundFactory()

	create := usecase.NewCreateTransaction(deps.Manager, deps.TxRepo, deps.InvRepo, deps.InstRepo, deps.IdempotencyRepo, deps.CardProvider, deps.CategoryProvider, splitter, deps.Clock, deps.IDGen)
	list := usecase.NewListTransactions(deps.Manager, deps.TxRepo, deps.InvRepo, closer, deps.Clock)
	get := usecase.NewGetTransaction(deps.TxRepo, deps.InstRepo)
	update := usecase.NewUpdateTransaction(deps.Manager, deps.TxRepo, deps.InvRepo, deps.InstRepo, deps.CardProvider, deps.CategoryProvider, splitter, deps.Clock, deps.IDGen)
	del := usecase.NewDeleteTransaction(deps.Manager, deps.TxRepo, deps.InstRepo, deps.Clock)
	refund := usecase.NewRefundTransaction(deps.Manager, deps.TxRepo, factory, deps.Clock, deps.IDGen)

	listInv := usecase.NewListInvoices(deps.Manager, deps.InvRepo, closer, deps.Clock)
	getInv := usecase.NewGetInvoice(deps.Manager, deps.InvRepo, deps.InstRepo, closer, deps.Clock)
	payInv := usecase.NewPayInvoice(deps.Manager, deps.InvRepo, deps.Clock)

	anticipate := usecase.NewAnticipateInstallment(deps.Manager, deps.TxRepo, deps.InvRepo, deps.InstRepo, deps.Clock)
	summary := usecase.NewMonthlySummary(deps.TxRepo, deps.InvRepo, deps.InstRepo)

	txHandler := handlers.NewTransactionHandler(create, list, get, update, del, refund)
	invHandler := handlers.NewInvoiceHandler(listInv, getInv, payInv)
	instHandler := handlers.NewInstallmentHandler(anticipate)
	summaryHandler := handlers.NewSummaryHandler(summary)

	return &Module{
		CreateTransaction:     create,
		ListTransactions:      list,
		GetTransaction:        get,
		UpdateTransaction:     update,
		DeleteTransaction:     del,
		RefundTransaction:     refund,
		ListInvoices:          listInv,
		GetInvoice:            getInv,
		PayInvoice:            payInv,
		AnticipateInstallment: anticipate,
		MonthlySummary:        summary,
		TransactionHandler:    txHandler,
		InvoiceHandler:        invHandler,
		InstallmentHandler:    instHandler,
		SummaryHandler:        summaryHandler,
	}
}

// RegisterHTTP registers all finance routes on the provided router.
func (m *Module) RegisterHTTP(router fiber.Router, parser pkgjwt.Parser) {
	routes.RegisterFinanceRoutes(router, m.TransactionHandler, m.InvoiceHandler, m.InstallmentHandler, m.SummaryHandler, parser)
}
