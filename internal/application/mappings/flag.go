package mappings

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
)

func ToManyFlagResponse(entities []entities.Flag) (r []responses.FlagResponse) {
	for _, e := range entities {
		item := responses.FlagResponse{
			ID:     e.ID,
			Name:   e.Name,
			Active: e.Active,
		}
		r = append(r, item)
	}

	return r
}
