package app

import (
	"fmt"
	"log"
	"os"

	"github.com/jailtonjunior94/financialcontrol-api/src/app/configuration"
)

func Run() {
	app := configuration.App()

	fmt.Printf("🚀 API is running on http://0.0.0.0:%v", os.Getenv("PORT"))
	log.Fatal(app.Listen(fmt.Sprintf(":%v", os.Getenv("PORT"))))
}
