package domain

import "errors"

//nolint:staticcheck // Preserva mensagens públicas de domínio expostas pelo contrato HTTP.
var (
	ErrCategoryNotFound             = errors.New("categoria não encontrada")
	ErrInvalidCategoryID            = errors.New("ID de categoria inválido")
	ErrInvalidCategoryName          = errors.New("nome de categoria inválido")
	ErrInvalidCategoryColor         = errors.New("cor de categoria inválida")
	ErrInvalidCategoryIcon          = errors.New("ícone de categoria inválido")
	ErrCategoryHierarchyUnsupported = errors.New("subcategorias não são suportadas pelo schema atual")
	ErrParentNotFound               = errors.New("categoria pai não encontrada")
	ErrParentInactive               = errors.New("categoria pai inativa")
	ErrSubcategoryDepthExceeded     = errors.New("profundidade máxima de subcategoria excedida")
	ErrCategoryNameAlreadyExists    = errors.New("já existe categoria com este nome no escopo")
)
