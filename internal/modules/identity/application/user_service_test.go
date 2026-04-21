package application_test

import (
	"testing"

	appresponses "github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	identityapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application"

	"github.com/stretchr/testify/require"
)

func TestCreateUserStoresGeneratedPasswordHash(t *testing.T) {
	repository := &userRepositoryStub{}
	service := identityapp.NewUserService(
		repository,
		&hashAdapterStub{generateHash: "hashed-password"},
	)

	response := service.CreateUser(&identityapp.UserRequest{
		Name:     "John",
		Email:    "john@example.com",
		Password: "plain-password",
	})

	require.Equal(t, appresponses.Created(nil).StatusCode, response.StatusCode)
	require.NotNil(t, repository.addedUser)
	require.Equal(t, "hashed-password", repository.addedUser.Password)
}
