package citokens

import "context"

type contextKey string

const tokenContextKey contextKey = "ciToken"

// WithToken attaches a CI token to the context.
func WithToken(ctx context.Context, token Token) context.Context {
	return context.WithValue(ctx, tokenContextKey, token)
}

// TokenFromContext extracts a CI token from context.
func TokenFromContext(ctx context.Context) (Token, bool) {
	token, ok := ctx.Value(tokenContextKey).(Token)
	return token, ok
}
