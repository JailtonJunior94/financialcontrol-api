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
)

type Container struct {
	SqlConnection         database.ISqlConnection
	HashAdapter           adapters.IHashAdapter
	JwtAdapter            adapters.IJwtAdapter
	UuidAdapter           adapters.IUuidAdapter
	UserRepository        interfaces.IUserRepository
	TransactionRepository interfaces.ITransactionRepository
	BillRepository        interfaces.IBillRepository
	FlagRepository        interfaces.IFlagRepository
	CardRepository        interfaces.ICardRepository
	InvoiceRepository     interfaces.IInvoiceRepository
	CategoryRepository    interfaces.ICategoryRepository
	UserService           domainusecases.IUserService
	AuthService           domainusecases.IAuthService
	TransactionService    domainusecases.ITransactionService
	BillService           domainusecases.IBillService
	FlagService           domainusecases.IFlagService
	CardService           domainusecases.ICardService
	InvoiceService        domainusecases.IInvoiceService
	CategoryService       domainusecases.ICategoryService
	UserController        *controllers.UserController
	AuthController        *controllers.AuthController
	TransactionController *controllers.TransactionController
	BillController        *controllers.BillController
	FlagController        *controllers.FlagController
	CardController        *controllers.CardController
	InvoiceController     *controllers.InvoiceController
	CategoryController    *controllers.CategoryController
	UpdateUseCase         *appusecase.UpdateTransactionUseCase
	UpdateTransactionBill *appusecase.UpdateTransactionBill
}

func Build(sqlConnection database.ISqlConnection) *Container {
	c := &Container{
		SqlConnection: sqlConnection,
		HashAdapter:   adapters.NewHashAdapter(),
		JwtAdapter:    adapters.NewJwtAdapter(),
		UuidAdapter:   adapters.NewUuidAdapter(),
	}

	c.UserRepository = repositories.NewUserRepository(c.SqlConnection)
	c.BillRepository = repositories.NewBillRepository(c.SqlConnection)
	c.FlagRepository = repositories.NewFlagRepository(c.SqlConnection)
	c.CardRepository = repositories.NewCardRepository(c.SqlConnection)
	c.InvoiceRepository = repositories.NewInvoiceRepository(c.SqlConnection)
	c.CategoryRepository = repositories.NewCategoryRepository(c.SqlConnection)
	c.TransactionRepository = repositories.NewTransactionRepository(c.SqlConnection)

	eventDispatcher := events.NewEventDispatcher()

	c.BillService = services.NewBillService(c.BillRepository)
	c.FlagService = services.NewFlagService(c.FlagRepository)
	c.CardService = services.NewCardService(c.CardRepository)
	c.CategoryService = services.NewCategoryService(c.CategoryRepository)
	c.UserService = services.NewUserService(c.UserRepository, c.HashAdapter)
	c.TransactionService = services.NewTransactionService(c.TransactionRepository)
	c.AuthService = services.NewAuthService(c.UserRepository, c.HashAdapter, c.JwtAdapter)
	c.InvoiceService = services.NewInvoiceService(c.CardRepository, c.InvoiceRepository, eventDispatcher)

	eventDispatcher.AddListener("invoice_changed", handlers.NewInvoiceChangedListener(
		c.InvoiceRepository,
		c.TransactionService,
		c.TransactionRepository,
	))

	c.UserController = controllers.NewUserController(c.UserService)
	c.BillController = controllers.NewBillController(c.BillService)
	c.FlagController = controllers.NewFlagController(c.FlagService)
	c.CategoryController = controllers.NewCategoryController(c.CategoryService)
	c.CardController = controllers.NewCardController(c.CardService, c.JwtAdapter)
	c.AuthController = controllers.NewAuthController(c.AuthService, c.JwtAdapter)
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
