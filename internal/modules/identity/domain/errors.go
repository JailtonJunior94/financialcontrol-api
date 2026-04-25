package domain

import "errors"

const (
	EmailIsRequiredMessage       = "O E-mail é obrigatório"
	PasswordIsRequiredMessage    = "A Senha é obrigatória"
	InvalidUserOrPasswordMessage = "Usuário e/ou senha inválidos"
	ErrorCreateUserMessage       = "Não foi possível cadastrar usuário"
)

var (
	EmailIsRequired       = errors.New(EmailIsRequiredMessage)
	PasswordIsRequired    = errors.New(PasswordIsRequiredMessage)
	InvalidUserOrPassword = errors.New(InvalidUserOrPasswordMessage)
	ErrorCreateUser       = errors.New(ErrorCreateUserMessage)
)
