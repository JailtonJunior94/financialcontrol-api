package identitycontext

import (
	"context"
)

type Identity struct {
	UserID string
	Email  string
}

type contextKey struct{}

func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

func FromContext(ctx context.Context) (Identity, error) {
	id, ok := ctx.Value(contextKey{}).(Identity)
	if !ok {
		return Identity{}, ErrNoIdentity
	}
	return id, nil
}
