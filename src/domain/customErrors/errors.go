package customErrors

import "errors"

var (
	InternalServerError     = errors.New(InternalServerErrorMessage)
	InvalidToken            = errors.New(InvalidTokenMessage)
	EmailIsRequired         = errors.New(EmailIsRequiredMessage)
	PasswordIsRequired      = errors.New(PasswordIsRequiredMessage)
	InvalidUserOrPassword   = errors.New(InvalidUserOrPasswordMessage)
	TitleIsRequired         = errors.New(TitleIsRequiredMessage)
	ValueIsRequired         = errors.New(ValueIsRequiredMessage)
	TypeIsRequired          = errors.New(TypeIsRequiredMessage)
	DateIsRequired          = errors.New(DateIsRequiredMessage)
	NameIsRequired          = errors.New(NameIsRequiredMessage)
	TransactionNotFound     = errors.New(TransactionNotFoundMessage)
	TransactionItemNotFound = errors.New(TransactionItemNotFoundMessage)
)
