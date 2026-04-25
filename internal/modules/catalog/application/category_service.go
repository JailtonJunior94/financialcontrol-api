package application

import "github.com/jailtonjunior94/financialcontrol-api/pkg/web"

type DefaultCategoryService struct {
	repository CategoryRepository
}

func NewCategoryService(repository CategoryRepository) CategoryService {
	return &DefaultCategoryService{repository: repository}
}

func (s *DefaultCategoryService) Categories() *web.HttpResponse {
	categories, err := s.repository.GetCategories()
	if err != nil {
		return web.ServerError()
	}

	return web.Ok(ToManyCategoryResponse(categories))
}
