package jwt

import "errors"

var (
	ErrInvalidToken      = errors.New("token inválido")
	ErrExpiredToken      = errors.New("token expirado")
	ErrAlgorithmMismatch = errors.New("algoritmo de token não permitido")
	ErrSecretTooShort    = errors.New("segredo JWT deve ter no mínimo 32 bytes")
)
