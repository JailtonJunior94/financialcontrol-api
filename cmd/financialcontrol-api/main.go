package main

import (
	"log"

	bootstrapcli "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/cli"
)

func main() {
	if err := bootstrapcli.Execute(); err != nil {
		log.Fatal(err)
	}
}
