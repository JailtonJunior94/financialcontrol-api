package projections

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// CategoryView is a read-only projection of a category's data used within the finance domain.
// Returned by CategoryProvider — never import categories module directly from domain (RF-23).
type CategoryView struct {
	ID       vos.CategoryID
	UserID   identityvo.UserID
	Name     string
	ParentID *vos.CategoryID
	Active   bool
}
