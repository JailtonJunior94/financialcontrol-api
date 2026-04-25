package http

import platformsecurity "github.com/jailtonjunior94/financialcontrol-api/pkg/security"

type ClaimsResolver interface {
	UserID(authorizationHeader string) (string, error)
}

type claimsResolver struct {
	jwtAdapter platformsecurity.TokenAdapter
}

func NewClaimsResolver(jwtAdapter platformsecurity.TokenAdapter) ClaimsResolver {
	return &claimsResolver{jwtAdapter: jwtAdapter}
}

func (r *claimsResolver) UserID(authorizationHeader string) (string, error) {
	userID, err := r.jwtAdapter.ExtractClaims(authorizationHeader)
	if err != nil {
		return "", err
	}

	return *userID, nil
}
