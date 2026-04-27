package identitycontext_test

import (
	"context"
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
	"github.com/stretchr/testify/suite"
)

type otherContextKey struct{}

type ContextSuite struct {
	suite.Suite
	ctx context.Context
}

func TestContextSuite(t *testing.T) {
	suite.Run(t, new(ContextSuite))
}

func (s *ContextSuite) SetupTest() {
	s.ctx = context.Background()
}

func (s *ContextSuite) TestWithIdentityAndFromContext() {
	scenarios := []struct {
		name   string
		setup  func() context.Context
		expect func(id identitycontext.Identity, err error)
	}{
		{
			name: "identity stored and retrieved correctly",
			setup: func() context.Context {
				return identitycontext.WithIdentity(s.ctx, identitycontext.Identity{
					UserID: "user-123",
					Email:  "user@example.com",
				})
			},
			expect: func(id identitycontext.Identity, err error) {
				s.NoError(err)
				s.Equal("user-123", id.UserID)
				s.Equal("user@example.com", id.Email)
			},
		},
		{
			name: "empty identity stored and retrieved correctly",
			setup: func() context.Context {
				return identitycontext.WithIdentity(s.ctx, identitycontext.Identity{})
			},
			expect: func(id identitycontext.Identity, err error) {
				s.NoError(err)
				s.Equal("", id.UserID)
				s.Equal("", id.Email)
			},
		},
		{
			name: "identity overwritten by second WithIdentity",
			setup: func() context.Context {
				ctx := identitycontext.WithIdentity(s.ctx, identitycontext.Identity{UserID: "old"})
				return identitycontext.WithIdentity(ctx, identitycontext.Identity{UserID: "new", Email: "new@example.com"})
			},
			expect: func(id identitycontext.Identity, err error) {
				s.NoError(err)
				s.Equal("new", id.UserID)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			ctx := sc.setup()
			id, err := identitycontext.FromContext(ctx)
			sc.expect(id, err)
		})
	}
}

func (s *ContextSuite) TestFromContextAbsent() {
	scenarios := []struct {
		name   string
		ctx    context.Context
		expect func(id identitycontext.Identity, err error)
	}{
		{
			name: "plain context returns ErrNoIdentity",
			ctx:  s.ctx,
			expect: func(id identitycontext.Identity, err error) {
				s.ErrorIs(err, identitycontext.ErrNoIdentity)
				s.Equal(identitycontext.Identity{}, id)
			},
		},
		{
			name: "context with different value key returns ErrNoIdentity",
			ctx:  context.WithValue(s.ctx, otherContextKey{}, "value"),
			expect: func(id identitycontext.Identity, err error) {
				s.ErrorIs(err, identitycontext.ErrNoIdentity)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			id, err := identitycontext.FromContext(sc.ctx)
			sc.expect(id, err)
		})
	}
}
