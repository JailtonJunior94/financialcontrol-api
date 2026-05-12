//go:build integration

package mssql_test

import (
	"context"
	"testing"
	"time"

	devkitdb "github.com/JailtonJunior94/devkit-go/pkg/database"
	devkitmgr "github.com/JailtonJunior94/devkit-go/pkg/database/manager"
	"github.com/stretchr/testify/suite"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
	repomssql "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/infrastructure/persistence/mssql"
	dbmssql "github.com/jailtonjunior94/financialcontrol-api/pkg/database/mssql"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

const createFlagTable = `
IF NOT EXISTS (
    SELECT * FROM INFORMATION_SCHEMA.TABLES
    WHERE TABLE_SCHEMA = 'dbo' AND TABLE_NAME = 'Flag'
)
CREATE TABLE dbo.[Flag] (
    [Id]     CHAR(36)      NOT NULL PRIMARY KEY,
    [Name]   NVARCHAR(100) NOT NULL,
    [Active] BIT           NOT NULL DEFAULT 1
)
`

const createCardTable = `
IF NOT EXISTS (
    SELECT * FROM INFORMATION_SCHEMA.TABLES
    WHERE TABLE_SCHEMA = 'dbo' AND TABLE_NAME = 'Card'
)
CREATE TABLE dbo.[Card] (
    [Id]             CHAR(36)       NOT NULL PRIMARY KEY,
    [UserId]         CHAR(36)       NOT NULL,
    [FlagId]         CHAR(36)       NOT NULL,
    [Name]           NVARCHAR(100)  NOT NULL,
    [Number]         NVARCHAR(50)   NOT NULL,
    [Description]    NVARCHAR(255)  NOT NULL,
    [ClosingDay]     INT            NOT NULL,
    [ExpirationDate] DATETIME2      NOT NULL,
    [CreatedAt]      DATETIME2      NOT NULL,
    [UpdatedAt]      DATETIME2      NOT NULL,
    [Active]         BIT            NOT NULL DEFAULT 1,
    CONSTRAINT FK_Card_Flag FOREIGN KEY ([FlagId]) REFERENCES dbo.[Flag]([Id])
)
`

const insertSeedFlag = `
IF NOT EXISTS (SELECT 1 FROM dbo.[Flag] WHERE [Id] = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa')
INSERT INTO dbo.[Flag] ([Id], [Name], [Active])
VALUES ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Visa', 1)
`

type CardRepositorySuite struct {
	suite.Suite

	mgr    devkitmgr.Manager
	dbtx   devkitdb.DBTX
	ctx    context.Context
	flagID vos.FlagID
}

func TestCardRepositorySuite(t *testing.T) {
	suite.Run(t, new(CardRepositorySuite))
}

func (s *CardRepositorySuite) SetupSuite() {
	db, _, err := dbmssql.GetSharedTestDatabase()
	s.Require().NoError(err, "failed to get shared test database")

	_, err = db.ExecContext(context.Background(), createFlagTable)
	s.Require().NoError(err, "failed to create flag table")

	_, err = db.ExecContext(context.Background(), createCardTable)
	s.Require().NoError(err, "failed to create card table")

	_, err = db.ExecContext(context.Background(), insertSeedFlag)
	s.Require().NoError(err, "failed to seed flag")

	flagID, err := vos.ParseFlagID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	s.Require().NoError(err)
	s.flagID = flagID

	mgr, _, err := dbmssql.GetSharedTestManager()
	s.Require().NoError(err, "failed to get shared test manager")
	s.mgr = mgr
}

func (s *CardRepositorySuite) SetupTest() {
	s.ctx = context.Background()
	s.dbtx = s.mgr.DBTX(s.ctx)

	db, _, err := dbmssql.GetSharedTestDatabase()
	s.Require().NoError(err)
	_, _ = db.ExecContext(s.ctx, "DELETE FROM dbo.[Card]")
}

func (s *CardRepositorySuite) TearDownTest() {
	db, _, err := dbmssql.GetSharedTestDatabase()
	if err == nil {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM dbo.[Card]")
	}
}

func (s *CardRepositorySuite) newCard(userID identityvo.UserID) *entities.Card {
	card, err := entities.NewCard(userID, s.flagID, "Meu Cartão", "123456789012", "desc", 10, time.Date(2026, time.May, 20, 0, 0, 0, 0, time.UTC))
	s.Require().NoError(err)
	return card
}

func (s *CardRepositorySuite) TestAddGetByIDUpdateList() {
	repo := repomssql.NewCardRepository(s.dbtx)
	userID := identityvo.NewUserID()

	card := s.newCard(userID)
	s.Require().NoError(repo.Add(s.ctx, card))

	got, err := repo.GetByID(s.ctx, userID, card.ID())
	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Equal(card.ID().String(), got.ID().String())
	s.Equal(card.Name().String(), got.Name().String())
	s.True(card.ExpirationDate().Equal(got.ExpirationDate()))

	err = card.Update(s.flagID, "Cartão Atualizado", "123456789013", "nova desc", 15, time.Date(2027, time.January, 25, 0, 0, 0, 0, time.UTC))
	s.Require().NoError(err)
	s.Require().NoError(repo.Update(s.ctx, card))

	updated, err := repo.GetByID(s.ctx, userID, card.ID())
	s.Require().NoError(err)
	s.Equal("Cartão Atualizado", updated.Name().String())
	s.True(card.ExpirationDate().Equal(updated.ExpirationDate()))

	list, err := repo.List(s.ctx, userID, dtos.Pagination{})
	s.Require().NoError(err)
	s.Len(list, 1)
}

func (s *CardRepositorySuite) TestIsolationByUserID() {
	repo := repomssql.NewCardRepository(s.dbtx)

	user1 := identityvo.NewUserID()
	user2 := identityvo.NewUserID()

	card1 := s.newCard(user1)
	s.Require().NoError(repo.Add(s.ctx, card1))

	scenarios := []struct {
		name   string
		userID identityvo.UserID
		cardID vos.CardID
		expect func(got *entities.Card, err error)
	}{
		{
			name:   "dono correto encontra o cartão",
			userID: user1,
			cardID: card1.ID(),
			expect: func(got *entities.Card, err error) {
				s.Require().NoError(err)
				s.Require().NotNil(got)
				s.Equal(card1.ID().String(), got.ID().String())
			},
		},
		{
			name:   "usuário diferente recebe ErrCardNotFound",
			userID: user2,
			cardID: card1.ID(),
			expect: func(got *entities.Card, err error) {
				s.Require().ErrorIs(err, domain.ErrCardNotFound)
				s.Nil(got)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			got, err := repo.GetByID(s.ctx, sc.userID, sc.cardID)
			sc.expect(got, err)
		})
	}
}

func (s *CardRepositorySuite) TestListIgnoresInactiveCards() {
	repo := repomssql.NewCardRepository(s.dbtx)
	userID := identityvo.NewUserID()

	active := s.newCard(userID)
	inactive := s.newCard(userID)
	s.Require().NoError(repo.Add(s.ctx, active))
	s.Require().NoError(repo.Add(s.ctx, inactive))

	inactive.Deactivate()
	s.Require().NoError(repo.Update(s.ctx, inactive))

	list, err := repo.List(s.ctx, userID, dtos.Pagination{})
	s.Require().NoError(err)
	s.Len(list, 1)
	s.Equal(active.ID().String(), list[0].ID().String())
}

func (s *CardRepositorySuite) TestListWithoutPaginationReturnsAllCards() {
	repo := repomssql.NewCardRepository(s.dbtx)
	userID := identityvo.NewUserID()

	first := s.newCard(userID)
	second := s.newCard(userID)
	s.Require().NoError(repo.Add(s.ctx, first))
	s.Require().NoError(repo.Add(s.ctx, second))

	list, err := repo.List(s.ctx, userID, dtos.Pagination{})
	s.Require().NoError(err)
	s.Len(list, 2)
}

func (s *CardRepositorySuite) TestListWithPaginationStillSupportsSlices() {
	repo := repomssql.NewCardRepository(s.dbtx)
	userID := identityvo.NewUserID()

	first := s.newCard(userID)
	second := s.newCard(userID)
	s.Require().NoError(repo.Add(s.ctx, first))
	s.Require().NoError(repo.Add(s.ctx, second))

	list, err := repo.List(s.ctx, userID, dtos.NewPagination(1, 1))
	s.Require().NoError(err)
	s.Len(list, 1)
}

func (s *CardRepositorySuite) TestUpdateFiltersByUserID() {
	repo := repomssql.NewCardRepository(s.dbtx)
	owner := identityvo.NewUserID()
	attacker := identityvo.NewUserID()

	original := s.newCard(owner)
	s.Require().NoError(repo.Add(s.ctx, original))

	tampered, err := entities.RehydrateCard(
		original.ID(),
		attacker,
		original.FlagID(),
		"Cartão Sequestrado",
		original.Number().String(),
		"hacked",
		original.ClosingDay().Int(),
		original.ExpirationDate(),
		original.CreatedAt(),
		time.Now().UTC(),
		original.Active(),
	)
	s.Require().NoError(err)

	s.Require().NoError(repo.Update(s.ctx, tampered))

	persisted, err := repo.GetByID(s.ctx, owner, original.ID())
	s.Require().NoError(err)
	s.Equal(original.Name().String(), persisted.Name().String(),
		"row owned by another user must remain untouched")
	s.Equal(original.Description(), persisted.Description())
}

func (s *CardRepositorySuite) TestGetByIDNotFound() {
	repo := repomssql.NewCardRepository(s.dbtx)
	userID := identityvo.NewUserID()
	nonExistentID := vos.NewCardID()

	got, err := repo.GetByID(s.ctx, userID, nonExistentID)
	s.Require().ErrorIs(err, domain.ErrCardNotFound)
	s.Nil(got)
}
