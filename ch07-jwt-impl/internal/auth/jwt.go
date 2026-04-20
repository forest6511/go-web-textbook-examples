package auth

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AccessClaims は本書で使う Access Token のペイロード
type AccessClaims struct {
	UserID int64  `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type Config struct {
	HMACSecret   []byte
	Issuer       string
	Audience     string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
	Leeway       time.Duration
	SecureCookie bool
}

var ErrMissingSecret = errors.New("JWT_SECRET is empty or too short")

func LoadConfig() (*Config, error) {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return nil, ErrMissingSecret
	}
	return &Config{
		HMACSecret:   []byte(secret),
		Issuer:       envOr("JWT_ISSUER", "go-web-textbook"),
		Audience:     envOr("JWT_AUDIENCE", "go-web-textbook-api"),
		AccessTTL:    envDur("JWT_ACCESS_TTL", 15*time.Minute),
		RefreshTTL:   envDur("JWT_REFRESH_TTL", 7*24*time.Hour),
		Leeway:       envDur("JWT_LEEWAY", 30*time.Second),
		SecureCookie: os.Getenv("APP_ENV") == "production",
	}, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envDur(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

// NewAccessToken は signed JWT を返す
func NewAccessToken(
	userID int64, role string, cfg *Config,
) (string, error) {
	now := time.Now()
	claims := AccessClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			Subject:   strconv.FormatInt(userID, 10),
			Audience:  jwt.ClaimStrings{cfg.Audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.AccessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(cfg.HMACSecret)
}

// ParseAccessToken は JWT を検証して AccessClaims を返す
func ParseAccessToken(raw string, cfg *Config) (*AccessClaims, error) {
	claims := &AccessClaims{}
	_, err := jwt.ParseWithClaims(raw, claims,
		func(tok *jwt.Token) (any, error) {
			if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf(
					"unexpected signing method: %v",
					tok.Header["alg"])
			}
			return cfg.HMACSecret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithLeeway(cfg.Leeway),
		jwt.WithIssuer(cfg.Issuer),
		jwt.WithAudience(cfg.Audience),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
	return claims, nil
}
