package domain

import "errors"

const InvalidUserOrPasswordMessage = "Usuário e/ou senha inválidos"

var (
	ErrInvalidEmail       = errors.New("e-mail inválido")
	ErrInvalidPassword    = errors.New("senha inválida")
	ErrUserNotFound       = errors.New("usuário não encontrado")
	ErrInvalidCredentials = errors.New("usuário e/ou senha inválidos")
	ErrUserAlreadyExists  = errors.New("usuário já cadastrado")
	ErrTokenIssuance      = errors.New("falha ao emitir token")
	ErrIdentityInvalid    = errors.New("identidade no contexto possui user ID inválido")
)
