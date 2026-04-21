package application

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	identitydomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain"
)

type UserRepository interface {
	Add(user *identitydomain.User) (*identitydomain.User, error)
	GetByEmail(email string) (*identitydomain.User, error)
	GetByID(id string) (*identitydomain.User, error)
}

type HashAdapter interface {
	GenerateHash(str string) (string, error)
	CheckHash(hash, str string) bool
}

type TokenAdapter interface {
	GenerateTokenJWT(id, email string) (string, error)
	ExtractClaims(tokenString string) (*string, error)
}

type AuthService interface {
	Authenticate(request *AuthRequest) *responses.HttpResponse
	Me(userID string) *responses.HttpResponse
}

type UserService interface {
	CreateUser(request *UserRequest) *responses.HttpResponse
}
