package interfaces

import (
	"context"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/entities"
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
)

type UserRepository interface {
	Add(ctx context.Context, u *entities.User) error
	GetByEmail(ctx context.Context, email vos.Email) (*entities.User, error)
	GetByID(ctx context.Context, id vos.UserID) (*entities.User, error)
}
