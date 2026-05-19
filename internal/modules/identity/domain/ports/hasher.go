package ports

import "github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"

type Hasher interface {
	Hash(plain string) (vos.HashedPassword, error)
	Verify(hashed vos.HashedPassword, plain string) bool
}
