package invoicingapp

import (
	"time"

	appresponses "github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
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

func ToInvoiceResponse(entity *invoicingdomain.Invoice) *appresponses.InvoiceResponse {
	if entity == nil {
		return nil
	}

	return &appresponses.InvoiceResponse{
		ID:           entity.ID,
		CardId:       entity.CardId,
		Date:         shared.NewTime(shared.Time{Date: entity.Date}).FormatDate(),
		Total:        entity.Total,
		Card:         toCardMinimalResponse(&entity.Card),
		InvoiceItems: ToManyInvoiceItemResponse(entity.InvoiceItems),
	}
}

func ToManyInvoiceResponse(entities []invoicingdomain.Invoice) []*appresponses.InvoiceResponse {
	response := make([]*appresponses.InvoiceResponse, len(entities))
	for i, entity := range entities {
		response[i] = &appresponses.InvoiceResponse{
			ID:     entity.ID,
			CardId: entity.CardId,
			Date:   shared.NewTime(shared.Time{Date: entity.Date}).FormatDate(),
			Total:  entity.Total,
		}
	}

	return response
}

func ToManyInvoiceItemResponse(entities []invoicingdomain.InvoiceItem) []*appresponses.InvoiceItemResponse {
	response := make([]*appresponses.InvoiceItemResponse, len(entities))
	for i, entity := range entities {
		response[i] = &appresponses.InvoiceItemResponse{
			ID:               entity.ID,
			InvoiceControl:   entity.InvoiceControl,
			PurchaseDate:     shared.NewTime(shared.Time{Date: entity.PurchaseDate}).FormatDate(),
			Description:      entity.Description,
			TotalAmount:      entity.TotalAmount,
			Installment:      entity.Installment,
			InstallmentValue: entity.InstallmentValue,
			Tags:             entity.Tags,
			Category: appresponses.CategoryResponse{
				ID:     entity.Category.ID,
				Name:   entity.Category.Name,
				Active: entity.Category.Active,
			},
		}
	}

	return response
}

func toCardMinimalResponse(entity *invoicingdomain.Card) *appresponses.CardMinimalResponse {
	return &appresponses.CardMinimalResponse{
		ID:     entity.ID,
		Name:   entity.Name,
		Active: entity.Active,
	}
}
