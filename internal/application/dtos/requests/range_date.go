package requests

import (
	"errors"
	"time"
)

type RangeDateRequest struct {
	StartDate time.Time `query:"start"`
	EndDate   time.Time `query:"end"`
}

func (r *RangeDateRequest) IsValid() error {
	if r.StartDate.IsZero() {
		return errors.New("A Data de Inicio é obrigatória")
	}

	if r.EndDate.IsZero() {
		return errors.New("A Data Final é obrigatória")
	}

	return nil
}
