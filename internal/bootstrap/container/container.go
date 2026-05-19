package container

import (
	"context"
	"fmt"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"
	"github.com/JailtonJunior94/devkit-go/pkg/observability"

	bootstrapconfig "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/config"
	bootstrapdbinstrumented "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/database/instrumented"
	bootstrapobs "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability"
	bootstrapmetrics "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/metrics"
	bootstrapredactor "github.com/jailtonjunior94/financialcontrol-api/internal/bootstrap/observability/redactor"
	cards "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards"
	categories "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories"
	finance "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance"
	financeproviders "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/providers"
	identity "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/config"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/database"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
	platformsecurity "github.com/jailtonjunior94/financialcontrol-api/pkg/security"
	pkguuid "github.com/jailtonjunior94/financialcontrol-api/pkg/uuid"
)

// Container holds every wired dependency for the application.
// It is the sole place allowed to import across module boundaries.
type Container struct {
	DBManager        manager.Manager
	Observability    observability.Observability
	Identity         bootstrapconfig.ServiceIdentity
	ShutdownTimeout  bootstrapconfig.ShutdownTimeout
	HashAdapter      platformsecurity.HashAdapter
	JwtParser        pkgjwt.Parser
	UuidAdapter      pkguuid.IUuidAdapter
	BusinessMetrics  *bootstrapmetrics.BusinessMetrics
	IdentityModule   *identity.Module
	CardsModule      *cards.Module
	CategoriesModule *categories.Module
	FinanceModule    *finance.Module
}

// Build wires all application dependencies. mgr, obs, id and timeout are obtained
// from the startup sequence in BuildRuntime and passed in explicitly.
func Build(ctx context.Context, mgr manager.Manager, obs observability.Observability, id bootstrapconfig.ServiceIdentity, timeout bootstrapconfig.ShutdownTimeout) (*Container, error) {
	c := &Container{
		DBManager:       mgr,
		Observability:   obs,
		Identity:        id,
		ShutdownTimeout: timeout,
		HashAdapter:     platformsecurity.NewHashAdapter(),
		UuidAdapter:     pkguuid.NewUuidAdapter(),
	}

	bm, err := bootstrapmetrics.NewBusinessMetrics(obs)
	if err != nil {
		return nil, fmt.Errorf("bootstrap container: build business metrics: %w", err)
	}
	c.BusinessMetrics = bm

	jwtCfg := pkgjwt.Config{
		Secret:    []byte(config.JwtSecret),
		AccessTTL: time.Hour * time.Duration(config.ExpirationAt),
	}

	jwtIssuer, err := pkgjwt.NewIssuer(jwtCfg)
	if err != nil {
		return nil, fmt.Errorf("build jwt issuer: %w", err)
	}

	jwtParser, err := pkgjwt.NewParser(jwtCfg)
	if err != nil {
		return nil, fmt.Errorf("build jwt parser: %w", err)
	}
	c.JwtParser = jwtParser

	dbtx := mgr.DBTX(ctx)

	c.IdentityModule = identity.NewModule(identity.Deps{
		DB:          dbtx,
		HashAdapter: c.HashAdapter,
		JwtIssuer:   jwtIssuer,
	})

	c.CardsModule = cards.NewModule(cards.Deps{DB: dbtx})
	c.CategoriesModule = categories.NewModule(categories.Deps{DB: dbtx})

	cardProvider := financeproviders.NewCardProviderAdapter(c.CardsModule.CardRepository())
	catProvider := financeproviders.NewCategoryProviderAdapter(c.CategoriesModule.CategoryRepository())

	c.FinanceModule = finance.NewModule(finance.Deps{
		Manager:          mgr,
		DB:               dbtx,
		CardProvider:     cardProvider,
		CategoryProvider: catProvider,
		Metrics:          c.BusinessMetrics,
	})

	return c, nil
}

// newObservability is the seam used by BuildRuntime to construct the observability
// provider. Tests swap it via SetNewObservabilityForTest to inject spies.
var newObservability = bootstrapobs.New

// SetNewObservabilityForTest swaps the observability factory and returns a restore
// function. Test-only; production code must not call this.
func SetNewObservabilityForTest(fn func(context.Context, bootstrapobs.Settings) (observability.Observability, error)) (restore func()) {
	prev := newObservability
	newObservability = fn
	return func() { newObservability = prev }
}

// BuildRuntime runs the full startup sequence:
// config.Load → identity → shutdown timeout → observability → database → module wiring.
// Any failure propagates an error wrapped with "bootstrap container:". When a step
// after observability construction fails, BuildRuntime rolls back by calling
// obs.Shutdown (and mgr.Shutdown when applicable) so no provider goroutines leak.
// Callers (cmd/main.go) are responsible for os.Exit on non-nil error.
func BuildRuntime(ctx context.Context) (*Container, error) {
	if err := config.Load(); err != nil {
		return nil, fmt.Errorf("bootstrap container: %w", err)
	}

	id, err := bootstrapconfig.NewServiceIdentity()
	if err != nil {
		return nil, fmt.Errorf("bootstrap container: %w", err)
	}

	timeout, err := bootstrapconfig.NewShutdownTimeout()
	if err != nil {
		return nil, fmt.Errorf("bootstrap container: %w", err)
	}

	settings, err := bootstrapobs.LoadSettings(id)
	if err != nil {
		return nil, fmt.Errorf("bootstrap container: %w", err)
	}

	obs, err := newObservability(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("bootstrap container: %w", err)
	}

	dsn := config.SqlConnectionString
	if dsn != "" {
		dsn, err = bootstrapdbinstrumented.ConfigureConnectionString(dsn, id.Environment())
		if err != nil {
			_ = obs.Shutdown(ctx)
			return nil, fmt.Errorf("bootstrap container: %w", err)
		}
	}

	mgr, err := database.OpenManager(ctx, dsn)
	if err != nil {
		_ = obs.Shutdown(ctx)
		return nil, fmt.Errorf("bootstrap container: %w", err)
	}
	mgr = bootstrapdbinstrumented.WrapManager(
		mgr,
		obs,
		bootstrapdbinstrumented.DatabaseNameFromConnectionString(dsn),
		time.Duration(settings.SlowQueryThresholdMS)*time.Millisecond,
		bootstrapredactor.DefaultDenylist,
	)

	c, err := Build(ctx, mgr, obs, id, timeout)
	if err != nil {
		_ = mgr.Shutdown(ctx)
		_ = obs.Shutdown(ctx)
		return nil, fmt.Errorf("bootstrap container: %w", err)
	}

	return c, nil
}
