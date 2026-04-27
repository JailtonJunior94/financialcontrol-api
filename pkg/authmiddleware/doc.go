// Package authmiddleware provides Fiber middleware for JWT-based authentication.
// It uses pkg/jwt for token validation and pkg/identitycontext for context propagation.
// Two middlewares are provided: Protected (blocks 401 on failure) and Optional
// (passes through when Authorization header is absent, 401 on malformed/invalid token).
package authmiddleware
