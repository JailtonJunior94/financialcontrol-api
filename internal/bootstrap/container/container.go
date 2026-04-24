package container

import (
	appusecase "github.com/jailtonjunior94/financialcontrol-api/internal/application/usecase"
	platformevents "github.com/jailtonjunior94/financialcontrol-api/internal/platform/events"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/adapters"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/config"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/database"
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
	transactionsapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/application"
	transactionshttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/http"
	transactionsinfra "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/infrastructure"
	platformsecurity "github.com/jailtonjunior94/financialcontrol-api/internal/platform/security"
)

// Container holds every wired dependency for the application.
// It is the sole place allowed to import across module boundaries.
type Container struct {
	SqlConnection         database.ISqlConnection
	HashAdapter           platformsecurity.HashAdapter
	JwtAdapter            platformsecurity.TokenAdapter
	UuidAdapter           adapters.IUuidAdapter
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
	UpdateUseCase         *appusecase.UpdateTransactionUseCase
	UpdateTransactionBill *appusecase.UpdateTransactionBill
}

func Build(sqlConnection database.ISqlConnection) *Container {
	c := &Container{
		SqlConnection: sqlConnection,
		HashAdapter:   platformsecurity.NewHashAdapter(),
		JwtAdapter:    platformsecurity.NewJWTAdapter(),
		UuidAdapter:   adapters.NewUuidAdapter(),
	}

	c.UserRepository = identityinfra.NewUserRepository(c.SqlConnection)
	c.BillRepository = billinginfra.NewBillRepository(c.SqlConnection)
	c.FlagRepository = cataloginfra.NewFlagRepository(c.SqlConnection)
	c.CardRepository = cardsinfra.NewCardRepository(c.SqlConnection)
	c.InvoiceRepository = invoicinginfra.NewInvoiceRepository(c.SqlConnection)
	c.CategoryRepository = cataloginfra.NewCategoryRepository(c.SqlConnection)
	c.TransactionRepository = transactionsinfra.NewTransactionRepository(c.SqlConnection)

	dispatcher := platformevents.NewDispatcher()

	c.BillService = billingapp.NewBillService(c.BillRepository)
	c.FlagService = catalogapp.NewFlagService(c.FlagRepository)
	c.CardService = cardsapp.NewCardService(c.CardRepository)
	c.CategoryService = catalogapp.NewCategoryService(c.CategoryRepository)
	c.UserService = identityapp.NewUserService(c.UserRepository, c.HashAdapter)
	c.TransactionService = transactionsapp.NewTransactionService(c.TransactionRepository)
	c.AuthService = identityapp.NewAuthService(c.UserRepository, c.HashAdapter, c.JwtAdapter)
	invoicePublisher := invoicingapp.NewInvoiceChangedEventPublisher(dispatcher)
	c.InvoiceService = invoicingapp.NewInvoiceService(c.CardRepository, c.InvoiceRepository, invoicePublisher)

	// Wire the modular invoice_changed handler via the adapter so the event
	// dispatcher calls the new context-aware handler without service locator.
	invoiceSyncAdapter := transactionsapp.NewInvoiceSyncAdapter(c.TransactionRepository, c.TransactionService)
	invoiceChangedHandler := invoicingapp.NewInvoiceChangedHandler(c.InvoiceRepository, invoiceSyncAdapter)
	dispatcher.AddListener("invoice_changed", invoicingapp.NewInvoiceChangedListenerAdapter(invoiceChangedHandler))

	c.UserController = identityhttp.NewUserController(c.UserService)
	c.BillController = billinghttp.NewBillController(c.BillService)
	c.FlagController = cataloghttp.NewFlagController(c.FlagService)
	c.CategoryController = cataloghttp.NewCategoryController(c.CategoryService)
	c.CardController = cardshttp.NewCardController(c.CardService, cardshttp.NewClaimsResolver(c.JwtAdapter))
	c.AuthController = identityhttp.NewAuthController(c.AuthService, identityhttp.NewClaimsResolver(c.JwtAdapter))
	c.InvoiceController = invoicinghttp.NewInvoiceController(c.InvoiceService, invoicinghttp.NewClaimsResolver(c.JwtAdapter))
	c.TransactionController = transactionshttp.NewTransactionController(c.TransactionService, transactionshttp.NewClaimsResolver(c.JwtAdapter))

	c.UpdateUseCase = appusecase.NewUpdateTransactionUseCase(
		c.TransactionRepository,
		c.InvoiceRepository,
		c.TransactionService,
	)
	c.UpdateTransactionBill = appusecase.NewUpdateTransactionBill(
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
