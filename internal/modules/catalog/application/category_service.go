package application

import "github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"

type DefaultCategoryService struct {
	repository CategoryRepository
}

func NewCategoryService(repository CategoryRepository) CategoryService {
	return &DefaultCategoryService{repository: repository}
}

func (s *DefaultCategoryService) Categories() *responses.HttpResponse {
	categories, err := s.repository.GetCategories()
	if err != nil {
		return responses.ServerError()
	}

	return responses.Ok(ToManyCategoryResponse(categories))
}
