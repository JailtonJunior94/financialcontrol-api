package domain

import "errors"

var (
	BillNotFound     = errors.New("Não foi encontrado conta do mês")
	BillItemNotFound = errors.New("Não foi encontrado nenhum item")
	BillExists       = errors.New("Já existe mês cadastrado para despesas")
)
