package adapters

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
)

type tokenIssuer struct {
	inner pkgjwt.Issuer
}

func NewTokenIssuer(inner pkgjwt.Issuer) ports.TokenIssuer {
	return &tokenIssuer{inner: inner}
}

func (a *tokenIssuer) Issue(ctx context.Context, userID vos.UserID, email vos.Email) (string, time.Time, error) {
	return a.inner.Issue(ctx, pkgjwt.Identity{
		UserID: userID.String(),
		Email:  email.String(),
	})
}
