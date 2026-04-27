package vos_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/internal/modules/identity/domain/vos"
	"github.com/stretchr/testify/suite"
)

type EmailSuite struct{ suite.Suite }

func TestEmailSuite(t *testing.T) { suite.Run(t, new(EmailSuite)) }

func (s *EmailSuite) TestNewEmail() {
	scenarios := []struct {
		name   string
		input  string
		expect func(e vos.Email, err error)
	}{
		{
			name:  "valid email",
			input: "user@example.com",
			expect: func(e vos.Email, err error) {
				s.NoError(err)
				s.Equal(vos.Email("user@example.com"), e)
			},
		},
		{
			name:  "empty email",
			input: "",
			expect: func(e vos.Email, err error) {
				s.Error(err)
				s.ErrorIs(err, vos.ErrInvalidEmail)
			},
		},
		{
			name:  "missing @",
			input: "notanemail",
			expect: func(e vos.Email, err error) {
				s.Error(err)
				s.ErrorIs(err, vos.ErrInvalidEmail)
			},
		},
		{
			name:  "missing domain",
			input: "user@",
			expect: func(e vos.Email, err error) {
				s.Error(err)
				s.ErrorIs(err, vos.ErrInvalidEmail)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			e, err := vos.NewEmail(sc.input)
			sc.expect(e, err)
		})
	}
}

func (s *EmailSuite) TestString() {
	e, _ := vos.NewEmail("user@example.com")
	s.Equal("user@example.com", e.String())
}
