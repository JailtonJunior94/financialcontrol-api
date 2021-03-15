package adapters

import (
	"fmt"
	"strings"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/src/domain/customErrors"
	"github.com/jailtonjunior94/financialcontrol-api/src/infrastructure/environments"

	"github.com/dgrijalva/jwt-go"
)

type IJwtAdapter interface {
	GenerateTokenJWT(id, email string) (r string, err error)
	ExtractClaims(tokenString string) (id *string, err error)
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
	claims["exp"] = time.Now().Local().Add(time.Hour * 24 * time.Duration(environments.ExpirationAt))

	t, err := token.SignedString([]byte(environments.JwtSecret))
	if err != nil {
		return "", err
	}
	return t, err
}

func (j *JwtAdapter) ExtractClaims(tokenString string) (id *string, err error) {
	tokenString = strings.Split(tokenString, " ")[1]
	hmacSecret := []byte(environments.JwtSecret)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return hmacSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, customErrors.InvalidToken
	}

	sub := fmt.Sprintf("%v", claims["sub"])
	return &sub, nil
}
