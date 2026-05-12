package ports

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/projections"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// CardProvider is the read-only port used by the finance domain to access card data
// without depending on the cards module directly (RF-22, decisão A3.c).
type CardProvider interface {
	GetByID(ctx context.Context, userID identityvo.UserID, cardID vos.CardID) (projections.CardView, error)
}
