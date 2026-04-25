package customerrors

import "errors"

//nolint:staticcheck // Shared sentinel errors for cross-module use.
var (
	InternalServerError = errors.New(InternalServerErrorMessage)
	InvalidToken        = errors.New(InvalidTokenMessage)
	TitleIsRequired     = errors.New(TitleIsRequiredMessage)
	ValueIsRequired     = errors.New(ValueIsRequiredMessage)
	TypeIsRequired      = errors.New(TypeIsRequiredMessage)
	DateIsRequired      = errors.New(DateIsRequiredMessage)
	NameIsRequired      = errors.New(NameIsRequiredMessage)
)
