package services

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/interfaces"
)

type CredentialsChecker struct {
	hasher interfaces.Hasher
}

func NewCredentialsChecker(hasher interfaces.Hasher) *CredentialsChecker {
	return &CredentialsChecker{hasher: hasher}
}

func (c *CredentialsChecker) Check(user *entities.User, plainPassword string) bool {
	return c.hasher.Verify(user.Password(), plainPassword)
}
