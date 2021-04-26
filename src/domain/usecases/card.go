package usecases

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/requests"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
)

type ICardService interface {
	Cards(userId string) *responses.HttpResponse
	CardById(id, userId string) *responses.HttpResponse
	CreateCard(userId string, request *requests.CardRequest) *responses.HttpResponse
	UpdateCard(id, userId string, request *requests.CardRequest) *responses.HttpResponse
	RemoveCard(id, userId string) *responses.HttpResponse
}
