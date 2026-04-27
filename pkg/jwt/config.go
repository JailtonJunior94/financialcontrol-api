package jwt

import "fmt"

const minSecretLen = 32

func setDefaults(cfg Config) Config {
	if cfg.SigningAlgorithm == "" {
		cfg.SigningAlgorithm = "HS256"
	}
	if len(cfg.AllowedAlgorithms) == 0 {
		cfg.AllowedAlgorithms = []string{"HS256"}
	}
	return cfg
}

func validateConfig(cfg Config) error {
	if len(cfg.Secret) < minSecretLen {
		return ErrSecretTooShort
	}
	for _, alg := range cfg.AllowedAlgorithms {
		if cfg.SigningAlgorithm == alg {
			return nil
		}
	}
	return fmt.Errorf("signing algorithm %q is not in AllowedAlgorithms", cfg.SigningAlgorithm)
}
