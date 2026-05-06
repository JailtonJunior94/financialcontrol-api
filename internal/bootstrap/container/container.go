package container

import (
	"context"
	"log"
	"time"

	billingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/application"
	billinghttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/http"
	billinginfra "github.com/jailtonjunior94/financialcontrol-api/internal/modules/billing/infrastructure"
	cards "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards"
	cardsvos "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	cardsmssql "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/persistence/mssql"
	categories "github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories"
	identity "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity"
	invoicingapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/application"
	invoicinghttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/http"
	invoicinginfra "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/infrastructure"
	planningsync "github.com/jailtonjunior94/financialcontrol-api/internal/modules/planning/sync"
	transactionsapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/application"
	transactionshttp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/http"
	transactionsinfra "github.com/jailtonjunior94/financialcontrol-api/internal/modules/transactions/infrastructure"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/config"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/database"
	platformevents "github.com/jailtonjunior94/financialcontrol-api/pkg/events"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
	platformsecurity "github.com/jailtonjunior94/financialcontrol-api/pkg/security"
	pkguuid "github.com/jailtonjunior94/financialcontrol-api/pkg/uuid"
)

// Container holds every wired dependency for the application.
// It is the sole place allowed to import across module boundaries.
type Container struct {
	SqlConnection         database.ISqlConnection
	HashAdapter           platformsecurity.HashAdapter
	JwtParser             pkgjwt.Parser
	UuidAdapter           pkguuid.IUuidAdapter
	IdentityModule        *identity.Module
	CardsModule           *cards.Module
	CategoriesModule      *categories.Module
	TransactionRepository transactionsapp.TransactionRepository
	BillRepository        billingapp.BillRepository
	InvoiceRepository     invoicingapp.InvoiceRepository
	TransactionService    transactionsapp.TransactionAppService
	BillService           billingapp.BillService
	InvoiceService        invoicingapp.InvoiceService
	TransactionController *transactionshttp.TransactionController
	BillController        *billinghttp.BillController
	InvoiceController     *invoicinghttp.InvoiceController
	UpdateUseCase         *planningsync.UpdateTransactionUseCase
	UpdateTransactionBill *planningsync.UpdateTransactionBill
}

// invoicingCardRepositoryAdapter adapts the new cards mssql repository to the
// invoicing module's CardRepository interface. The container is the only place
// allowed to import across module boundaries.
type invoicingCardRepositoryAdapter struct {
	repo *cardsmssql.CardRepository
}

func (a *invoicingCardRepositoryAdapter) GetCardById(id, userID string) (*invoicingapp.CardView, error) {
	uid, err := identityvo.ParseUserID(userID)
	if err != nil {
		return nil, err
	}
	cid, err := cardsvos.ParseCardID(id)
	if err != nil {
		return nil, err
	}
	card, err := a.repo.GetByID(context.Background(), uid, cid)
	if err != nil {
		return nil, err
	}
	return &invoicingapp.CardView{ID: card.ID().String(), ClosingDay: card.ClosingDay().Int()}, nil
}

func Build(sqlConnection database.ISqlConnection) *Container {
	c := &Container{
		SqlConnection: sqlConnection,
		HashAdapter:   platformsecurity.NewHashAdapter(),
		UuidAdapter:   pkguuid.NewUuidAdapter(),
	}

	jwtCfg := pkgjwt.Config{
		Secret:    []byte(config.JwtSecret),
		AccessTTL: time.Hour * time.Duration(config.ExpirationAt),
	}

	jwtIssuer, err := pkgjwt.NewIssuer(jwtCfg)
	if err != nil {
		log.Fatalf("build jwt issuer: %v", err)
	}

	jwtParser, err := pkgjwt.NewParser(jwtCfg)
	if err != nil {
		log.Fatalf("build jwt parser: %v", err)
	}
	c.JwtParser = jwtParser

	c.IdentityModule = identity.NewModule(identity.Deps{
		DB:          sqlConnection,
		Hasher:      identity.NewHasherAdapter(c.HashAdapter),
		TokenIssuer: identity.NewTokenIssuerAdapter(jwtIssuer),
	})

	c.BillRepository = billinginfra.NewBillRepository(c.SqlConnection)
	c.InvoiceRepository = invoicinginfra.NewInvoiceRepository(c.SqlConnection)
	c.TransactionRepository = transactionsinfra.NewTransactionRepository(c.SqlConnection)

	c.CardsModule = cards.NewModule(cards.Deps{DB: sqlConnection, JwtParser: jwtParser})
	c.CategoriesModule = categories.NewModule(categories.Deps{DB: sqlConnection, JwtParser: jwtParser})

	var dispatcher platformevents.EventDispatcher = platformevents.NewInProcessDispatcher()

	c.BillService = billingapp.NewBillService(c.BillRepository)
	c.TransactionService = transactionsapp.NewTransactionService(c.TransactionRepository)
	invoicePublisher := invoicingapp.NewInvoiceChangedEventPublisher(dispatcher)
	invoicingCardRepo := &invoicingCardRepositoryAdapter{repo: cardsmssql.NewCardRepository(sqlConnection.Connect())}
	c.InvoiceService = invoicingapp.NewInvoiceService(invoicingCardRepo, c.InvoiceRepository, invoicePublisher)

	// Wire the modular invoice_changed handler via the adapter so the event
	// dispatcher calls the new context-aware handler without service locator.
	invoiceSyncAdapter := transactionsapp.NewInvoiceSyncAdapter(c.TransactionRepository, c.TransactionService)
	invoiceChangedHandler := invoicingapp.NewInvoiceChangedHandler(invoiceSyncAdapter)
	dispatcher.AddListener("invoice_changed", invoicingapp.NewInvoiceChangedListenerAdapter(invoiceChangedHandler))

	c.BillController = billinghttp.NewBillController(c.BillService)
	c.InvoiceController = invoicinghttp.NewInvoiceController(c.InvoiceService)
	c.TransactionController = transactionshttp.NewTransactionController(c.TransactionService)

	c.UpdateUseCase = planningsync.NewUpdateTransactionUseCase(
		c.TransactionRepository,
		c.InvoiceRepository,
		c.TransactionService,
	)
	c.UpdateTransactionBill = planningsync.NewUpdateTransactionBill(
		c.BillRepository,
		c.TransactionService,
		c.TransactionRepository,
	)

	return c
}

func BuildRuntime() *Container {
	config.SetupEnvironments()
	return Build(database.NewConnection())
}
