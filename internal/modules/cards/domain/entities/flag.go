package entities

import (
	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/cards/domain/vos"
)

type Flag struct {
	id     vos.FlagID
	name   string
	active bool
}

func NewFlag(id vos.FlagID, name string, active bool) Flag {
	return Flag{id: id, name: name, active: active}
}

func (f Flag) ID() vos.FlagID { return f.id }
func (f Flag) Name() string   { return f.name }
func (f Flag) Active() bool   { return f.active }
