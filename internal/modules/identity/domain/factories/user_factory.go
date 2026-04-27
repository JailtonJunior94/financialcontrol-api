package factories

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
)

// New builds a User entity from raw primitives, validating VOs in the process.
func New(name, email, hashedPassword string) (*entities.User, error) {
	em, err := vos.NewEmail(email)
	if err != nil {
		return nil, err
	}
	pwd, err := vos.NewHashedPassword(hashedPassword)
	if err != nil {
		return nil, err
	}
	return entities.NewUser(name, em, pwd)
}
