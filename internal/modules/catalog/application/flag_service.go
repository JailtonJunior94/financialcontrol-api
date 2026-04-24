package application

import "github.com/jailtonjunior94/financialcontrol-api/internal/platform/web"

type DefaultFlagService struct {
	repository FlagRepository
}

func NewFlagService(repository FlagRepository) FlagService {
	return &DefaultFlagService{repository: repository}
}

func (s *DefaultFlagService) Flags() *web.HttpResponse {
	flags, err := s.repository.GetFlags()
	if err != nil {
		return web.ServerError()
	}

	return web.Ok(ToManyFlagResponse(flags))
}
