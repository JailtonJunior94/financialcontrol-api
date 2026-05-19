package services_test

import (
	"testing"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/ports/mocks"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/services"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"

	"github.com/stretchr/testify/suite"
)

type CredentialsCheckerSuite struct {
	suite.Suite
	hasher *mocks.Hasher
	sut    *services.CredentialsChecker
}

func TestCredentialsCheckerSuite(t *testing.T) { suite.Run(t, new(CredentialsCheckerSuite)) }

func (s *CredentialsCheckerSuite) SetupTest() {
	s.hasher = mocks.NewHasher(s.T())
	s.sut = services.NewCredentialsChecker(s.hasher)
}

func (s *CredentialsCheckerSuite) TestCheck_ReturnsTrue_WhenPasswordMatches() {
	email, _ := vos.NewEmail("user@example.com")
	hashedPwd, _ := vos.NewHashedPassword("$2a$10$hash")
	now := time.Now()
	user := entities.RehydrateUser(vos.NewUserID(), "User", email, hashedPwd, now, now, true)

	s.hasher.EXPECT().Verify(hashedPwd, "plain-password").Return(true)

	result := s.sut.Check(user, "plain-password")
	s.True(result)
}

func (s *CredentialsCheckerSuite) TestCheck_ReturnsFalse_WhenPasswordDoesNotMatch() {
	email, _ := vos.NewEmail("user@example.com")
	hashedPwd, _ := vos.NewHashedPassword("$2a$10$hash")
	now := time.Now()
	user := entities.RehydrateUser(vos.NewUserID(), "User", email, hashedPwd, now, now, true)

	s.hasher.EXPECT().Verify(hashedPwd, "wrong-password").Return(false)

	result := s.sut.Check(user, "wrong-password")
	s.False(result)
}
