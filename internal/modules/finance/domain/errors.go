package domain

import "errors"

var (
	ErrTransactionNotFound              = errors.New("transação não encontrada")
	ErrInvoiceNotFound                  = errors.New("fatura não encontrada")
	ErrInstallmentNotFound              = errors.New("parcela não encontrada")
	ErrCardNotFound                     = errors.New("cartão não encontrado")
	ErrCategoryNotFound                 = errors.New("categoria não encontrada")
	ErrSubcategoryNotFound              = errors.New("subcategoria não encontrada")
	ErrInvalidTransactionType           = errors.New("tipo de transação inválido")
	ErrInvalidPaymentMethod             = errors.New("meio de pagamento inválido")
	ErrInvalidAmount                    = errors.New("valor deve ser positivo")
	ErrDescriptionRequired              = errors.New("descrição é obrigatória (1..255 chars)")
	ErrOccurredAtTooFarInFuture         = errors.New("data não pode ser superior a hoje + 365 dias")
	ErrCardRequiredForPaymentMethod     = errors.New("cartão obrigatório para débito/crédito")
	ErrCardNotActive                    = errors.New("cartão inativo")
	ErrCategoryNotActive                = errors.New("categoria inativa")
	ErrSubcategoryNotActive             = errors.New("subcategoria inativa")
	ErrSubcategoryNotChildOfCategory    = errors.New("subcategoria não é filha direta da categoria")
	ErrInstallmentCountOutOfRange       = errors.New("número de parcelas inválido (1..24)")
	ErrInstallmentInClosedOrPaidInvoice = errors.New("parcela em fatura fechada ou paga")
	ErrInvoiceAlreadyPaid               = errors.New("fatura já está paga")
	ErrInvoiceCannotPay                 = errors.New("estado da fatura não permite pagamento")
	ErrRefundAlreadyExists              = errors.New("já existe estorno ativo para esta transação")
	ErrRefundOfRefundNotAllowed         = errors.New("não é permitido estornar um estorno")
	ErrTransactionHasDependentRefund    = errors.New("transação possui estorno dependente")
	ErrIdempotencyMismatch              = errors.New("idempotency key reutilizada com payload distinto")
	ErrUnsupportedCurrency              = errors.New("moeda não suportada")
	ErrSplitterRequired                 = errors.New("splitter e assigner são obrigatórios para installment_purchase")

	// Sentinelas de validação de fronteira (HTTP 400) — decisão G3.a
	ErrInvalidDateRange            = errors.New("intervalo de datas inválido: 'from' deve ser <= 'to'")
	ErrInvalidPaginationLimits     = errors.New("paginação inválida: page >= 1; page_size entre 1 e 100")
	ErrInvalidISODate              = errors.New("data inválida: usar formato ISO 8601 com timezone")
	ErrSubcategoryEqualsCategory   = errors.New("subcategoria não pode ser igual à categoria")
	ErrInvalidIdempotencyKeyFormat = errors.New("idempotency key deve ter 1..64 caracteres após trim")
	ErrInvalidYearMonth            = errors.New("year/month inválidos: year com 4 dígitos, month entre 1 e 12")
	ErrInvalidMoneyFormat          = errors.New("valor monetário deve ser string decimal (ex.: \"123.45\")") // H3.a
)
