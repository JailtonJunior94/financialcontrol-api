package invoicingapp

import (
	"time"

	invoicingdomain "github.com/jailtonjunior94/financialcontrol-api/internal/modules/invoicing/domain"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/shared"
)

func ToInvoiceEntity(request *InvoiceRequest, date time.Time, total float64) *invoicingdomain.Invoice {
	return invoicingdomain.NewInvoice(request.CardId, date, total)
}

func ToInvoiceItemEntity(request *InvoiceRequest, invoiceID string, installment int, invoiceControl int64) *invoicingdomain.InvoiceItem {
	entity := invoicingdomain.NewInvoiceItem(
		invoiceID,
		request.CategoryId,
		request.Description,
		request.Tags,
		request.PurchaseDate,
		request.TotalAmount,
	)
	entity.AddInstallment(installment, request.TotalAmount/float64(request.QuantityInvoice), invoiceControl)
	return entity
}

func ToInvoiceResponse(entity *invoicingdomain.Invoice) *InvoiceResponse {
	if entity == nil {
		return nil
	}

	return &InvoiceResponse{
		ID:           entity.ID,
		CardId:       entity.CardId,
		Date:         shared.NewTime(shared.Time{Date: entity.Date}).FormatDate(),
		Total:        entity.Total,
		Card:         toCardMinimalResponse(&entity.Card),
		InvoiceItems: ToManyInvoiceItemResponse(entity.InvoiceItems),
	}
}

func ToManyInvoiceResponse(entities []invoicingdomain.Invoice) []*InvoiceResponse {
	response := make([]*InvoiceResponse, len(entities))
	for i, entity := range entities {
		response[i] = &InvoiceResponse{
			ID:     entity.ID,
			CardId: entity.CardId,
			Date:   shared.NewTime(shared.Time{Date: entity.Date}).FormatDate(),
			Total:  entity.Total,
		}
	}

	return response
}

func ToManyInvoiceItemResponse(entities []invoicingdomain.InvoiceItem) []*InvoiceItemResponse {
	response := make([]*InvoiceItemResponse, len(entities))
	for i, entity := range entities {
		response[i] = &InvoiceItemResponse{
			ID:               entity.ID,
			InvoiceControl:   entity.InvoiceControl,
			PurchaseDate:     shared.NewTime(shared.Time{Date: entity.PurchaseDate}).FormatDate(),
			Description:      entity.Description,
			TotalAmount:      entity.TotalAmount,
			Installment:      entity.Installment,
			InstallmentValue: entity.InstallmentValue,
			Tags:             entity.Tags,
			Category: CategoryResponse{
				ID:     entity.Category.ID,
				Name:   entity.Category.Name,
				Active: entity.Category.Active,
			},
		}
	}

	return response
}

func toCardMinimalResponse(entity *invoicingdomain.Card) *CardMinimalResponse {
	return &CardMinimalResponse{
		ID:     entity.ID,
		Name:   entity.Name,
		Active: entity.Active,
	}
}
