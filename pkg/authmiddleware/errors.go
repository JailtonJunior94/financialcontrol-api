package authmiddleware

import "errors"

var (
	errMissingHeader   = errors.New("authorization header ausente")
	errMalformedHeader = errors.New("authorization header malformado")
)
