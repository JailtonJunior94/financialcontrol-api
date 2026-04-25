package application_test

import (
	"errors"
	"testing"

	appresponses "github.com/jailtonjunior94/financialcontrol-api/pkg/web"
	identityapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application"
	identitydomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"

	"github.com/stretchr/testify/require"
)

type userRepositoryStub struct {
	userByEmail *identitydomain.User
	userByID    *identitydomain.User
	addedUser   *identitydomain.User
	emailErr    error
	idErr       error
}

func (s *userRepositoryStub) Add(user *identitydomain.User) (*identitydomain.User, error) {
	s.addedUser = user
	return user, nil
}

func (s *userRepositoryStub) GetByEmail(email string) (*identitydomain.User, error) {
	return s.userByEmail, s.emailErr
}

func (s *userRepositoryStub) GetByID(id string) (*identitydomain.User, error) {
	return s.userByID, s.idErr
}

type hashAdapterStub struct {
	checkHashResult bool
	generateHash    string
	generateErr     error
}

func (s *hashAdapterStub) GenerateHash(str string) (string, error) {
	return s.generateHash, s.generateErr
}

func (s *hashAdapterStub) CheckHash(hash, str string) bool {
	return s.checkHashResult
}

type tokenAdapterStub struct {
	token string
	err   error
}

func (s *tokenAdapterStub) GenerateTokenJWT(id, email string) (string, error) {
	return s.token, s.err
}

func (s *tokenAdapterStub) ExtractClaims(tokenString string) (*string, error) {
	if s.err != nil {
		return nil, s.err
	}

	value := "user-id"
	return &value, nil
}

func TestAuthenticateReturnsTokenWhenCredentialsAreValid(t *testing.T) {
	service := identityapp.NewAuthService(
		&userRepositoryStub{
			userByEmail: &identitydomain.User{
				Name:     "John",
				Email:    "john@example.com",
				Password: "stored-hash",
			},
		},
		&hashAdapterStub{checkHashResult: true},
		&tokenAdapterStub{token: "jwt-token"},
	)

	response := service.Authenticate(&identityapp.AuthRequest{
		Email:    "john@example.com",
		Password: "secret",
	})

	require.Equal(t, appresponses.Ok(nil).StatusCode, response.StatusCode)

	payload, ok := response.Data.(*identityapp.AuthResponse)
	require.True(t, ok)
	require.Equal(t, "jwt-token", payload.Token)
}

func TestAuthenticateReturnsBadRequestWhenPasswordDoesNotMatch(t *testing.T) {
	service := identityapp.NewAuthService(
		&userRepositoryStub{
			userByEmail: &identitydomain.User{
				Email:    "john@example.com",
				Password: "stored-hash",
			},
		},
		&hashAdapterStub{checkHashResult: false},
		&tokenAdapterStub{token: "jwt-token"},
	)

	response := service.Authenticate(&identityapp.AuthRequest{
		Email:    "john@example.com",
		Password: "wrong",
	})

	require.Equal(t, appresponses.BadRequest("").StatusCode, response.StatusCode)
}

func TestAuthenticateReturnsServerErrorWhenRepositoryFails(t *testing.T) {
	service := identityapp.NewAuthService(
		&userRepositoryStub{emailErr: errors.New("db error")},
		&hashAdapterStub{checkHashResult: true},
		&tokenAdapterStub{token: "jwt-token"},
	)

	response := service.Authenticate(&identityapp.AuthRequest{
		Email:    "john@example.com",
		Password: "secret",
	})

	require.Equal(t, appresponses.ServerError().StatusCode, response.StatusCode)
}

func TestMeReturnsUserWhenRepositoryFindsIdentity(t *testing.T) {
	service := identityapp.NewAuthService(
		&userRepositoryStub{
			userByID: &identitydomain.User{
				Name:  "John",
				Email: "john@example.com",
			},
		},
		&hashAdapterStub{checkHashResult: true},
		&tokenAdapterStub{},
	)

	response := service.Me("user-id")

	require.Equal(t, appresponses.Ok(nil).StatusCode, response.StatusCode)

	payload, ok := response.Data.(*identityapp.UserResponse)
	require.True(t, ok)
	require.Equal(t, "john@example.com", payload.Email)
}
