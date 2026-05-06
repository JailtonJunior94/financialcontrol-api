package dtos

import (
	"time"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/categories/domain/entities"
)

// CategoryRequest is the create/update payload.
type CategoryRequest struct {
	Name     string  `json:"name"     validate:"required,max=100"`
	Color    string  `json:"color"    validate:"required"`
	Icon     string  `json:"icon"     validate:"required,max=64"`
	ParentID *string `json:"parentId,omitempty" validate:"omitempty,uuid"`
}

// CategorySummary is a lightweight projection of a parent category for hydration.
type CategorySummary struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

type CategoryResponse struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Color     string           `json:"color"`
	Icon      string           `json:"icon"`
	ParentID  *string          `json:"parentId,omitempty"`
	Parent    *CategorySummary `json:"parent,omitempty"`
	CreatedAt time.Time        `json:"createdAt"`
	UpdatedAt time.Time        `json:"updatedAt"`
}

func ToCategoryResponse(c *entities.Category) CategoryResponse {
	resp := CategoryResponse{
		ID:        c.ID().String(),
		Name:      c.Name().String(),
		Color:     c.Color().String(),
		Icon:      c.Icon().String(),
		CreatedAt: c.CreatedAt(),
		UpdatedAt: c.UpdatedAt(),
	}
	if pid := c.ParentID(); pid != nil {
		s := pid.String()
		resp.ParentID = &s
	}
	return resp
}

func ToCategoryResponses(items []entities.Category) []CategoryResponse {
	out := make([]CategoryResponse, len(items))
	for i := range items {
		out[i] = ToCategoryResponse(&items[i])
	}
	return out
}

func ToCategorySummary(c *entities.Category) CategorySummary {
	return CategorySummary{
		ID:    c.ID().String(),
		Name:  c.Name().String(),
		Color: c.Color().String(),
		Icon:  c.Icon().String(),
	}
}
