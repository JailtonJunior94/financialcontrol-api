package security

import "golang.org/x/crypto/bcrypt"

type HashAdapter interface {
	GenerateHash(str string) (string, error)
	CheckHash(hash, str string) bool
}

type BcryptHashAdapter struct{}

func NewHashAdapter() HashAdapter {
	return &BcryptHashAdapter{}
}

func (h *BcryptHashAdapter) GenerateHash(str string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(str), 5)
	return string(bytes), err
}

func (h *BcryptHashAdapter) CheckHash(hash, str string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(str))
	return err == nil
}
