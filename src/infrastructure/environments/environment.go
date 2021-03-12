package environments

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

var (
	Environment         = ""
	SqlConnectionString = ""
	JwtSecret           = ""
	ExpirationAt        = 0
)

func SetupEnvironments() {
	var err error

	viper.SetConfigName(fmt.Sprintf("config.%s", os.Getenv("ENVIRONMENT")))
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err = viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	Environment = viper.GetString("environment")
	SqlConnectionString = viper.GetString("mssql.connectionString")
	JwtSecret = viper.GetString("security.jwtSecret")
	ExpirationAt = viper.GetInt("security.expirationAt")
}
