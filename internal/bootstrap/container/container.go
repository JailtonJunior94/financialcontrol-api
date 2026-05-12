package container

import (
	"context"
	"fmt"
	"time"

	"github.com/JailtonJunior94/devkit-go/pkg/database/manager"

	cards "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards"
	cardsmssql "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/persistence/mssql"
	categories "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories"
	categoriesmssql "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/infrastructure/persistence/mssql"
	finance "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance"
	financeclock "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/clock"
	financeidempotency "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/idempotency"
	financeidgen "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/idgen"
	financemssql "github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/infrastructure/persistence/mssql"
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
	HashAdapter      platformsecurity.HashAdapter
	JwtParser        pkgjwt.Parser
	UuidAdapter      pkguuid.IUuidAdapter
	IdentityModule   *identity.Module
	CardsModule      *cards.Module
	CategoriesModule *categories.Module
	FinanceModule    *finance.Module
}

// Build wires all application dependencies. mgr is the database Manager obtained
// from database.OpenManager.
func Build(mgr manager.Manager) (*Container, error) {
	c := &Container{
		DBManager:   mgr,
		HashAdapter: platformsecurity.NewHashAdapter(),
		UuidAdapter: pkguuid.NewUuidAdapter(),
	}

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

	dbtx := mgr.DBTX(context.Background())

	c.IdentityModule = identity.NewModule(identity.Deps{
		DB:          dbtx,
		Hasher:      identity.NewHasherAdapter(c.HashAdapter),
		TokenIssuer: identity.NewTokenIssuerAdapter(jwtIssuer),
	})

	c.CardsModule = cards.NewModule(cards.Deps{DB: dbtx, JwtParser: jwtParser})
	c.CategoriesModule = categories.NewModule(categories.Deps{DB: dbtx, JwtParser: jwtParser})

	cardRepo := cardsmssql.NewCardRepository(dbtx)
	catRepo := categoriesmssql.NewCategoryRepository(dbtx)
	cardProvider := financeproviders.NewCardProviderAdapter(cardRepo)
	catProvider := financeproviders.NewCategoryProviderAdapter(catRepo)

	txRepo := financemssql.NewTransactionRepository(dbtx)
	invRepo := financemssql.NewInvoiceRepository(dbtx)
	instRepo := financemssql.NewInstallmentRepository(dbtx)
	idempRepo := financeidempotency.NewMSSQLRepository(dbtx)

	c.FinanceModule = finance.NewModule(finance.Deps{
		Manager:          mgr,
		JwtParser:        jwtParser,
		CardProvider:     cardProvider,
		CategoryProvider: catProvider,
		TxRepo:           txRepo,
		InvRepo:          invRepo,
		InstRepo:         instRepo,
		IdempotencyRepo:  idempRepo,
		Clock:            financeclock.NewSystemClock(),
		IDGen:            financeidgen.NewUUIDGenerator(),
	})

	return c, nil
}

func BuildRuntime() (*Container, error) {
	if err := config.Load(); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	mgr, err := database.OpenManager(context.Background(), config.SqlConnectionString)
	if err != nil {
		return nil, err
	}
	return Build(mgr)
}
