package application

import "github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"

func ToManyFlagResponse(items []entities.Flag) []FlagResponse {
	if len(items) == 0 {
		return make([]FlagResponse, 0)
	}

	result := make([]FlagResponse, 0, len(items))
	for _, item := range items {
		result = append(result, FlagResponse{
			ID:     item.ID,
			Name:   item.Name,
			Active: item.Active,
		})
	}

	return result
}

func ToManyCategoryResponse(items []entities.Category) []CategoryResponse {
	if len(items) == 0 {
		return make([]CategoryResponse, 0)
	}

	result := make([]CategoryResponse, 0, len(items))
	for _, item := range items {
		result = append(result, CategoryResponse{
			ID:       item.ID,
			Name:     item.Name,
			Sequence: item.Sequence,
			Active:   item.Active,
		})
	}

	return result
}
