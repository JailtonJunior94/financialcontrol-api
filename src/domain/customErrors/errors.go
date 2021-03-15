package customErrors

import "errors"

var (
	InternalServerError   = errors.New(InternalServerErrorMessage)
	InvalidToken          = errors.New(InvalidTokenMessage)
	EmailIsRequired       = errors.New(EmailIsRequiredMessage)
	PasswordIsRequired    = errors.New(PasswordIsRequiredMessage)
	InvalidUserOrPassword = errors.New(InvalidUserOrPasswordMessage)
)
