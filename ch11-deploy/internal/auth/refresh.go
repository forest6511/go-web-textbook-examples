package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// NewRefreshToken は 32 バイト乱数を生成し、生文字列とハッシュを返す
func NewRefreshToken() (raw string, hash []byte, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", nil, err
	}
	raw = base64.RawURLEncoding.EncodeToString(buf)
	sum := sha256.Sum256([]byte(raw))
	return raw, sum[:], nil
}

// HashRefreshToken は検証時に受け取った平文を同じ規則でハッシュ化する
func HashRefreshToken(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}
