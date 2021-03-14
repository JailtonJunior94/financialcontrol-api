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
	SqlConnection  database.ISqlConnection
	HashAdapter    adapters.IHashAdapter
	JwtAdapter     adapters.IJwtAdapter
	UuidAdapter    adapters.IUuidAdapter
	UserRepository interfaces.IUserRepository
	UserService    usecases.IUserService
	UserController *controllers.UserController
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

	/* Services */
	UserService = services.NewUserService(UserRepository, HashAdapter)

	/* Controllers */
	UserController = controllers.NewUserController(UserService)
}
