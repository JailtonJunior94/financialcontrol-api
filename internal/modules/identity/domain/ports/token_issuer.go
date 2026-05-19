package ports

import (
	"context"
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
)

type TokenIssuer interface {
	Issue(ctx context.Context, userID vos.UserID, email vos.Email) (token string, expiresAt time.Time, err error)
}
