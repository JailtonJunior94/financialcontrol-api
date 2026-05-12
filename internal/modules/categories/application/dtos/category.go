package dtos

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/entities"
)

// CategoryRequest is the create/update payload.
type CategoryRequest struct {
	Name     string  `json:"name" validate:"required,max=100"`
	Sequence int     `json:"sequence"`
	Color    string  `json:"color,omitempty"`
	Icon     string  `json:"icon,omitempty"`
	ParentID *string `json:"parentId,omitempty" validate:"omitempty,uuid"`
}

type CategoryResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Sequence  int       `json:"sequence"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Active    bool      `json:"active"`
}

func ToCategoryResponse(c *entities.Category) CategoryResponse {
	return CategoryResponse{
		ID:        c.ID().String(),
		Name:      c.Name().String(),
		Sequence:  c.Sequence(),
		CreatedAt: c.CreatedAt(),
		UpdatedAt: c.UpdatedAt(),
		Active:    c.IsActive(),
	}
}

func ToCategoryResponses(items []entities.Category) []CategoryResponse {
	out := make([]CategoryResponse, len(items))
	for i := range items {
		out[i] = ToCategoryResponse(&items[i])
	}
	return out
}
