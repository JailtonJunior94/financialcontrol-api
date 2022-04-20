package services_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/services"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"

	"github.com/stretchr/testify/mock"
)

type AuthServiceMock struct {
	mock.Mock
}

func (m *AuthServiceMock) GetByEmail(email string) (user *entities.User, err error) {
	fmt.Println("Mocked GetByEmail")
	fmt.Printf("Params: %s\n", email)

	args := m.Called(email)

	i := args.Get(0)
	if i == nil {
		return nil, args.Error(1)
	}

	s := i.(*entities.User)
	return s, args.Error(1)
}

func (m *AuthServiceMock) GetByID(email string) (user *entities.User, err error) {
	return nil, nil
}

func (m *AuthServiceMock) Add(p *entities.User) (user *entities.User, err error) {
	return nil, nil
}

func TestGetByEmailError(t *testing.T) {
	/* Arrange */
	authServiceMock := new(AuthServiceMock)

	service := services.NewAuthService(authServiceMock, nil, nil)
	authServiceMock.On("GetByEmail", mock.Anything).Return(nil, errors.New("Nenhum usuário encontrado"))

	/* Act */
	service.Authenticate(&requests.AuthRequest{
		Email:    "meuemail@email.com",
		Password: "minha senha",
	})

	/* Assert */
	authServiceMock.AssertExpectations(t)
}

func TestGetByEmailNotFound(t *testing.T) {
	/* Arrange */
	authServiceMock := new(AuthServiceMock)

	service := services.NewAuthService(authServiceMock, nil, nil)
	authServiceMock.On("GetByEmail", mock.Anything).Return(nil, nil)

	/* Act */
	service.Authenticate(&requests.AuthRequest{
		Email:    "meuemail@email.com",
		Password: "minha senha",
	})

	/* Assert */
	authServiceMock.AssertExpectations(t)
}
