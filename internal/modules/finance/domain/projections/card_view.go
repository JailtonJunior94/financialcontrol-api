package projections

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/vos"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"
)

// CardView is a read-only projection of a card's data used within the finance domain.
// It contains only the fields required by finance domain rules (RF-22).
// Returned by CardProvider — never import cards module directly from domain.
type CardView struct {
	ID           vos.CardID
	UserID       identityvo.UserID
	FlagName     string
	ClosingDay   int
	DueDay       int
	BillingCycle int
	Active       bool
}
