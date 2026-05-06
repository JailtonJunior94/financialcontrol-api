package domain

import "errors"

//nolint:staticcheck // Preserva mensagens públicas legadas consumidas pelos chamadores.
var (
	TransactionNotFound     = errors.New("Não foi possível encontrar a Transação")
	TransactionItemNotFound = errors.New("Não foi possível encontrar o Item da Transação")
	TransactionExists       = errors.New("Já existe mês cadastrado para apontamento")
)
