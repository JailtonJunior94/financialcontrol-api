package usecase

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/application/dtos"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/finance/domain/entities"
)

func toTransactionResponse(t *entities.Transaction) dtos.TransactionResponse {
	resp := dtos.TransactionResponse{
		ID:              t.ID().String(),
		Description:     t.Description(),
		Amount:          t.Amount().String(),
		OccurredAt:      t.OccurredAt(),
		TransactionType: t.TransactionType().String(),
		PaymentMethod:   t.PaymentMethod().String(),
		CategoryID:      t.CategoryID().String(),
		LegacyOrigin:    t.LegacyOrigin(),
		CreatedAt:       t.CreatedAt(),
		UpdatedAt:       t.UpdatedAt(),
	}
	if t.CardID() != nil {
		s := t.CardID().String()
		resp.CardID = &s
	}
	if t.SubcategoryID() != nil {
		s := t.SubcategoryID().String()
		resp.SubcategoryID = &s
	}
	if t.OriginalTransactionID() != nil {
		s := t.OriginalTransactionID().String()
		resp.OriginalTransactionID = &s
	}
	for _, inst := range t.Installments() {
		resp.Installments = append(resp.Installments, toInstallmentResponse(inst))
	}
	return resp
}

func toInstallmentResponse(i *entities.Installment) dtos.InstallmentResponse {
	return dtos.InstallmentResponse{
		ID:            i.ID().String(),
		TransactionID: i.TransactionID().String(),
		InvoiceID:     i.InvoiceID().String(),
		Number:        i.Number().Value(),
		Total:         i.Total().Value(),
		Amount:        i.Amount().String(),
		Status:        i.Status().String(),
		CreatedAt:     i.CreatedAt(),
		UpdatedAt:     i.UpdatedAt(),
	}
}

func toInvoiceResponse(inv *entities.Invoice) dtos.InvoiceResponse {
	return dtos.InvoiceResponse{
		ID:          inv.ID().String(),
		UserID:      inv.UserID().String(),
		CardID:      inv.CardID().String(),
		State:       inv.State().String(),
		CycleStart:  inv.CycleStart(),
		CycleEnd:    inv.CycleEnd(),
		ClosingDate: inv.ClosingDate(),
		DueDate:     inv.DueDate(),
		Total:       inv.Total().String(),
		PaidAt:      inv.PaidAt(),
		CreatedAt:   inv.CreatedAt(),
		UpdatedAt:   inv.UpdatedAt(),
	}
}

func toInvoiceItemResponse(inst *entities.Installment) dtos.InvoiceItemResponse {
	return dtos.InvoiceItemResponse{
		InstallmentID: inst.ID().String(),
		TransactionID: inst.TransactionID().String(),
		Number:        inst.Number().Value(),
		Total:         inst.Total().Value(),
		Amount:        inst.Amount().String(),
		Status:        inst.Status().String(),
	}
}
