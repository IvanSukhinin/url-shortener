package jwt

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type Claims struct {
	UUID  uuid.UUID `json:"uid"`
	Email string    `json:"email"`
	jwt.StandardClaims
}

// Parse parses claims and validates jwt token.
func Parse(tokenString string, secret []byte) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	return claims, err
}
