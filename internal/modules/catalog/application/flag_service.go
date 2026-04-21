package application

import "github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"

type DefaultFlagService struct {
	repository FlagRepository
}

func NewFlagService(repository FlagRepository) FlagService {
	return &DefaultFlagService{repository: repository}
}

func (s *DefaultFlagService) Flags() *responses.HttpResponse {
	flags, err := s.repository.GetFlags()
	if err != nil {
		return responses.ServerError()
	}

	return responses.Ok(ToManyFlagResponse(flags))
}
