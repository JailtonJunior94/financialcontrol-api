package ports

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/projections"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// CategoryProvider is the read-only port used by the finance domain to access category data
// without depending on the categories module directly (RF-23, decisão G2.a).
type CategoryProvider interface {
	GetByID(ctx context.Context, userID identityvo.UserID, categoryID vos.CategoryID) (projections.CategoryView, error)
}
