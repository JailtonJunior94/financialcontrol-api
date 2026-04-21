package application_test

import (
	"errors"
	"testing"

	appresponses "github.com/jailtonjunior94/financialcontrol-api/internal/application/dtos/responses"
	"github.com/jailtonjunior94/financialcontrol-api/internal/domain/entities"
	catalogapp "github.com/jailtonjunior94/financialcontrol-api/internal/modules/catalog/application"

	"github.com/stretchr/testify/require"
)

type flagRepositoryStub struct {
	flags []entities.Flag
	err   error
}

func (s *flagRepositoryStub) GetFlags() ([]entities.Flag, error) {
	return s.flags, s.err
}

type categoryRepositoryStub struct {
	categories []entities.Category
	err        error
}

func (s *categoryRepositoryStub) GetCategories() ([]entities.Category, error) {
	return s.categories, s.err
}

func TestFlagsReturnsMappedPayload(t *testing.T) {
	service := catalogapp.NewFlagService(&flagRepositoryStub{
		flags: []entities.Flag{
			{Entity: entities.Entity{ID: "flag-id", Active: true}, Name: "Visa"},
		},
	})

	response := service.Flags()

	require.Equal(t, appresponses.Ok(nil).StatusCode, response.StatusCode)

	payload, ok := response.Data.([]catalogapp.FlagResponse)
	require.True(t, ok)
	require.Len(t, payload, 1)
	require.Equal(t, "Visa", payload[0].Name)
}

func TestFlagsReturnsServerErrorWhenRepositoryFails(t *testing.T) {
	service := catalogapp.NewFlagService(&flagRepositoryStub{err: errors.New("db error")})

	response := service.Flags()

	require.Equal(t, appresponses.ServerError().StatusCode, response.StatusCode)
}

func TestCategoriesReturnsMappedPayload(t *testing.T) {
	service := catalogapp.NewCategoryService(&categoryRepositoryStub{
		categories: []entities.Category{
			{Entity: entities.Entity{ID: "category-id", Active: true}, Name: "Moradia", Sequence: 1},
		},
	})

	response := service.Categories()

	require.Equal(t, appresponses.Ok(nil).StatusCode, response.StatusCode)

	payload, ok := response.Data.([]catalogapp.CategoryResponse)
	require.True(t, ok)
	require.Len(t, payload, 1)
	require.Equal(t, "Moradia", payload[0].Name)
}

func TestCategoriesReturnsServerErrorWhenRepositoryFails(t *testing.T) {
	service := catalogapp.NewCategoryService(&categoryRepositoryStub{err: errors.New("db error")})

	response := service.Categories()

	require.Equal(t, appresponses.ServerError().StatusCode, response.StatusCode)
}
