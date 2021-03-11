package adapters

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/app/configuration"

	"github.com/dgrijalva/jwt-go"
)

type IJwtAdapter interface {
	GenerateTokenJWT(id, email string) (r string, err error)
}

type JwtAdapter struct {
}

func NewJwtAdapter() IJwtAdapter {
	return &JwtAdapter{}
}

func (j *JwtAdapter) GenerateTokenJWT(id, email string) (r string, err error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["sub"] = id
	claims["email"] = email
	claims["exp"] = time.Now().Local().Add(time.Hour * 24 * time.Duration(configuration.ExpirationAt))

	t, err := token.SignedString([]byte(configuration.JwtSecret))
	if err != nil {
		return "", err
	}
	return t, err
}
