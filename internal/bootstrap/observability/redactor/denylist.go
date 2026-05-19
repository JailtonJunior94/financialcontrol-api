package redactor

// DefaultDenylist is the project-wide PII denylist applied to logs, spans and
// db.statement sanitisation. Extend via PR; quarterly audit required (RF-09.4).
var DefaultDenylist = MustCompile(
	// PCI / credentials
	"pan", "card_number", "cardnumber", "cvv", "cvc",
	"password", "passwd", "secret", "token",
	"access_token", "refresh_token", "authorization",
	// Brazilian documents
	"cpf", "cnpj", "rg",
	// Contact
	"email", "phone", "telefone", "address", "endereco", "zipcode", "cep",
	// Banking and identity
	"account_number", "agency", "bank_account",
	"pix_key", "pix_chave", "full_name", "holder_name", "nome_completo",
)
