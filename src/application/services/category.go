package services

import (
	"github.com/jailtonjunior94/financialcontrol-api/src/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/src/application/mappings"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/src/domain/usecases"
)

type CategoryService struct {
	CategoryRepository interfaces.ICategoryRepository
}

func NewCategoryService(r interfaces.ICategoryRepository) usecases.ICategoryService {
	return &CategoryService{CategoryRepository: r}
}

func (s *CategoryService) Categories() *responses.HttpResponse {
	categories, err := s.CategoryRepository.GetCategories()
	if err != nil {
		return responses.ServerError()
	}

	return responses.Ok(mappings.ToManyCategoryResponse(categories))
}
