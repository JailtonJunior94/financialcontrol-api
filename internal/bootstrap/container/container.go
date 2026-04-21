package container

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/services"
	appusecase "github.com/jailtonjunior94/financialcontrol-api/internal/application/usecase"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/events"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/interfaces"
	domainusecases "github.com/jailtonjunior94/financialcontrol-api/internal/domain/usecases"
	"github.com/jailtonjunior94/financialcontrol-api/internal/http/controllers"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/adapters"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/config"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/repositories"
	cardsapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application"
	cardshttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/http"
	cardsinfra "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure"
	catalogapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog/application"
	cataloghttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog/http"
	cataloginfra "github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog/infrastructure"
	identityapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application"
	identityhttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/http"
	identityinfra "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/infrastructure"
	platformsecurity "github.com/jailtonjunior94/financialcontrol-api/internal/platform/security"
)

type Container struct {
	SqlConnection         database.ISqlConnection
	HashAdapter           platformsecurity.HashAdapter
	JwtAdapter            platformsecurity.TokenAdapter
	UuidAdapter           adapters.IUuidAdapter
	UserRepository        identityapp.UserRepository
	TransactionRepository interfaces.ITransactionRepository
	BillRepository        interfaces.IBillRepository
	FlagRepository        catalogapp.FlagRepository
	CardRepository        cardsapp.CardRepository
	InvoiceRepository     interfaces.IInvoiceRepository
	CategoryRepository    catalogapp.CategoryRepository
	UserService           identityapp.UserService
	AuthService           identityapp.AuthService
	TransactionService    domainusecases.ITransactionService
	BillService           domainusecases.IBillService
	FlagService           catalogapp.FlagService
	CardService           cardsapp.CardService
	InvoiceService        domainusecases.IInvoiceService
	CategoryService       catalogapp.CategoryService
	UserController        *identityhttp.UserController
	AuthController        *identityhttp.AuthController
	TransactionController *controllers.TransactionController
	BillController        *controllers.BillController
	FlagController        *cataloghttp.FlagController
	CardController        *cardshttp.CardController
	InvoiceController     *controllers.InvoiceController
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
	c.BillRepository = repositories.NewBillRepository(c.SqlConnection)
	c.FlagRepository = cataloginfra.NewFlagRepository(c.SqlConnection)
	c.CardRepository = cardsinfra.NewCardRepository(c.SqlConnection)
	c.InvoiceRepository = repositories.NewInvoiceRepository(c.SqlConnection)
	c.CategoryRepository = cataloginfra.NewCategoryRepository(c.SqlConnection)
	c.TransactionRepository = repositories.NewTransactionRepository(c.SqlConnection)

	eventDispatcher := events.NewEventDispatcher()

	c.BillService = services.NewBillService(c.BillRepository)
	c.FlagService = catalogapp.NewFlagService(c.FlagRepository)
	c.CardService = cardsapp.NewCardService(c.CardRepository)
	c.CategoryService = catalogapp.NewCategoryService(c.CategoryRepository)
	c.UserService = identityapp.NewUserService(c.UserRepository, c.HashAdapter)
	c.TransactionService = services.NewTransactionService(c.TransactionRepository)
	c.AuthService = identityapp.NewAuthService(c.UserRepository, c.HashAdapter, c.JwtAdapter)
	c.InvoiceService = services.NewInvoiceService(c.CardRepository, c.InvoiceRepository, eventDispatcher)

	eventDispatcher.AddListener("invoice_changed", handlers.NewInvoiceChangedListener(
		c.InvoiceRepository,
		c.TransactionService,
		c.TransactionRepository,
	))

	c.UserController = identityhttp.NewUserController(c.UserService)
	c.BillController = controllers.NewBillController(c.BillService)
	c.FlagController = cataloghttp.NewFlagController(c.FlagService)
	c.CategoryController = cataloghttp.NewCategoryController(c.CategoryService)
	c.CardController = cardshttp.NewCardController(c.CardService, cardshttp.NewClaimsResolver(c.JwtAdapter))
	c.AuthController = identityhttp.NewAuthController(c.AuthService, identityhttp.NewClaimsResolver(c.JwtAdapter))
	c.InvoiceController = controllers.NewInvoiceController(c.InvoiceService, c.JwtAdapter)
	c.TransactionController = controllers.NewTransactionController(c.JwtAdapter, c.TransactionService)

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
