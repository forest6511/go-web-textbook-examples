package auth

import "context"

// Principal は認証済みリクエストの主体
type Principal struct {
	UserID int64
	Role   string
}

type principalCtxKey struct{}

// WithPrincipal は context に Principal を埋める
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalCtxKey{}, p)
}

// PrincipalFromContext は context から Principal を取り出す
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalCtxKey{}).(Principal)
	return p, ok
}
