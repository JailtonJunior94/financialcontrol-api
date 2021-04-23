package interfaces

import "github.com/jailtonjunior94/financialcontrol-api/src/domain/entities"

type IFlagRepository interface {
	GetFlags() (flags []entities.Flag, err error)
}
