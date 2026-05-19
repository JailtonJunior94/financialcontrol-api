package config

import (
	"errors"
	"fmt"
	"os"
)

var (
	ErrServiceNameRequired    = errors.New("SERVICE_NAME is required")
	ErrServiceVersionRequired = errors.New("SERVICE_VERSION is required")
	ErrInvalidEnvironment     = errors.New("ENVIRONMENT must be one of: development, staging, production")
)

var validEnvironments = map[string]struct{}{
	"development": {},
	"staging":     {},
	"production":  {},
}

// ServiceIdentity holds validated service identity read from environment variables.
type ServiceIdentity struct {
	name        string
	version     string
	environment string
}

// NewServiceIdentity reads SERVICE_NAME, SERVICE_VERSION and ENVIRONMENT and validates them.
// Fails with ErrServiceNameRequired, ErrServiceVersionRequired or ErrInvalidEnvironment if invalid.
func NewServiceIdentity() (ServiceIdentity, error) {
	name := os.Getenv("SERVICE_NAME")
	if name == "" {
		return ServiceIdentity{}, fmt.Errorf("bootstrap config: %w", ErrServiceNameRequired)
	}

	version := os.Getenv("SERVICE_VERSION")
	if version == "" {
		return ServiceIdentity{}, fmt.Errorf("bootstrap config: %w", ErrServiceVersionRequired)
	}

	environment := os.Getenv("ENVIRONMENT")
	if _, ok := validEnvironments[environment]; !ok {
		return ServiceIdentity{}, fmt.Errorf("bootstrap config: %w", ErrInvalidEnvironment)
	}

	return ServiceIdentity{name: name, version: version, environment: environment}, nil
}

func (s ServiceIdentity) Name() string        { return s.name }
func (s ServiceIdentity) Version() string     { return s.version }
func (s ServiceIdentity) Environment() string { return s.environment }
