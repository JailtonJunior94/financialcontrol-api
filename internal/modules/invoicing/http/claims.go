package http

import platformsecurity "github.com/jailtonjunior94/financialcontrol-api/internal/platform/security"

type ClaimsResolver interface {
	UserID(authorizationHeader string) (string, error)
}

type tokenClaimsResolver struct {
	tokenAdapter platformsecurity.TokenAdapter
}

func NewClaimsResolver(tokenAdapter platformsecurity.TokenAdapter) ClaimsResolver {
	return &tokenClaimsResolver{tokenAdapter: tokenAdapter}
}

func (r *tokenClaimsResolver) UserID(authorizationHeader string) (string, error) {
	userID, err := r.tokenAdapter.ExtractClaims(authorizationHeader)
	if err != nil {
		return "", err
	}

	return *userID, nil
}
