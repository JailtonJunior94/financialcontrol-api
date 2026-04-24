package http

import "github.com/jailtonjunior94/financialcontrol-api/internal/infrastructure/adapters"

type ClaimsResolver interface {
	UserID(authorizationHeader string) (string, error)
}

type claimsResolver struct {
	jwtAdapter adapters.IJwtAdapter
}

func NewClaimsResolver(jwtAdapter adapters.IJwtAdapter) ClaimsResolver {
	return &claimsResolver{jwtAdapter: jwtAdapter}
}

func (r *claimsResolver) UserID(authorizationHeader string) (string, error) {
	userID, err := r.jwtAdapter.ExtractClaims(authorizationHeader)
	if err != nil {
		return "", err
	}

	return *userID, nil
}
