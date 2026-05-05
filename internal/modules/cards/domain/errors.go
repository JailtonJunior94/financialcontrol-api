package domain

import "errors"

var (
	ErrCardNotFound        = errors.New("Cartão não encontrado")
	ErrFlagNotFound        = errors.New("Bandeira não encontrada")
	ErrInvalidCardID       = errors.New("ID de cartão inválido")
	ErrInvalidFlagID       = errors.New("ID de bandeira inválido")
	ErrInvalidCardName     = errors.New("Nome do cartão inválido")
	ErrInvalidCardNumber   = errors.New("Número do cartão inválido")
	ErrInvalidClosingDay   = errors.New("Melhor dia de compra inválido (1..31)")
	ErrInvalidDueDay       = errors.New("Dia de vencimento inválido (1..31)")
	ErrInvalidBillingCycle = errors.New("Ciclo de faturamento inválido")
)
