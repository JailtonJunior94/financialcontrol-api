package domain

import "errors"

//nolint:staticcheck // Preserva mensagens públicas de domínio expostas pelo contrato HTTP.
var (
	ErrCardNotFound        = errors.New("cartão não encontrado")
	ErrFlagNotFound        = errors.New("bandeira não encontrada")
	ErrInvalidCardID       = errors.New("ID de cartão inválido")
	ErrInvalidFlagID       = errors.New("ID de bandeira inválido")
	ErrInvalidCardName     = errors.New("nome do cartão inválido")
	ErrInvalidCardNumber   = errors.New("número do cartão inválido")
	ErrInvalidClosingDay   = errors.New("melhor dia de compra inválido (1..31)")
	ErrInvalidDueDay       = errors.New("dia de vencimento inválido (1..31)")
	ErrInvalidBillingCycle = errors.New("ciclo de faturamento inválido")
)
