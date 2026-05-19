package cards

import (
	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"
	"github.com/gofiber/fiber/v2"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/usecase"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/http/handlers"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/http/routes"
	mssqlrepo "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/persistence/mssql"
)

// Deps holds the external dependencies required to build the cards module.
type Deps struct {
	DB devkitdb.DBTX
}

// Module holds all wired use cases and handlers for the cards domain.
type Module struct {
	ListCards      usecase.ListCards
	GetCard        usecase.GetCard
	CreateCard     usecase.CreateCard
	UpdateCard     usecase.UpdateCard
	DeactivateCard usecase.DeactivateCard
	ListFlags      usecase.ListFlags
	CardHandler    *handlers.CardHandler
	FlagHandler    *handlers.FlagHandler
}

// NewModule builds the cards module from its external dependencies.
func NewModule(deps Deps) *Module {
	cardRepo := mssqlrepo.NewCardRepository(deps.DB)
	flagRepo := mssqlrepo.NewFlagRepository(deps.DB)

	listCards := usecase.NewListCards(cardRepo)
	getCard := usecase.NewGetCard(cardRepo)
	createCard := usecase.NewCreateCard(cardRepo, flagRepo)
	updateCard := usecase.NewUpdateCard(cardRepo, flagRepo)
	deactivateCard := usecase.NewDeactivateCard(cardRepo)
	listFlags := usecase.NewListFlags(flagRepo)

	cardHandler := handlers.NewCardHandler(listCards, getCard, createCard, updateCard, deactivateCard)
	flagHandler := handlers.NewFlagHandler(listFlags)

	return &Module{
		ListCards:      listCards,
		GetCard:        getCard,
		CreateCard:     createCard,
		UpdateCard:     updateCard,
		DeactivateCard: deactivateCard,
		ListFlags:      listFlags,
		CardHandler:    cardHandler,
		FlagHandler:    flagHandler,
	}
}

// RegisterHTTP registers all cards routes on the provided router.
func (m *Module) RegisterHTTP(router fiber.Router, protected fiber.Handler) {
	routes.RegisterCardRoutes(router, m.CardHandler, m.FlagHandler, protected)
}
