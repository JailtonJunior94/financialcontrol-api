package adapters

import uuid "github.com/satori/go.uuid"

type IUuidAdapter interface {
	GenerateUuid() string
}

type UuidAdapter struct {
}

func NewUuidAdapter() IUuidAdapter {
	return &UuidAdapter{}
}

func (u *UuidAdapter) GenerateUuid() string {
	return uuid.NewV4().String()
}
