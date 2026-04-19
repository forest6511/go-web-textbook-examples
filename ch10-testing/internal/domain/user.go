package domain

type User struct {
	ID           int64
	Email        string
	PasswordHash []byte
	Role         string
}
