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
		//nolint:staticcheck // Legacy validation message is part of the current API contract.
		return errors.New("A Data de Inicio é obrigatória")
	}

	if r.EndDate.IsZero() {
		//nolint:staticcheck // Legacy validation message is part of the current API contract.
		return errors.New("A Data Final é obrigatória")
	}

	return nil
}
