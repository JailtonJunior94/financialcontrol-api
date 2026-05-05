package vos

import sharedidentityvo "github.com/jailtonjunior94/financialcontrol-api/pkg/identityvo"

// UserID remains available in the legacy identity domain package as a
// compatibility shim, but the canonical implementation now lives in pkg/identityvo.
type UserID = sharedidentityvo.UserID

var ErrInvalidUserID = sharedidentityvo.ErrInvalidUserID

func NewUserID() UserID {
	return sharedidentityvo.NewUserID()
}

func ParseUserID(s string) (UserID, error) {
	return sharedidentityvo.ParseUserID(s)
}
