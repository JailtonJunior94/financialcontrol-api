package http

import "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/application"

type ClaimsResolver interface {
	UserID(authorizationHeader string) (string, error)
}

type AuthorizationClaimsResolver struct {
	tokenAdapter application.TokenAdapter
}

func NewClaimsResolver(tokenAdapter application.TokenAdapter) ClaimsResolver {
	return &AuthorizationClaimsResolver{tokenAdapter: tokenAdapter}
}

func (r *AuthorizationClaimsResolver) UserID(authorizationHeader string) (string, error) {
	userID, err := r.tokenAdapter.ExtractClaims(authorizationHeader)
	if err != nil {
		return "", err
	}

	return *userID, nil
}
