package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/migration"

	// Register named smoke hooks via their init() functions.
	_ "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/migration/hooks"
)

func main() {
	smokeHook := flag.String("smoke", "", "name of the smoke hook to run between migrations (e.g. finance)")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	var opts []migration.Option
	if *smokeHook != "" {
		opts = append(opts, migration.WithSmokeHook(*smokeHook))
	}

	if err := migration.Run(ctx, opts...); err != nil {
		log.SetFlags(0)
		log.Println(err.Error())
		os.Exit(1)
	}
}
