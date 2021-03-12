package configuration

import "github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/database"

var (
	SqlConnection database.ISqlConnection
)

func SetupDependencyInjection(sqlConnection database.ISqlConnection) {
	/* Database */
	SqlConnection = sqlConnection
}
