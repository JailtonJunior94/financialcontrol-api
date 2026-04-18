package mappings

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
)

func ToManyCategoryResponse(entities []entities.Category) (r []responses.CategoryResponse) {
	for _, e := range entities {
		item := responses.CategoryResponse{
			ID:       e.ID,
			Name:     e.Name,
			Sequence: e.Sequence,
			Active:   e.Active,
		}
		r = append(r, item)
	}

	return r
}
