package customErrors

const (
	InternalServerErrorMessage     = "Ocorreu um erro inesperado"
	InvalidTokenMessage            = "Token inválido ou expirado"
	UnprocessableEntityMessage     = "Unprocessable Entity"
	ErrorCreateUserMessage         = "Não foi possível cadastrar usuário"
	EmailIsRequiredMessage         = "O E-mail é obrigatório"
	PasswordIsRequiredMessage      = "A Senha é obrigatória"
	InvalidUserOrPasswordMessage   = "Usuário e/ou senha inválidos"
	MissingJWTMessage              = "Missing or malformed JWT"
	JwtErrorMessage                = "JWT ausente ou malformado"
	TitleIsRequiredMessage         = "O Título é obrigatório"
	ValueIsRequiredMessage         = "O Valor é obrigatório"
	TypeIsRequiredMessage          = "O Tipo é obrigatório"
	DateIsRequiredMessage          = "A Data é obrigatória"
	NameIsRequiredMessage          = "O Nome é obrigatório"
	TransactionNotFoundMessage     = "Não foi possível encontrar a Transação"
	TransactionItemNotFoundMessage = "Não foi possível encontrar o Item da Transação"
)
