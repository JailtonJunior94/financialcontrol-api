package domain

import "errors"

//nolint:staticcheck // Preserva mensagens públicas de domínio expostas pelo contrato HTTP.
var (
	ErrCategoryNotFound          = errors.New("Categoria não encontrada")
	ErrInvalidCategoryID         = errors.New("ID de categoria inválido")
	ErrInvalidCategoryName       = errors.New("Nome de categoria inválido")
	ErrInvalidCategoryColor      = errors.New("Cor de categoria inválida")
	ErrInvalidCategoryIcon       = errors.New("Ícone de categoria inválido")
	ErrParentNotFound            = errors.New("Categoria pai não encontrada")
	ErrParentInactive            = errors.New("Categoria pai inativa")
	ErrSubcategoryDepthExceeded  = errors.New("Profundidade máxima de subcategoria excedida")
	ErrCategoryNameAlreadyExists = errors.New("Já existe categoria com este nome no escopo")
)
