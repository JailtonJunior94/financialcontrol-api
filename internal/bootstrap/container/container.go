package container

import (
	billingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/application"
	billinghttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/http"
	billinginfra "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/infrastructure"
	cardsapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application"
	cardshttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/http"
	cardsinfra "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure"
	catalogapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog/application"
	cataloghttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog/http"
	cataloginfra "github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog/infrastructure"
	identityapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application"
	identityhttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/http"
	identityinfra "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/infrastructure"
	invoicingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/application"
	invoicinghttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/http"
	invoicinginfra "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/infrastructure"
	planningsync "github.com/jailtonjunior94/financialcontrol-api/internal/modules/planning/sync"
	transactionsapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/application"
	transactionshttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/http"
	transactionsinfra "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/infrastructure"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/config"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/database"
	platformevents "github.com/jailtonjunior94/financialcontrol-api/pkg/events"
	platformsecurity "github.com/jailtonjunior94/financialcontrol-api/pkg/security"
	pkguuid "github.com/jailtonjunior94/financialcontrol-api/pkg/uuid"
)

// Container holds every wired dependency for the application.
// It is the sole place allowed to import across module boundaries.
type Container struct {
	SqlConnection         database.ISqlConnection
	HashAdapter           platformsecurity.HashAdapter
	JwtAdapter            platformsecurity.TokenAdapter
	UuidAdapter           pkguuid.IUuidAdapter
	UserRepository        identityapp.UserRepository
	TransactionRepository transactionsapp.TransactionRepository
	BillRepository        billingapp.BillRepository
	FlagRepository        catalogapp.FlagRepository
	CardRepository        cardsapp.CardRepository
	InvoiceRepository     invoicingapp.InvoiceRepository
	CategoryRepository    catalogapp.CategoryRepository
	UserService           identityapp.UserService
	AuthService           identityapp.AuthService
	TransactionService    transactionsapp.TransactionAppService
	BillService           billingapp.BillService
	FlagService           catalogapp.FlagService
	CardService           cardsapp.CardService
	InvoiceService        invoicingapp.InvoiceService
	CategoryService       catalogapp.CategoryService
	UserController        *identityhttp.UserController
	AuthController        *identityhttp.AuthController
	TransactionController *transactionshttp.TransactionController
	BillController        *billinghttp.BillController
	FlagController        *cataloghttp.FlagController
	CardController        *cardshttp.CardController
	InvoiceController     *invoicinghttp.InvoiceController
	CategoryController    *cataloghttp.CategoryController
	UpdateUseCase         *planningsync.UpdateTransactionUseCase
	UpdateTransactionBill *planningsync.UpdateTransactionBill
}

// invoicingCardRepositoryAdapter adapts the cards repository to the invoicing
// module's CardRepository interface. The container is the only place allowed
// to import across module boundaries.
type invoicingCardRepositoryAdapter struct {
	repo *cardsinfra.CardRepository
}

func (a *invoicingCardRepositoryAdapter) GetCardById(id, userID string) (*invoicingapp.CardView, error) {
	card, err := a.repo.GetCardById(id, userID)
	if err != nil || card == nil {
		return nil, err
	}
	return &invoicingapp.CardView{ID: card.ID, ClosingDay: card.ClosingDay}, nil
}

func Build(sqlConnection database.ISqlConnection) *Container {
	c := &Container{
		SqlConnection: sqlConnection,
		HashAdapter:   platformsecurity.NewHashAdapter(),
		JwtAdapter:    platformsecurity.NewJWTAdapter(),
		UuidAdapter:   pkguuid.NewUuidAdapter(),
	}

	c.UserRepository = identityinfra.NewUserRepository(c.SqlConnection)
	c.BillRepository = billinginfra.NewBillRepository(c.SqlConnection)
	c.FlagRepository = cataloginfra.NewFlagRepository(c.SqlConnection)
	cardRepo := cardsinfra.NewCardRepository(c.SqlConnection)
	c.CardRepository = cardRepo
	c.InvoiceRepository = invoicinginfra.NewInvoiceRepository(c.SqlConnection)
	c.CategoryRepository = cataloginfra.NewCategoryRepository(c.SqlConnection)
	c.TransactionRepository = transactionsinfra.NewTransactionRepository(c.SqlConnection)

	var dispatcher platformevents.EventDispatcher = platformevents.NewInProcessDispatcher()

	c.BillService = billingapp.NewBillService(c.BillRepository)
	c.FlagService = catalogapp.NewFlagService(c.FlagRepository)
	c.CardService = cardsapp.NewCardService(c.CardRepository)
	c.CategoryService = catalogapp.NewCategoryService(c.CategoryRepository)
	c.UserService = identityapp.NewUserService(c.UserRepository, c.HashAdapter)
	c.TransactionService = transactionsapp.NewTransactionService(c.TransactionRepository)
	c.AuthService = identityapp.NewAuthService(c.UserRepository, c.HashAdapter, c.JwtAdapter)
	invoicePublisher := invoicingapp.NewInvoiceChangedEventPublisher(dispatcher)
	invoicingCardRepo := &invoicingCardRepositoryAdapter{repo: cardRepo}
	c.InvoiceService = invoicingapp.NewInvoiceService(invoicingCardRepo, c.InvoiceRepository, invoicePublisher)

	// Wire the modular invoice_changed handler via the adapter so the event
	// dispatcher calls the new context-aware handler without service locator.
	invoiceSyncAdapter := transactionsapp.NewInvoiceSyncAdapter(c.TransactionRepository, c.TransactionService)
	invoiceChangedHandler := invoicingapp.NewInvoiceChangedHandler(invoiceSyncAdapter)
	dispatcher.AddListener("invoice_changed", invoicingapp.NewInvoiceChangedListenerAdapter(invoiceChangedHandler))

	c.UserController = identityhttp.NewUserController(c.UserService)
	c.BillController = billinghttp.NewBillController(c.BillService)
	c.FlagController = cataloghttp.NewFlagController(c.FlagService)
	c.CategoryController = cataloghttp.NewCategoryController(c.CategoryService)
	c.CardController = cardshttp.NewCardController(c.CardService, cardshttp.NewClaimsResolver(c.JwtAdapter))
	c.AuthController = identityhttp.NewAuthController(c.AuthService, identityhttp.NewClaimsResolver(c.JwtAdapter))
	c.InvoiceController = invoicinghttp.NewInvoiceController(c.InvoiceService, invoicinghttp.NewClaimsResolver(c.JwtAdapter))
	c.TransactionController = transactionshttp.NewTransactionController(c.TransactionService, transactionshttp.NewClaimsResolver(c.JwtAdapter))

	c.UpdateUseCase = planningsync.NewUpdateTransactionUseCase(
		c.TransactionRepository,
		c.InvoiceRepository,
		c.TransactionService,
	)
	c.UpdateTransactionBill = planningsync.NewUpdateTransactionBill(
		c.BillRepository,
		c.TransactionService,
		c.TransactionRepository,
	)

	return c
}

func BuildRuntime() *Container {
	config.SetupEnvironments()
	return Build(database.NewConnection())
}
