package authmiddleware

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/jailtonjunior94/financialcontrol-api/pkg/identitycontext"
	pkgjwt "github.com/jailtonjunior94/financialcontrol-api/pkg/jwt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const wwwAuthenticate = `Bearer realm="financialcontrol", error="invalid_token"`

// Protected blocks requests without a valid Bearer token, responding 401 per RFC 6750.
func Protected(parser pkgjwt.Parser) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := extractToken(c)
		if err != nil {
			warnAuthFailure(c, reasonFrom(err))
			return unauthorized(c, err.Error())
		}
		id, err := parser.Parse(c.UserContext(), token)
		if err != nil {
			warnAuthFailure(c, reasonFromJWT(err))
			return unauthorized(c, "token inválido")
		}
		c.SetUserContext(identitycontext.WithIdentity(c.UserContext(), identitycontext.Identity{
			UserID: id.UserID,
			Email:  id.Email,
		}))
		return c.Next()
	}
}

// Optional passes through when Authorization is absent; validates and populates context on success;
// responds 401 when the header is present but the token is malformed or invalid.
func Optional(parser pkgjwt.Parser) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token, err := extractToken(c)
		if err != nil {
			if errors.Is(err, errMissingHeader) {
				return c.Next()
			}
			warnAuthFailure(c, reasonFrom(err))
			return unauthorized(c, err.Error())
		}
		id, err := parser.Parse(c.UserContext(), token)
		if err != nil {
			warnAuthFailure(c, reasonFromJWT(err))
			return unauthorized(c, "token inválido")
		}
		c.SetUserContext(identitycontext.WithIdentity(c.UserContext(), identitycontext.Identity{
			UserID: id.UserID,
			Email:  id.Email,
		}))
		return c.Next()
	}
}

func extractToken(c *fiber.Ctx) (string, error) {
	header := c.Get("Authorization")
	if header == "" {
		return "", errMissingHeader
	}
	if !strings.HasPrefix(header, "Bearer ") {
		return "", errMalformedHeader
	}
	token := strings.TrimPrefix(header, "Bearer ")
	if token == "" {
		return "", errMalformedHeader
	}
	return token, nil
}

func unauthorized(c *fiber.Ctx, msg string) error {
	c.Set("WWW-Authenticate", wwwAuthenticate)
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": msg})
}

func warnAuthFailure(c *fiber.Ctx, reason string) {
	slog.Warn("auth failure",
		"route", c.Path(),
		"reason", reason,
		"request_id", requestID(c),
	)
}

func requestID(c *fiber.Ctx) string {
	if id := c.Get("X-Request-ID"); id != "" {
		return id
	}
	return uuid.New().String()
}

func reasonFrom(err error) string {
	if errors.Is(err, errMissingHeader) {
		return "missing_token"
	}
	return "malformed_header"
}

func reasonFromJWT(err error) string {
	if errors.Is(err, pkgjwt.ErrExpiredToken) {
		return "expired"
	}
	if errors.Is(err, pkgjwt.ErrAlgorithmMismatch) {
		return "alg_mismatch"
	}
	return "invalid_token"
}
