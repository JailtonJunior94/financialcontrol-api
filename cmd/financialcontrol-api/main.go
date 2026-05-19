package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	bootstrapcli "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/cli"
)

// version is injected at build time via -ldflags "-X main.version=<git-sha>".
// Falls back to "dev" for local builds without explicit injection.
var version = "dev"

func main() {
	if os.Getenv("SERVICE_VERSION") == "" {
		if err := os.Setenv("SERVICE_VERSION", version); err != nil {
			log.Fatal(err)
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := bootstrapcli.RunServer(ctx); err != nil {
		log.Fatal(err)
	}
}
