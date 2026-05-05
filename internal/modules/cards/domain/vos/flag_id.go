package vos

import "errors"

type FlagID string

var ErrInvalidFlagID = errors.New("ID de bandeira inválido")

func ParseFlagID(s string) (FlagID, error) {
	if !isValidUUID(s) {
		return "", ErrInvalidFlagID
	}
	return FlagID(s), nil
}

func (f FlagID) String() string { return string(f) }
