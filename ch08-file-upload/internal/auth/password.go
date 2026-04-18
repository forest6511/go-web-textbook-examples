package auth

import (
	"errors"
	"os"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

// HashPassword は平文パスワードを bcrypt でハッシュ化する
func HashPassword(password string) ([]byte, error) {
	cost := bcryptCost()
	return bcrypt.GenerateFromPassword([]byte(password), cost)
}

// VerifyPassword はハッシュと平文を比較する
func VerifyPassword(hash []byte, password string) error {
	err := bcrypt.CompareHashAndPassword(hash, []byte(password))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrInvalidCredentials
	}
	return err
}

func bcryptCost() int {
	if s := os.Getenv("BCRYPT_COST"); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			return n
		}
	}
	return bcrypt.DefaultCost
}

// dummyHash はユーザー不在時も bcrypt 比較時間を消費してタイミング差を消す
var dummyHash []byte

func init() {
	h, err := bcrypt.GenerateFromPassword(
		[]byte("dummy-password-for-timing"), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	dummyHash = h
}

// ConsumeDummy はユーザー不在時に呼び、時間差を消す
func ConsumeDummy(password string) {
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
}
