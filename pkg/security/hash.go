package security

import "golang.org/x/crypto/bcrypt"

type HashAdapter interface {
	Hash(plain string) (string, error)
	Verify(hashed, plain string) bool
}

type BcryptHashAdapter struct{}

func NewHashAdapter() HashAdapter {
	return &BcryptHashAdapter{}
}

func (h *BcryptHashAdapter) Hash(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), 5)
	return string(bytes), err
}

func (h *BcryptHashAdapter) Verify(hashed, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	return err == nil
}
