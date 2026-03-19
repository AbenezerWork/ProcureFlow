package identity

import (
	"time"

	"github.com/google/uuid"
)

type Claims struct {
	UserID    uuid.UUID
	ExpiresAt time.Time
}

type Token struct {
	AccessToken string
	ExpiresAt   time.Time
}

type Session struct {
	User  User
	Token Token
}
