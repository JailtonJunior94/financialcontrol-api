package ioc

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/services"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/events"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/adapters"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/repositories"
	"github.com/jailtonjunior94/financialcontrol-api/src/presentation/controllers"
)

var (
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
	UserService           usecases.IUserService
	AuthService           usecases.IAuthService
	TransactionService    usecases.ITransactionService
	BillService           usecases.IBillService
	FlagService           usecases.IFlagService
	CardService           usecases.ICardService
	InvoiceService        usecases.IInvoiceService
	CategoryService       usecases.ICategoryService
	UserController        *controllers.UserController
	AuthController        *controllers.AuthController
	TransactionController *controllers.TransactionController
	BillController        *controllers.BillController
	FlagController        *controllers.FlagController
	CardController        *controllers.CardController
	InvoiceController     *controllers.InvoiceController
	CategoryController    *controllers.CategoryController
)

func SetupDependencyInjection(sqlConnection database.ISqlConnection) {
	/* Database */
	SqlConnection = sqlConnection

	/* Adapters */
	HashAdapter = adapters.NewHashAdapter()
	JwtAdapter = adapters.NewJwtAdapter()
	UuidAdapter = adapters.NewUuidAdapter()

	/* Repositories */
	UserRepository = repositories.NewUserRepository(SqlConnection)
	TransactionRepository = repositories.NewTransactionRepository(SqlConnection)
	BillRepository = repositories.NewBillRepository(SqlConnection)
	FlagRepository = repositories.NewFlagRepository(SqlConnection)
	CardRepository = repositories.NewCardRepository(SqlConnection)
	InvoiceRepository = repositories.NewInvoiceRepository(SqlConnection)
	CategoryRepository = repositories.NewCategoryRepository(SqlConnection)

	/* Register Events */
	EventDispatcher := events.NewEventDispatcher()
	EventDispatcher.AddListener("invoice_changed", handlers.NewInvoiceChangedListener(InvoiceRepository))

	/* Services */
	UserService = services.NewUserService(UserRepository, HashAdapter)
	AuthService = services.NewAuthService(UserRepository, HashAdapter, JwtAdapter)
	TransactionService = services.NewTransactionService(TransactionRepository)
	BillService = services.NewBillService(BillRepository)
	FlagService = services.NewFlagService(FlagRepository)
	CardService = services.NewCardService(CardRepository)
	InvoiceService = services.NewInvoiceService(CardRepository, InvoiceRepository, EventDispatcher)
	CategoryService = services.NewCategoryService(CategoryRepository)

	/* Controllers */
	UserController = controllers.NewUserController(UserService)
	AuthController = controllers.NewAuthController(AuthService, JwtAdapter)
	TransactionController = controllers.NewTransactionController(JwtAdapter, TransactionService)
	BillController = controllers.NewBillController(BillService)
	FlagController = controllers.NewFlagController(FlagService)
	CardController = controllers.NewCardController(CardService, JwtAdapter)
	InvoiceController = controllers.NewInvoiceController(InvoiceService, JwtAdapter)
	CategoryController = controllers.NewCategoryController(CategoryService)
}
