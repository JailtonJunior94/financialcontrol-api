package customErrors

import "errors"

var (
	InternalServerError = errors.New(InternalServerErrorMessage)
	InvalidToken        = errors.New(InvalidTokenMessage)
)
