package auth

import (
	"testing"
	"time"
)

// BenchmarkIssueAccessToken は Access Token 発行のホットパスを計測する。
// 実行: go test -bench=. -benchmem ./internal/auth
func BenchmarkIssueAccessToken(b *testing.B) {
	cfg := &Config{
		HMACSecret:   []byte("0123456789abcdef0123456789abcdef"),
		Issuer:       "bench",
		Audience:     "bench",
		AccessTTL:    15 * time.Minute,
		RefreshTTL:   7 * 24 * time.Hour,
		Leeway:       30 * time.Second,
		SecureCookie: false,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = NewAccessToken(1, string(RoleUser), cfg)
	}
}
