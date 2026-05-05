package vos

import (
	"crypto/rand"
	"errors"
	"fmt"
	"regexp"
)

type CardID string

var (
	ErrInvalidCardID = errors.New("ID de cartão inválido")
	uuidRE           = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

func newUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func isValidUUID(s string) bool {
	return uuidRE.MatchString(s)
}

func NewCardID() CardID {
	return CardID(newUUID())
}

func ParseCardID(s string) (CardID, error) {
	if !isValidUUID(s) {
		return "", ErrInvalidCardID
	}
	return CardID(s), nil
}

func (c CardID) String() string { return string(c) }
