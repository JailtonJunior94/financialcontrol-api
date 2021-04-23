package services

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/mappings"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
)

type FlagService struct {
	FlagRepository interfaces.IFlagRepository
}

func NewFlagService(r interfaces.IFlagRepository) usecases.IFlagService {
	return &FlagService{FlagRepository: r}
}

func (s *FlagService) Flags() *responses.HttpResponse {
	flags, err := s.FlagRepository.GetFlags()
	if err != nil {
		return responses.ServerError()
	}

	return responses.Ok(mappings.ToManyFlagResponse(flags))
}
