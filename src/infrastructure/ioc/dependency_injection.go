package ioc

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/services"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/usecase"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/events"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/adapters"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/database"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/repositories"
	"github.com/jailtonjunior94/financialcontrol-api/src/presentation/controllers"

	uc "github.com/jailtonjunior94/financialcontrol-api/src/application/usecase"
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
	UpdateUseCase         *usecase.UpdateTransactionUseCase
	UpdateTransactionBill *usecase.UpdateTransactionBill
)

func SetupDependencyInjection(sqlConnection database.ISqlConnection) {
	/* Database */
	SqlConnection = sqlConnection

	/* Adapters */
	JwtAdapter = adapters.NewJwtAdapter()
	HashAdapter = adapters.NewHashAdapter()
	UuidAdapter = adapters.NewUuidAdapter()

	/* Repositories */
	UserRepository = repositories.NewUserRepository(SqlConnection)
	BillRepository = repositories.NewBillRepository(SqlConnection)
	FlagRepository = repositories.NewFlagRepository(SqlConnection)
	CardRepository = repositories.NewCardRepository(SqlConnection)
	InvoiceRepository = repositories.NewInvoiceRepository(SqlConnection)
	CategoryRepository = repositories.NewCategoryRepository(SqlConnection)
	TransactionRepository = repositories.NewTransactionRepository(SqlConnection)

	/* Register Events */
	EventDispatcher := events.NewEventDispatcher()

	/* Services */
	BillService = services.NewBillService(BillRepository)
	FlagService = services.NewFlagService(FlagRepository)
	CardService = services.NewCardService(CardRepository)
	CategoryService = services.NewCategoryService(CategoryRepository)
	UserService = services.NewUserService(UserRepository, HashAdapter)
	TransactionService = services.NewTransactionService(TransactionRepository)
	AuthService = services.NewAuthService(UserRepository, HashAdapter, JwtAdapter)
	InvoiceService = services.NewInvoiceService(CardRepository, InvoiceRepository, EventDispatcher)

	EventDispatcher.AddListener("invoice_changed", handlers.NewInvoiceChangedListener(InvoiceRepository, TransactionService, TransactionRepository))

	/* Controllers */
	UserController = controllers.NewUserController(UserService)
	BillController = controllers.NewBillController(BillService)
	FlagController = controllers.NewFlagController(FlagService)
	CategoryController = controllers.NewCategoryController(CategoryService)
	CardController = controllers.NewCardController(CardService, JwtAdapter)
	AuthController = controllers.NewAuthController(AuthService, JwtAdapter)
	InvoiceController = controllers.NewInvoiceController(InvoiceService, JwtAdapter)
	TransactionController = controllers.NewTransactionController(JwtAdapter, TransactionService)

	/* Use Cases */
	UpdateUseCase = uc.NewUpdateTransactionUseCase(TransactionRepository, InvoiceRepository, TransactionService)
	UpdateTransactionBill = uc.NewUpdateTransactionBill(BillRepository, TransactionService, TransactionRepository)
}
