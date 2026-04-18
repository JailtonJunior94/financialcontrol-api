package services

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/mappings"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/interfaces"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/usecases"
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
