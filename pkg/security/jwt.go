package security

import (
	"fmt"
	"strings"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/config"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/customerrors"

	"github.com/dgrijalva/jwt-go"
)

type TokenAdapter interface {
	GenerateTokenJWT(id, email string) (string, error)
	ExtractClaims(tokenString string) (*string, error)
}

type JWTAdapter struct{}

func NewJWTAdapter() TokenAdapter {
	return &JWTAdapter{}
}

func (j *JWTAdapter) GenerateTokenJWT(id, email string) (string, error) {
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

func (j *JWTAdapter) ExtractClaims(tokenString string) (*string, error) {
	parts := strings.Split(tokenString, " ")
	if len(parts) < 2 {
		return nil, customerrors.InvalidToken
	}

	tokenString = parts[1]
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
