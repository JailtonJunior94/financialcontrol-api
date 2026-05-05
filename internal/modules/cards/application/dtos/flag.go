package dtos

import "github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/entities"

type FlagResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

func ToFlagResponse(f entities.Flag) FlagResponse {
	return FlagResponse{
		ID:     f.ID().String(),
		Name:   f.Name(),
		Active: f.Active(),
	}
}

func ToFlagResponses(flags []entities.Flag) []FlagResponse {
	result := make([]FlagResponse, len(flags))
	for i, f := range flags {
		result[i] = ToFlagResponse(f)
	}
	return result
}
