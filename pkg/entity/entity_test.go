package entity_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/entity"
	"github.com/stretchr/testify/assert"
)

func TestEntity_NewEntity(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "generates non-empty ID and valid CreatedAt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e entity.Entity
			e.NewEntity()

			assert.NotEmpty(t, e.ID)
			assert.False(t, e.CreatedAt.IsZero())
			assert.True(t, e.Active)
		})
	}
}

func TestEntity_ChangeUpdatedAt(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "updates UpdatedAt to a non-zero time"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e entity.Entity
			e.NewEntity()
			e.ChangeUpdatedAt()

			assert.False(t, e.UpdatedAt.Time.IsZero())
		})
	}
}

func TestEntity_ChangeStatus(t *testing.T) {
	tests := []struct {
		name           string
		initialActive  bool
		newStatus      bool
		expectedActive bool
	}{
		{
			name:           "sets Active to false",
			initialActive:  true,
			newStatus:      false,
			expectedActive: false,
		},
		{
			name:           "sets Active to true",
			initialActive:  false,
			newStatus:      true,
			expectedActive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var e entity.Entity
			e.Active = tt.initialActive
			e.ChangeStatus(tt.newStatus)

			assert.Equal(t, tt.expectedActive, e.Active)
		})
	}
}
