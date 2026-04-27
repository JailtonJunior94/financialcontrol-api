package security_test

import (
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/security"
	"github.com/stretchr/testify/suite"
)

type HashAdapterSuite struct {
	suite.Suite
	sut security.HashAdapter
}

func TestHashAdapterSuite(t *testing.T) {
	suite.Run(t, new(HashAdapterSuite))
}

func (s *HashAdapterSuite) SetupTest() {
	s.sut = security.NewHashAdapter()
}

func (s *HashAdapterSuite) TestHash() {
	scenarios := []struct {
		name   string
		plain  string
		expect func(hash string, err error)
	}{
		{
			name:  "valid plain text produces non-empty hash",
			plain: "mysecretpassword",
			expect: func(hash string, err error) {
				s.NoError(err)
				s.NotEmpty(hash)
			},
		},
		{
			name:  "empty plain text produces a valid hash",
			plain: "",
			expect: func(hash string, err error) {
				s.NoError(err)
				s.NotEmpty(hash)
			},
		},
		{
			name:  "two calls with same input produce different hashes (bcrypt salt)",
			plain: "password123",
			expect: func(hash string, err error) {
				s.NoError(err)
				second, err2 := s.sut.Hash("password123")
				s.NoError(err2)
				s.NotEqual(hash, second)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			hash, err := s.sut.Hash(sc.plain)
			sc.expect(hash, err)
		})
	}
}

func (s *HashAdapterSuite) TestVerify() {
	plain := "correctpassword"
	hash, err := s.sut.Hash(plain)
	s.Require().NoError(err)

	knownHash, err := s.sut.Hash("knownpassword")
	s.Require().NoError(err)

	scenarios := []struct {
		name   string
		hashed string
		plain  string
		expect func(ok bool)
	}{
		{
			name:   "correct password verifies against its hash",
			hashed: hash,
			plain:  plain,
			expect: func(ok bool) {
				s.True(ok)
			},
		},
		{
			name:   "wrong password does not verify",
			hashed: hash,
			plain:  "wrongpassword",
			expect: func(ok bool) {
				s.False(ok)
			},
		},
		{
			name:   "known hash round-trip",
			hashed: knownHash,
			plain:  "knownpassword",
			expect: func(ok bool) {
				s.True(ok)
			},
		},
		{
			name:   "invalid hash string returns false",
			hashed: "notahash",
			plain:  plain,
			expect: func(ok bool) {
				s.False(ok)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			ok := s.sut.Verify(sc.hashed, sc.plain)
			sc.expect(ok)
		})
	}
}
