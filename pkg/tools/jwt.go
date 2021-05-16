package tools

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	SecretKey = "ZWRnZXgtZ2F0ZXdheSBqd3QgYXV0aG9yOmx1b21pbnNoZW5n"
)

type LoginClaims struct {
	UserID int64      `json:"user_id,omitempty"`
	Claims jwt.Claims `json:"claims,omitempty"`
}

func (c LoginClaims) Valid() error {
	if c.UserID == 0 {
		return fmt.Errorf("user_id is invalid: user_id = %v", c.UserID)
	}
	return c.Claims.Valid()
}

func GenerateToken(userID int64, expireDuration time.Duration) (string, error) {
	expire := time.Now().Add(expireDuration)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, LoginClaims{
		UserID: userID,
		Claims: jwt.StandardClaims{
			ExpiresAt: expire.Unix(),
		},
	})
	return token.SignedString([]byte(SecretKey))
}
