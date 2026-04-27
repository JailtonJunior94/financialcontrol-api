package authmiddleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/authmiddleware"
	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"
	jwtmocks "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt/mocks"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// ProtectedSuite tests the Protected middleware.
type ProtectedSuite struct {
	suite.Suite
	parser *jwtmocks.Parser
}

func TestProtectedSuite(t *testing.T) {
	suite.Run(t, new(ProtectedSuite))
}

func (s *ProtectedSuite) SetupTest() {
	s.parser = jwtmocks.NewParser(s.T())
}

func (s *ProtectedSuite) TestProtected() {
	type args struct {
		authHeader string
	}

	scenarios := []struct {
		name   string
		args   args
		setup  func()
		expect func(resp *http.Response)
	}{
		{
			name:  "missing authorization header returns 401",
			args:  args{authHeader: ""},
			setup: func() {},
			expect: func(resp *http.Response) {
				s.Equal(fiber.StatusUnauthorized, resp.StatusCode)
				s.Equal(`Bearer realm="financialcontrol", error="invalid_token"`, resp.Header.Get("WWW-Authenticate"))
			},
		},
		{
			name:  "malformed authorization header returns 401",
			args:  args{authHeader: "Basic dXNlcjpwYXNz"},
			setup: func() {},
			expect: func(resp *http.Response) {
				s.Equal(fiber.StatusUnauthorized, resp.StatusCode)
				s.Equal(`Bearer realm="financialcontrol", error="invalid_token"`, resp.Header.Get("WWW-Authenticate"))
			},
		},
		{
			name: "invalid token returns 401",
			args: args{authHeader: "Bearer invalidtoken"},
			setup: func() {
				s.parser.EXPECT().
					Parse(mock.Anything, "invalidtoken").
					Return(pkgjwt.Identity{}, pkgjwt.ErrInvalidToken).
					Once()
			},
			expect: func(resp *http.Response) {
				s.Equal(fiber.StatusUnauthorized, resp.StatusCode)
			},
		},
		{
			name: "expired token returns 401",
			args: args{authHeader: "Bearer expiredtoken"},
			setup: func() {
				s.parser.EXPECT().
					Parse(mock.Anything, "expiredtoken").
					Return(pkgjwt.Identity{}, pkgjwt.ErrExpiredToken).
					Once()
			},
			expect: func(resp *http.Response) {
				s.Equal(fiber.StatusUnauthorized, resp.StatusCode)
			},
		},
		{
			name: "valid token populates UserContext with identity and returns 200",
			args: args{authHeader: "Bearer validtoken"},
			setup: func() {
				s.parser.EXPECT().
					Parse(mock.Anything, "validtoken").
					Return(pkgjwt.Identity{UserID: "user-1", Email: "user@example.com"}, nil).
					Once()
			},
			expect: func(resp *http.Response) {
				s.Equal(fiber.StatusOK, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				s.Equal("user-1:user@example.com", string(body))
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			app := fiber.New()
			app.Get("/test", authmiddleware.Protected(s.parser), func(c *fiber.Ctx) error {
				id, err := identitycontext.FromContext(c.UserContext())
				if err != nil {
					return c.SendStatus(fiber.StatusInternalServerError)
				}
				return c.SendString(id.UserID + ":" + id.Email)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if sc.args.authHeader != "" {
				req.Header.Set("Authorization", sc.args.authHeader)
			}
			resp, err := app.Test(req)
			s.Require().NoError(err)
			sc.expect(resp)
		})
	}
}

// OptionalSuite tests the Optional middleware.
type OptionalSuite struct {
	suite.Suite
	parser *jwtmocks.Parser
}

func TestOptionalSuite(t *testing.T) {
	suite.Run(t, new(OptionalSuite))
}

func (s *OptionalSuite) SetupTest() {
	s.parser = jwtmocks.NewParser(s.T())
}

func (s *OptionalSuite) TestOptional() {
	type args struct {
		authHeader string
	}

	scenarios := []struct {
		name   string
		args   args
		setup  func()
		expect func(resp *http.Response)
	}{
		{
			name:  "missing authorization header passes without identity",
			args:  args{authHeader: ""},
			setup: func() {},
			expect: func(resp *http.Response) {
				s.Equal(fiber.StatusOK, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				s.Equal("no-identity", string(body))
			},
		},
		{
			name: "valid token populates UserContext and passes",
			args: args{authHeader: "Bearer validtoken"},
			setup: func() {
				s.parser.EXPECT().
					Parse(mock.Anything, "validtoken").
					Return(pkgjwt.Identity{UserID: "user-2", Email: "user2@example.com"}, nil).
					Once()
			},
			expect: func(resp *http.Response) {
				s.Equal(fiber.StatusOK, resp.StatusCode)
				body, _ := io.ReadAll(resp.Body)
				s.Equal("user-2:user2@example.com", string(body))
			},
		},
		{
			name:  "malformed authorization header returns 401",
			args:  args{authHeader: "Token abc"},
			setup: func() {},
			expect: func(resp *http.Response) {
				s.Equal(fiber.StatusUnauthorized, resp.StatusCode)
				s.Equal(`Bearer realm="financialcontrol", error="invalid_token"`, resp.Header.Get("WWW-Authenticate"))
			},
		},
		{
			name: "invalid token returns 401",
			args: args{authHeader: "Bearer invalidtoken"},
			setup: func() {
				s.parser.EXPECT().
					Parse(mock.Anything, "invalidtoken").
					Return(pkgjwt.Identity{}, pkgjwt.ErrInvalidToken).
					Once()
			},
			expect: func(resp *http.Response) {
				s.Equal(fiber.StatusUnauthorized, resp.StatusCode)
			},
		},
	}

	for _, sc := range scenarios {
		s.Run(sc.name, func() {
			sc.setup()
			app := fiber.New()
			app.Get("/test", authmiddleware.Optional(s.parser), func(c *fiber.Ctx) error {
				id, err := identitycontext.FromContext(c.UserContext())
				if err != nil {
					return c.SendString("no-identity")
				}
				return c.SendString(id.UserID + ":" + id.Email)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if sc.args.authHeader != "" {
				req.Header.Set("Authorization", sc.args.authHeader)
			}
			resp, err := app.Test(req)
			s.Require().NoError(err)
			sc.expect(resp)
		})
	}
}
