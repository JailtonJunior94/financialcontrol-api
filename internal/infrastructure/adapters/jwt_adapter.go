package adapters

import (
	"fmt"
	"strings"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/config"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/customerrors"

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
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(config.ExpirationAt)).Unix()

	t, err := token.SignedString([]byte(config.JwtSecret))
	if err != nil {
		return "", err
	}
	return t, nil
}

func (j *JwtAdapter) ExtractClaims(tokenString string) (id *string, err error) {
	tokenString = strings.Split(tokenString, " ")[1]
	hmacSecret := []byte(config.JwtSecret)

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return hmacSecret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, customerrors.InvalidToken
	}

	sub := fmt.Sprintf("%v", claims["sub"])
	return &sub, nil
}
