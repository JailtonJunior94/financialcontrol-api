package finance

import (
	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"
	"github.com/gofiber/fiber/v2"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/usecase"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/services"
	financeclock "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/clock"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/http/routes"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/idempotency"
	financeidgen "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/idgen"
	mssqlrepo "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/persistence/mssql"
)

// Deps holds the external dependencies required to build the finance module.
// Repositories, clock and id generator are infrastructure owned by the module
// and constructed inside NewModule; only raw handles (Manager/DB) and
// cross-module providers cross the boundary.
type Deps struct {
	Manager          manager.Manager
	DB               devkitdb.DBTX
	CardProvider     ports.CardProvider
	CategoryProvider ports.CategoryProvider
	// Metrics is the business metrics recorder. Defaults to NoopRecorder when nil.
	Metrics ports.FinancialMetricsRecorder
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
	if deps.Metrics == nil {
		deps.Metrics = ports.NoopRecorder{}
	}

	txRepo := mssqlrepo.NewTransactionRepository(deps.DB)
	invRepo := mssqlrepo.NewInvoiceRepository(deps.DB)
	instRepo := mssqlrepo.NewInstallmentRepository(deps.DB)
	idempRepo := idempotency.NewMSSQLRepository(deps.DB)
	clock := financeclock.NewSystemClock()
	idgen := financeidgen.NewUUIDGenerator()

	splitter := services.NewInstallmentSplitter()
	closer := services.NewInvoiceCloser()
	factory := services.NewRefundFactory()

	create := usecase.NewCreateTransaction(deps.Manager, txRepo, invRepo, instRepo, idempRepo, deps.CardProvider, deps.CategoryProvider, splitter, clock, idgen, deps.Metrics)
	list := usecase.NewListTransactions(deps.Manager, txRepo, invRepo, closer, clock)
	get := usecase.NewGetTransaction(txRepo, instRepo)
	update := usecase.NewUpdateTransaction(deps.Manager, txRepo, invRepo, instRepo, deps.CardProvider, deps.CategoryProvider, splitter, clock, idgen)
	del := usecase.NewDeleteTransaction(deps.Manager, txRepo, instRepo, clock)
	refund := usecase.NewRefundTransaction(deps.Manager, txRepo, factory, clock, idgen, deps.Metrics)

	listInv := usecase.NewListInvoices(deps.Manager, invRepo, closer, clock)
	getInv := usecase.NewGetInvoice(deps.Manager, invRepo, instRepo, closer, clock)
	payInv := usecase.NewPayInvoice(deps.Manager, invRepo, clock, deps.Metrics)

	anticipate := usecase.NewAnticipateInstallment(deps.Manager, txRepo, invRepo, instRepo, clock)
	summary := usecase.NewMonthlySummary(txRepo, invRepo, instRepo)

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
func (m *Module) RegisterHTTP(router fiber.Router, protected fiber.Handler) {
	routes.RegisterFinanceRoutes(router, m.TransactionHandler, m.InvoiceHandler, m.InstallmentHandler, m.SummaryHandler, protected)
}
