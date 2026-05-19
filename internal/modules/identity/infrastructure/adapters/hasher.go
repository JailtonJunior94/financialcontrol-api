package adapters

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/ports"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
	platformsecurity "github.com/jailtonjunior94/financialcontrol-api/pkg/security"
)

type hasher struct {
	inner platformsecurity.HashAdapter
}

func NewHasher(inner platformsecurity.HashAdapter) ports.Hasher {
	return &hasher{inner: inner}
}

func (a *hasher) Hash(plain string) (vos.HashedPassword, error) {
	s, err := a.inner.Hash(plain)
	if err != nil {
		return "", err
	}
	return vos.NewHashedPassword(s)
}

func (a *hasher) Verify(hashed vos.HashedPassword, plain string) bool {
	return a.inner.Verify(hashed.String(), plain)
}
