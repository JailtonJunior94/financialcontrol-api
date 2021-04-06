package ioc

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/services"
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
	UserService           usecases.IUserService
	AuthService           usecases.IAuthService
	TransactionService    usecases.ITransactionService
	BillService           usecases.IBillService
	UserController        *controllers.UserController
	AuthController        *controllers.AuthController
	TransactionController *controllers.TransactionController
	BillController        *controllers.BillController
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

	/* Services */
	UserService = services.NewUserService(UserRepository, HashAdapter)
	AuthService = services.NewAuthService(UserRepository, HashAdapter, JwtAdapter)
	TransactionService = services.NewTransactionService(TransactionRepository)
	BillService = services.NewBillService(BillRepository)

	/* Controllers */
	UserController = controllers.NewUserController(UserService)
	AuthController = controllers.NewAuthController(AuthService)
	TransactionController = controllers.NewTransactionController(JwtAdapter, TransactionService)
	BillController = controllers.NewBillController(BillService)
}
