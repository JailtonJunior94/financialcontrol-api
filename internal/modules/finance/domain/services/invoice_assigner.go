package services

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/projections"
)

// InvoiceAssigner determines the billing cycle for a transaction based on the
// card's cutoff day in America/Sao_Paulo timezone (RF-15).
type InvoiceAssigner struct{}

func NewInvoiceAssigner() *InvoiceAssigner { return &InvoiceAssigner{} }

// AssignFor returns the open invoice that matches the billing cycle for occurredAt,
// applying the cutoff rule in America/Sao_Paulo (RF-15). Returns nil when no existing
// open invoice matches; the caller must create one via InvoiceRepository.AssignOrCreateOpen.
//
// Cutoff rule: if occurredAt day-in-SP >= card.ClosingDay, the purchase belongs to the
// NEXT month's cycle; otherwise it belongs to the CURRENT month's cycle.
func (a *InvoiceAssigner) AssignFor(
	card projections.CardView,
	occurredAt time.Time,
	openInvoices []*entities.Invoice,
	_ ports.Clock,
) (*entities.Invoice, error) {
	saoPaulo, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return nil, err
	}
	spTime := occurredAt.In(saoPaulo)
	targetClosing := computeTargetClosingDate(spTime, card.ClosingDay, saoPaulo)

	for _, inv := range openInvoices {
		closingSP := inv.ClosingDate().In(saoPaulo)
		if sameDate(closingSP, targetClosing) {
			return inv, nil
		}
	}
	return nil, nil
}

// computeTargetClosingDate returns the closing date (in saoPaulo location) for the
// billing cycle that occurredAt belongs to.
func computeTargetClosingDate(spTime time.Time, closingDay int, loc *time.Location) time.Time {
	day := spTime.Day()
	year := spTime.Year()
	month := spTime.Month()

	if day >= closingDay {
		month++
		if month > 12 {
			month = 1
			year++
		}
	}
	maxDay := daysInMonth(year, month, loc)
	if closingDay > maxDay {
		closingDay = maxDay
	}
	return time.Date(year, month, closingDay, 0, 0, 0, 0, loc)
}

func daysInMonth(year int, month time.Month, loc *time.Location) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
}

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}
